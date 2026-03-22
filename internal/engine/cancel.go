package engine

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// Cancel performs a cross-platform graceful shutdown of an FFmpeg process.
// It follows this escalation:
//  1. Write 'q' to FFmpeg stdin pipe (if provided)
//  2. Wait 3 seconds
//  3. Send os.Interrupt (Unix) or Kill (Windows)
//  4. Wait 2 seconds
//  5. Force kill
func Cancel(proc *os.Process, stdin io.WriteCloser) error {
	if proc == nil {
		return nil
	}

	// Step 1: Try 'q' command via stdin
	if stdin != nil {
		_, _ = io.WriteString(stdin, "q\n")
		_ = stdin.Close()
	}

	// Step 2: Wait for graceful exit
	done := make(chan struct{})
	go func() {
		// Try to wait for the process
		_, _ = proc.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(3 * time.Second):
	}

	// Step 3: Send interrupt (Unix) or kill (Windows)
	if runtime.GOOS != "windows" {
		_ = proc.Signal(os.Interrupt)
	} else {
		_ = proc.Kill()
		return nil
	}

	// Step 4: Wait again
	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
	}

	// Step 5: Force kill
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to kill ffmpeg process: %w", err)
	}

	return nil
}
