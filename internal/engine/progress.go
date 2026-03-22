package engine

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"
)

// ParseProgress reads FFmpeg's -progress pipe:1 output and sends
// ProgressUpdate values on the returned channel. The channel is closed
// when the reader is exhausted or progress=end is encountered.
func ParseProgress(reader io.Reader, duration time.Duration) <-chan ProgressUpdate {
	ch := make(chan ProgressUpdate, 1)

	go func() {
		defer close(ch)

		scanner := bufio.NewScanner(reader)
		startTime := time.Now()

		var current ProgressUpdate

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			switch key {
			case "frame":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					current.Frame = v
				}
			case "fps":
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					current.FPS = v
				}
			case "bitrate":
				current.Bitrate = value
			case "total_size":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					current.Size = v
				}
			case "out_time_us":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					outTime := time.Duration(v) * time.Microsecond
					elapsed := time.Since(startTime)
					current.TimeElapsed = elapsed

					if duration > 0 {
						pct := float64(outTime) / float64(duration) * 100
						if pct > 100 {
							pct = 100
						}
						if pct < 0 {
							pct = 0
						}
						current.Percent = pct

						// Calculate ETA
						if pct > 0 {
							totalEstimated := time.Duration(float64(elapsed) / pct * 100)
							current.ETA = totalEstimated - elapsed
							if current.ETA < 0 {
								current.ETA = 0
							}
						}
					}
				}
			case "speed":
				// value is like "1.5x" or "N/A"
				value = strings.TrimSuffix(value, "x")
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					current.Speed = v
				}
			case "progress":
				if value == "end" {
					current.Done = true
					current.Percent = 100
					// Final update must be delivered — use blocking send
					ch <- current
					return
				}
				// Non-blocking send for intermediate updates to prevent TUI stalls
				select {
				case ch <- current:
				default:
					// Drop update if consumer is slow — they'll get the next one
				}

				// Reset for next block (keep elapsed tracking)
				current = ProgressUpdate{
					Pass: current.Pass,
				}
			}
		}

		// If we reached EOF without progress=end, send final update
		if !current.Done {
			current.Done = true
			select {
			case ch <- current:
			default:
			}
		}
	}()

	return ch
}
