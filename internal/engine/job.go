package engine

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// JobStatus represents the current state of a batch job.
type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusEncoding JobStatus = "encoding"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
	JobStatusSkipped  JobStatus = "skipped"
)

// Job represents a single file in the batch queue.
type Job struct {
	ID          string    `json:"id"`
	InputPath   string    `json:"inputPath"`
	OutputPath  string    `json:"outputPath"`
	PresetKey   string    `json:"presetKey"`
	Status      JobStatus `json:"status"`
	Progress    float64   `json:"progress"`    // 0-100
	InputSize   int64     `json:"inputSize"`
	OutputSize  int64     `json:"outputSize"`
	Error       string    `json:"error"`
	Attempts    int       `json:"attempts"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
	Pass        int       `json:"pass"` // 0 for single-pass, 1 or 2 for two-pass tracking

	// Duration is the source video duration in seconds (for sorting/estimation).
	Duration float64 `json:"duration"`
}

// QueueStats holds aggregate statistics for the job queue.
type QueueStats struct {
	Total           int   `json:"total"`
	Pending         int   `json:"pending"`
	Encoding        int   `json:"encoding"`
	Complete        int   `json:"complete"`
	Failed          int   `json:"failed"`
	Skipped         int   `json:"skipped"`
	TotalInputSize  int64 `json:"totalInputSize"`
	TotalOutputSize int64 `json:"totalOutputSize"`
}

// JobQueue is a thread-safe queue of encoding jobs.
type JobQueue struct {
	Jobs  []Job `json:"jobs"`
	mutex sync.RWMutex
}

// NewJobQueue creates a new empty job queue.
func NewJobQueue() *JobQueue {
	return &JobQueue{
		Jobs: make([]Job, 0),
	}
}

// generateID creates a simple unique ID for a job.
func generateID() string {
	return fmt.Sprintf("%d-%08x", time.Now().UnixNano(), rand.Uint32())
}

// Add appends a job to the queue, assigning it an ID if empty.
func (q *JobQueue) Add(job Job) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if job.ID == "" {
		job.ID = generateID()
	}
	if job.Status == "" {
		job.Status = JobStatusPending
	}
	q.Jobs = append(q.Jobs, job)
}

// Next returns the next pending job and marks it as encoding.
// Returns nil, false if no pending jobs remain.
func (q *JobQueue) Next() (*Job, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i := range q.Jobs {
		if q.Jobs[i].Status == JobStatusPending {
			q.Jobs[i].Status = JobStatusEncoding
			q.Jobs[i].StartedAt = time.Now()
			job := q.Jobs[i] // copy
			return &job, true
		}
	}
	return nil, false
}

// Update applies a mutation function to the job with the given ID.
func (q *JobQueue) Update(id string, fn func(*Job)) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i := range q.Jobs {
		if q.Jobs[i].ID == id {
			fn(&q.Jobs[i])
			return
		}
	}
}

// ByStatus returns a copy of all jobs with the given status.
func (q *JobQueue) ByStatus(status JobStatus) []Job {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	var result []Job
	for _, j := range q.Jobs {
		if j.Status == status {
			result = append(result, j)
		}
	}
	return result
}

// Stats returns aggregate statistics for the queue.
func (q *JobQueue) Stats() QueueStats {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	var s QueueStats
	s.Total = len(q.Jobs)
	for _, j := range q.Jobs {
		switch j.Status {
		case JobStatusPending:
			s.Pending++
		case JobStatusEncoding:
			s.Encoding++
		case JobStatusComplete:
			s.Complete++
		case JobStatusFailed:
			s.Failed++
		case JobStatusSkipped:
			s.Skipped++
		}
		s.TotalInputSize += j.InputSize
		s.TotalOutputSize += j.OutputSize
	}
	return s
}

// Len returns the total number of jobs.
func (q *JobQueue) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.Jobs)
}

// Get returns a copy of the job at the given index.
func (q *JobQueue) Get(index int) (Job, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if index < 0 || index >= len(q.Jobs) {
		return Job{}, false
	}
	return q.Jobs[index], true
}

// All returns a copy of all jobs.
func (q *JobQueue) All() []Job {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	result := make([]Job, len(q.Jobs))
	copy(result, q.Jobs)
	return result
}

// Remove removes a job by ID.
func (q *JobQueue) Remove(id string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i := range q.Jobs {
		if q.Jobs[i].ID == id {
			q.Jobs = append(q.Jobs[:i], q.Jobs[i+1:]...)
			return
		}
	}
}

// Swap swaps the positions of two jobs by index.
func (q *JobQueue) Swap(i, j int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if i < 0 || j < 0 || i >= len(q.Jobs) || j >= len(q.Jobs) {
		return
	}
	q.Jobs[i], q.Jobs[j] = q.Jobs[j], q.Jobs[i]
}

// SkipOptions configures which jobs should be skipped.
type SkipOptions struct {
	SkipExisting  bool    // skip if output file already exists
	SkipOptimal   bool    // skip if source already uses target codec and bitrate is acceptable
	SizeThreshold float64 // skip if source is below this size in MB (0 = disabled)
}

// ShouldSkip determines whether a job should be skipped based on the given options.
// Returns true and a reason string if the job should be skipped.
func ShouldSkip(job Job, opts SkipOptions) (bool, string) {
	// Check if output file already exists
	if opts.SkipExisting && job.OutputPath != "" {
		if _, err := os.Stat(job.OutputPath); err == nil {
			return true, "output file already exists"
		}
	}

	// Check size threshold
	if opts.SizeThreshold > 0 {
		thresholdBytes := int64(opts.SizeThreshold * 1024 * 1024)
		if job.InputSize > 0 && job.InputSize < thresholdBytes {
			return true, fmt.Sprintf("input file below %.1f MB threshold", opts.SizeThreshold)
		}
	}

	// SkipOptimal requires probing — handled externally when source info is available
	// The caller can check source codec/bitrate vs preset target before calling this.

	return false, ""
}

// ShouldSkipOptimal checks if the source video is already optimal for the given preset.
// Returns true if the source already uses the target codec and its bitrate is below
// the preset's expected bitrate.
func ShouldSkipOptimal(info *VideoInfo, targetCodec string, targetCRF int) (bool, string) {
	if info == nil {
		return false, ""
	}

	// Check if source already uses the target codec
	sourceCodec := normalizeCodecName(info.Codec)
	if sourceCodec != targetCodec {
		return false, ""
	}

	// If the source uses the same codec, check if its bitrate is already low
	// Use the BPP estimate to determine if source is already well-compressed
	estimatedBPP := lookupBPP(targetCodec, targetCRF)
	if info.Bitrate <= 0 || info.Width <= 0 || info.Height <= 0 || info.Framerate <= 0 {
		return false, ""
	}

	sourceBPP := float64(info.Bitrate) / (float64(info.Width) * float64(info.Height) * info.Framerate)
	if sourceBPP <= estimatedBPP*1.1 {
		return true, fmt.Sprintf("source already compressed with %s at %.4f bpp (target: %.4f bpp)", sourceCodec, sourceBPP, estimatedBPP)
	}

	return false, ""
}

// Sort sorts jobs by the given mode. Valid modes: "size-asc" (default),
// "size-desc", "name", "duration".
func Sort(jobs []Job, mode string) {
	switch strings.ToLower(mode) {
	case "size-desc":
		sort.SliceStable(jobs, func(i, j int) bool {
			return jobs[i].InputSize > jobs[j].InputSize
		})
	case "name":
		sort.SliceStable(jobs, func(i, j int) bool {
			return strings.ToLower(filepath.Base(jobs[i].InputPath)) <
				strings.ToLower(filepath.Base(jobs[j].InputPath))
		})
	case "duration":
		sort.SliceStable(jobs, func(i, j int) bool {
			return jobs[i].Duration < jobs[j].Duration
		})
	default: // "size-asc" — smallest first for quick wins
		sort.SliceStable(jobs, func(i, j int) bool {
			return jobs[i].InputSize < jobs[j].InputSize
		})
	}
}

// SortQueue sorts the queue's jobs in place by the given mode.
func (q *JobQueue) SortQueue(mode string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	Sort(q.Jobs, mode)
}

// normalizeCodecName maps ffprobe codec names to our canonical names.
func normalizeCodecName(codec string) string {
	switch codec {
	case "h264", "avc":
		return "h264"
	case "h265", "hevc":
		return "h265"
	case "av1":
		return "av1"
	case "vp9":
		return "vp9"
	default:
		return codec
	}
}
