package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

// BatchEvent is the interface for all batch processing events.
type BatchEvent interface {
	batchEvent()
}

// JobStartedEvent is sent when a job begins encoding.
type JobStartedEvent struct {
	JobID     string
	InputPath string
}

func (JobStartedEvent) batchEvent() {}

// JobProgressEvent carries a progress update for a specific job.
type JobProgressEvent struct {
	JobID    string
	Progress float64 // 0-100
	Update   ProgressUpdate
}

func (JobProgressEvent) batchEvent() {}

// JobCompleteEvent is sent when a job finishes successfully.
type JobCompleteEvent struct {
	JobID      string
	InputPath  string
	OutputPath string
	InputSize  int64
	OutputSize int64
}

func (JobCompleteEvent) batchEvent() {}

// JobFailedEvent is sent when a job fails.
type JobFailedEvent struct {
	JobID     string
	InputPath string
	Error     string
	Attempt   int
}

func (JobFailedEvent) batchEvent() {}

// JobSkippedEvent is sent when a job is skipped.
type JobSkippedEvent struct {
	JobID     string
	InputPath string
	Reason    string
}

func (JobSkippedEvent) batchEvent() {}

// BatchCompleteEvent is sent when all jobs have been processed.
type BatchCompleteEvent struct {
	Stats QueueStats
}

func (BatchCompleteEvent) batchEvent() {}

// BatchOptions configures the batch encoding run.
type BatchOptions struct {
	Jobs        int           // number of parallel workers (default 1)
	Preset      presets.Preset
	HWEncoder   string
	OutputOpts  OutputOptions
	SkipOpts    SkipOptions
	MaxRetries  int // max attempts per job (default 2)
	OnSave      func() // called when queue state changes (for persistence)
}

// RunBatch starts batch encoding with a worker pool. It returns a channel
// of BatchEvent values. The channel is closed when all jobs are complete.
func RunBatch(ctx context.Context, queue *JobQueue, opts BatchOptions) <-chan BatchEvent {
	eventCh := make(chan BatchEvent, 16)

	numWorkers := opts.Jobs
	if numWorkers < 1 {
		numWorkers = 1
	}

	maxRetries := opts.MaxRetries
	if maxRetries < 1 {
		maxRetries = 2
	}

	if numWorkers > 1 {
		slog.Warn("running with multiple workers — file processing order is not guaranteed",
			"workers", numWorkers)
	}

	go func() {
		defer close(eventCh)

		var wg sync.WaitGroup
		jobCh := make(chan Job, numWorkers)

		// Launch workers
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for job := range jobCh {
					if ctx.Err() != nil {
						return
					}
					processJob(ctx, queue, job, opts, maxRetries, eventCh)
				}
			}(w)
		}

		// Feed jobs to workers
		func() {
			defer close(jobCh)
			for {
				if ctx.Err() != nil {
					return
				}
				job, ok := queue.Next()
				if !ok {
					return
				}
				select {
				case jobCh <- *job:
				case <-ctx.Done():
					// Put the job back to pending since we couldn't process it
					queue.Update(job.ID, func(j *Job) {
						j.Status = JobStatusPending
					})
					return
				}
			}
		}()

		wg.Wait()

		// Send batch complete event
		stats := queue.Stats()
		sendEvent(ctx, eventCh, BatchCompleteEvent{Stats: stats})
	}()

	return eventCh
}

