package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// EncodeError wraps an FFmpeg error with stderr output for debugging.
type EncodeError struct {
	ExitCode int
	Stderr   string
	Err      error
}

func (e *EncodeError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("ffmpeg exited with code %d: %s", e.ExitCode, e.Stderr)
	}
	return fmt.Sprintf("ffmpeg exited with code %d: %v", e.ExitCode, e.Err)
}

func (e *EncodeError) Unwrap() error {
	return e.Err
}

// Encode starts an FFmpeg encoding process and returns a channel of progress updates.
// The process is tied to the given context; cancelling the context will gracefully
// stop the encode.
func Encode(ctx context.Context, opts EncodeOptions) (<-chan ProgressUpdate, error) {
	args := BuildArgs(opts)

	slog.Debug("starting ffmpeg", "args", args)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Stdin pipe for sending 'q' to gracefully stop
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Stdout pipe for reading progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for error reporting
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Determine duration for progress calculation
	duration := time.Duration(0)
	// Probe the input to get duration
	probeInfo, probeErr := Probe(context.Background(), opts.Input)
	if probeErr == nil && probeInfo.Duration > 0 {
		duration = time.Duration(probeInfo.Duration * float64(time.Second))
	}

	progressCh := ParseProgress(stdout, duration)

	// Relay channel — adds error handling and process cleanup
	outCh := make(chan ProgressUpdate, 1)

	go func() {
		defer close(outCh)

		for update := range progressCh {
			update.Pass = opts.Pass
			select {
			case outCh <- update:
			default:
			}
		}

		// Wait for process to finish
		waitErr := cmd.Wait()

		if waitErr != nil {
			// Check if it was a context cancellation
			if ctx.Err() != nil {
				slog.Debug("encode cancelled by context")
				return
			}

			exitCode := 1
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}

			rawStderr := stderrBuf.String()
			friendlyMsg := ParseFFmpegError(rawStderr)
			encErr := &EncodeError{
				ExitCode: exitCode,
				Stderr:   friendlyMsg,
				Err:      waitErr,
			}

			select {
			case outCh <- ProgressUpdate{Done: true, Error: encErr}:
			default:
			}
			return
		}

		// Send final done if not already sent
		select {
		case outCh <- ProgressUpdate{Done: true, Percent: 100, Pass: opts.Pass}:
		default:
		}
	}()

	// Handle context cancellation — gracefully stop FFmpeg
	go func() {
		<-ctx.Done()
		if ctx.Err() != nil {
			slog.Debug("context cancelled, sending q to ffmpeg")
			gracefulShutdown(stdin, cmd.Process)
		}
	}()

	return outCh, nil
}

// EncodeTwoPass performs a two-pass encode, reporting combined progress 0-100%.
// Pass 1 maps to 0-50%, pass 2 maps to 50-100%.
// It cleans up the passlog file after completion.
func EncodeTwoPass(ctx context.Context, opts EncodeOptions) (<-chan ProgressUpdate, error) {
	// Generate passlog file path
	outputDir := filepath.Dir(opts.Output)
	passLogBase := filepath.Join(outputDir, fmt.Sprintf("shrinkray_%d.passlog",
		time.Now().UnixNano()))
	opts.PassLogFile = passLogBase

	outCh := make(chan ProgressUpdate, 1)

	go func() {
		defer close(outCh)
		defer cleanupPassLog(passLogBase)

		// === Pass 1 ===
		pass1Opts := opts
		pass1Opts.Pass = 1

		slog.Info("starting two-pass encode: pass 1")
		pass1Ch, err := Encode(ctx, pass1Opts)
		if err != nil {
			outCh <- ProgressUpdate{Done: true, Error: fmt.Errorf("pass 1 failed to start: %w", err)}
			return
		}

		for update := range pass1Ch {
			if update.Error != nil {
				outCh <- ProgressUpdate{Done: true, Error: fmt.Errorf("pass 1 failed: %w", update.Error)}
				return
			}
			if update.Done {
				// Send 50% mark
				outCh <- ProgressUpdate{
					Percent: 50,
					Pass:    1,
					FPS:     update.FPS,
					Speed:   update.Speed,
					Bitrate: update.Bitrate,
				}
				break
			}
			// Map pass 1 progress to 0-50%
			mapped := update
			mapped.Percent = update.Percent / 2
			mapped.Pass = 1
			select {
			case outCh <- mapped:
			default:
			}
		}

		// Check for cancellation between passes
		if ctx.Err() != nil {
			return
		}

		// === Pass 2 ===
		pass2Opts := opts
		pass2Opts.Pass = 2

		slog.Info("starting two-pass encode: pass 2")
		pass2Ch, err := Encode(ctx, pass2Opts)
		if err != nil {
			outCh <- ProgressUpdate{Done: true, Error: fmt.Errorf("pass 2 failed to start: %w", err)}
			return
		}

		for update := range pass2Ch {
			if update.Error != nil {
				outCh <- ProgressUpdate{Done: true, Error: fmt.Errorf("pass 2 failed: %w", update.Error)}
				return
			}
			if update.Done {
				// Verify output size if target was set
				if opts.Preset.TargetSizeMB > 0 {
					verifyOutputSize(opts.Output, opts.Preset.TargetSizeMB)
				}

				outCh <- ProgressUpdate{
					Done:    true,
					Percent: 100,
					Pass:    2,
				}
				return
			}
			// Map pass 2 progress to 50-100%
			mapped := update
			mapped.Percent = 50 + update.Percent/2
			mapped.Pass = 2
			select {
			case outCh <- mapped:
			default:
			}
		}
	}()

	return outCh, nil
}

// cleanupPassLog removes passlog files generated by FFmpeg two-pass encoding.
func cleanupPassLog(basePath string) {
	// FFmpeg creates files like basePath-0.log, basePath-0.log.mbtree
	patterns := []string{
		basePath + "*",
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				slog.Debug("failed to remove passlog file", "path", match, "error", err)
			}
		}
	}
}

// verifyOutputSize checks if the output file size is within 3% of the target.
func verifyOutputSize(outputPath string, targetMB float64) {
	stat, err := os.Stat(outputPath)
	if err != nil {
		return
	}

	targetBytes := targetMB * 1024 * 1024
	actualBytes := float64(stat.Size())
	overshoot := (actualBytes - targetBytes) / targetBytes * 100

	if overshoot > 3 {
		slog.Warn("output exceeds target size",
			"target_mb", targetMB,
			"actual_mb", actualBytes/1024/1024,
			"overshoot_pct", fmt.Sprintf("%.1f%%", overshoot),
		)
	}
}

// ShouldUseTwoPass returns true if the encode options require two-pass encoding.
func ShouldUseTwoPass(opts EncodeOptions) bool {
	return opts.Preset.TwoPass || opts.Preset.TargetSizeMB > 0
}

// gracefulShutdown tries to stop FFmpeg cleanly by writing 'q' to stdin,
// then escalates to killing the process.
func gracefulShutdown(stdin io.WriteCloser, proc interface{ Kill() error }) {
	// Try sending 'q' to FFmpeg
	if stdin != nil {
		_, _ = io.WriteString(stdin, "q\n")
		_ = stdin.Close()
	}

	// Give FFmpeg time to finish gracefully
	time.Sleep(3 * time.Second)

	// Force kill if still running
	if proc != nil {
		_ = proc.Kill()
	}
}
