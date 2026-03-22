package engine

import (
	"strings"
	"testing"
	"time"
)

func TestParseProgress_BasicOutput(t *testing.T) {
	// Simulate FFmpeg -progress pipe:1 output
	input := `frame=100
fps=30.0
bitrate=2500kbits/s
total_size=500000
out_time_us=3333333
speed=1.5x
progress=continue

frame=200
fps=29.5
bitrate=2400kbits/s
total_size=1000000
out_time_us=6666666
speed=1.4x
progress=continue

frame=300
fps=30.0
bitrate=2450kbits/s
total_size=1500000
out_time_us=10000000
speed=1.5x
progress=end
`

	duration := 10 * time.Second
	ch := ParseProgress(strings.NewReader(input), duration)

	var updates []ProgressUpdate
	for u := range ch {
		updates = append(updates, u)
	}

	if len(updates) < 2 {
		t.Fatalf("expected at least 2 updates, got %d", len(updates))
	}

	// First update should be around 33%
	first := updates[0]
	if first.Percent < 30 || first.Percent > 40 {
		t.Errorf("first update percent expected ~33%%, got %.1f%%", first.Percent)
	}
	if first.Frame != 100 {
		t.Errorf("first update frame expected 100, got %d", first.Frame)
	}
	if first.Speed != 1.5 {
		t.Errorf("first update speed expected 1.5, got %f", first.Speed)
	}
	if first.Bitrate != "2500kbits/s" {
		t.Errorf("first update bitrate expected 2500kbits/s, got %s", first.Bitrate)
	}

	// Last update should be done
	last := updates[len(updates)-1]
	if !last.Done {
		t.Error("last update should have Done=true")
	}
	if last.Percent != 100 {
		t.Errorf("last update percent expected 100, got %.1f", last.Percent)
	}
}

func TestParseProgress_EmptyReader(t *testing.T) {
	ch := ParseProgress(strings.NewReader(""), 10*time.Second)

	var updates []ProgressUpdate
	for u := range ch {
		updates = append(updates, u)
	}

	// Should get at least a final done update
	if len(updates) < 1 {
		t.Fatal("expected at least 1 update from empty reader")
	}
	if !updates[len(updates)-1].Done {
		t.Error("final update should be done")
	}
}

func TestParseProgress_ZeroDuration(t *testing.T) {
	input := `frame=50
fps=25.0
out_time_us=2000000
progress=continue

progress=end
`

	ch := ParseProgress(strings.NewReader(input), 0)

	var updates []ProgressUpdate
	for u := range ch {
		updates = append(updates, u)
	}

	// With zero duration, percent should be 0
	if len(updates) > 0 && updates[0].Percent != 0 {
		t.Errorf("with zero duration, percent should be 0, got %.1f", updates[0].Percent)
	}
}