func processJob(ctx context.Context, queue *JobQueue, job Job, opts BatchOptions, maxRetries int, eventCh chan<- BatchEvent) {
	// Check if should skip
	if skip, reason := ShouldSkip(job, opts.SkipOpts); skip {
		queue.Update(job.ID, func(j *Job) {
			j.Status = JobStatusSkipped
			j.CompletedAt = time.Now()
			j.Error = reason
		})
		if opts.OnSave != nil {
			opts.OnSave()
		}
		sendEvent(ctx, eventCh, JobSkippedEvent{
			JobID:     job.ID,
			InputPath: job.InputPath,
			Reason:    reason,
		})
		return
	}

	// Check skip optimal if we have enough info
	if opts.SkipOpts.SkipOptimal {
		info, err := Probe(ctx, job.InputPath)
		if err == nil {
			if skip, reason := ShouldSkipOptimal(info, opts.Preset.Codec, opts.Preset.CRF); skip {
				queue.Update(job.ID, func(j *Job) {
					j.Status = JobStatusSkipped
					j.CompletedAt = time.Now()
					j.Error = reason
				})
				if opts.OnSave != nil {
					opts.OnSave()
				}
				sendEvent(ctx, eventCh, JobSkippedEvent{
					JobID:     job.ID,
					InputPath: job.InputPath,
					Reason:    reason,
				})
				return
			}
		}
	}

	// Resolve output path
	outputPath := job.OutputPath
	if outputPath == "" {
		var err error
		outputPath, err = ResolveOutput(job.InputPath, opts.OutputOpts)
		if err != nil {
			queue.Update(job.ID, func(j *Job) {
				j.Status = JobStatusFailed
				j.Error = fmt.Sprintf("resolve output: %v", err)
				j.CompletedAt = time.Now()
			})
			if opts.OnSave != nil {
				opts.OnSave()
			}
			sendEvent(ctx, eventCh, JobFailedEvent{
				JobID:     job.ID,
				InputPath: job.InputPath,
				Error:     fmt.Sprintf("resolve output: %v", err),
				Attempt:   job.Attempts,
			})
			return
		}
		queue.Update(job.ID, func(j *Job) {
			j.OutputPath = outputPath
		})
	}

	sendEvent(ctx, eventCh, JobStartedEvent{
		JobID:     job.ID,
		InputPath: job.InputPath,
	})

	// Determine if in-place mode
	isInPlace := opts.OutputOpts.Mode == OutputModeInplace

	// Build encode options
	tempPath := TempPath(outputPath)
	encOpts := EncodeOptions{
		Input:     job.InputPath,
		Output:    tempPath,
		Preset:    opts.Preset,
		HWEncoder: opts.HWEncoder,
	}

	// Get source info for target-size calculation
	if opts.Preset.TargetSizeMB > 0 || isInPlace {
		info, err := Probe(ctx, job.InputPath)
		if err == nil {
			encOpts.SourceInfo = info
			if opts.Preset.TargetSizeMB > 0 {
				dur := time.Duration(info.Duration * float64(time.Second))
				audioBitrate := int64(128000)
				if info.AudioBitrate > 0 {
					audioBitrate = info.AudioBitrate
				}
				targetBytes := int64(opts.Preset.TargetSizeMB * 1024 * 1024)
				encOpts.VideoBitrate = CalculateBitrate(targetBytes, dur, audioBitrate)

				adaptW, adaptH := AdaptiveResolution(
					targetBytes, dur,
					info.Width, info.Height,
					info.Framerate,
				)
				if adaptW < info.Width || adaptH < info.Height {
					encOpts.ResolutionOverride = fmt.Sprintf("%dx%d", adaptW, adaptH)
				}
			}
		}
	}

	// Encode
	isTwoPass := ShouldUseTwoPass(encOpts)
	var progressCh <-chan ProgressUpdate
	var err error

	if isTwoPass {
		progressCh, err = EncodeTwoPass(ctx, encOpts)
	} else {
		progressCh, err = Encode(ctx, encOpts)
	}

	if err != nil {
		handleJobFailure(ctx, queue, job, opts, maxRetries, eventCh, err)
		return
	}

	// Read progress updates
	var lastErr error
	for update := range progressCh {
		if update.Error != nil {
			lastErr = update.Error
			break
		}
		queue.Update(job.ID, func(j *Job) {
			j.Progress = update.Percent
			j.Pass = update.Pass
		})
		sendEvent(ctx, eventCh, JobProgressEvent{
			JobID:    job.ID,
			Progress: update.Percent,
			Update:   update,
		})
	}

	if lastErr != nil {
		// Clean up temp file
		os.Remove(tempPath)
		handleJobFailure(ctx, queue, job, opts, maxRetries, eventCh, lastErr)
		return
	}

	// For in-place mode, verify the temp file before replacing
	if isInPlace {
		if err := verifyInPlace(ctx, job.InputPath, tempPath); err != nil {
			os.Remove(tempPath)
			handleJobFailure(ctx, queue, job, opts, maxRetries, eventCh, err)
			return
		}
	}

	// Move temp to final output
	if err := os.Rename(tempPath, outputPath); err != nil {
		os.Remove(tempPath)
		handleJobFailure(ctx, queue, job, opts, maxRetries, eventCh,
			fmt.Errorf("move output: %w", err))
		return
	}

	// Get output size
	var outputSize int64
	if stat, err := os.Stat(outputPath); err == nil {
		outputSize = stat.Size()
	}

	queue.Update(job.ID, func(j *Job) {
		j.Status = JobStatusComplete
		j.Progress = 100
		j.OutputSize = outputSize
		j.CompletedAt = time.Now()
	})
	if opts.OnSave != nil {
		opts.OnSave()
	}

	sendEvent(ctx, eventCh, JobCompleteEvent{
		JobID:      job.ID,
		InputPath:  job.InputPath,
		OutputPath: outputPath,
		InputSize:  job.InputSize,
		OutputSize: outputSize,
	})
}

func handleJobFailure(ctx context.Context, queue *JobQueue, job Job, opts BatchOptions, maxRetries int, eventCh chan<- BatchEvent, err error) {
	errMsg := err.Error()
	attempt := job.Attempts + 1

	queue.Update(job.ID, func(j *Job) {
		j.Attempts = attempt
		if attempt < maxRetries {
			// Re-queue for retry
			j.Status = JobStatusPending
			j.Progress = 0
			j.Error = fmt.Sprintf("attempt %d failed: %s", attempt, errMsg)
			slog.Info("re-queuing failed job for retry",
				"job_id", job.ID,
				"attempt", attempt,
				"max_retries", maxRetries,
				"error", errMsg,
			)
		} else {
			// Permanently failed
			j.Status = JobStatusFailed
			j.Error = errMsg
			j.CompletedAt = time.Now()
		}
	})
	if opts.OnSave != nil {
		opts.OnSave()
	}

	sendEvent(ctx, eventCh, JobFailedEvent{
		JobID:     job.ID,
		InputPath: job.InputPath,
		Error:     errMsg,
		Attempt:   attempt,
	})
}

// verifyInPlace probes the temp file and compares it to the source to ensure validity.
func verifyInPlace(ctx context.Context, sourcePath, tempPath string) error {
	sourceInfo, err := Probe(ctx, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to probe source for verification: %w", err)
	}

	tempInfo, err := Probe(ctx, tempPath)
	if err != nil {
		return fmt.Errorf("verification failed: output is not a valid video: %w", err)
	}

	// Check that temp has a video stream (width > 0)
	if tempInfo.Width <= 0 || tempInfo.Height <= 0 {
		return fmt.Errorf("verification failed: output has no video stream")
	}

	// Check duration is within 5% of source
	if sourceInfo.Duration > 0 {
		diff := tempInfo.Duration - sourceInfo.Duration
		if diff < 0 {
			diff = -diff
		}
		tolerance := sourceInfo.Duration * 0.05
		if diff > tolerance {
			return fmt.Errorf("verification failed: output duration %.1fs differs from source %.1fs by more than 5%%",
				tempInfo.Duration, sourceInfo.Duration)
		}
	}

	return nil
}

func sendEvent(ctx context.Context, ch chan<- BatchEvent, event BatchEvent) {
	select {
	case ch <- event:
	case <-ctx.Done():
	}
}
