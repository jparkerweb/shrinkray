package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutorename_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the base file
	basePath := filepath.Join(tmpDir, "video_shrunk.mp4")
	if err := os.WriteFile(basePath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Autorename(basePath)
	expected := filepath.Join(tmpDir, "video_shrunk(1).mp4")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestAutorename_MultipleExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base and (1) and (2)
	basePath := filepath.Join(tmpDir, "video_shrunk.mp4")
	for _, name := range []string{"video_shrunk.mp4", "video_shrunk(1).mp4", "video_shrunk(2).mp4"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	result := Autorename(basePath)
	expected := filepath.Join(tmpDir, "video_shrunk(3).mp4")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestAutorename_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "video_shrunk.mp4")
	// File doesn't exist, so autorename should return (1)
	result := Autorename(basePath)
	expected := filepath.Join(tmpDir, "video_shrunk(1).mp4")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestAutorename_Cap99(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "video_shrunk.mp4")

	// Create base file and files (1) through (99)
	if err := os.WriteFile(basePath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 99; i++ {
		name := filepath.Join(tmpDir, fmt.Sprintf("video_shrunk(%d).mp4", i))
		if err := os.WriteFile(name, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	result := Autorename(basePath)
	// When all 99 are taken, it returns (99) as fallback
	if !strings.Contains(result, "(99)") {
		t.Errorf("expected (99) fallback, got %s", result)
	}
}

func TestMetadataFromFlags(t *testing.T) {
	// Strip metadata
	m := MetadataFromFlags(true, false, "")
	if !m.StripAll {
		t.Error("expected StripAll true")
	}

	// Keep metadata (default)
	m = MetadataFromFlags(false, true, "")
	if m.StripAll {
		t.Error("expected StripAll false")
	}

	// Title
	m = MetadataFromFlags(false, true, "My Video")
	if m.Title != "My Video" {
		t.Errorf("expected title 'My Video', got %q", m.Title)
	}
}

func TestBuildArgs_MetadataStrip(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		MetadataMode: MetadataMode{
			StripAll: true,
		},
	}
	// Set a minimal preset
	opts.Preset.Codec = "h264"
	opts.Preset.AudioCodec = "aac"
	opts.Preset.AudioBitrate = "128k"

	args := BuildArgs(opts)
	found := false
	for i, arg := range args {
		if arg == "-map_metadata" && i+1 < len(args) && args[i+1] == "-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected -map_metadata -1 in args for strip metadata")
	}
}

func TestBuildArgs_MetadataTitle(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
		MetadataMode: MetadataMode{
			Title: "Test Title",
		},
	}
	opts.Preset.Codec = "h264"
	opts.Preset.AudioCodec = "aac"
	opts.Preset.AudioBitrate = "128k"

	args := BuildArgs(opts)
	found := false
	for i, arg := range args {
		if arg == "-metadata" && i+1 < len(args) && args[i+1] == "title=Test Title" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected -metadata title=Test Title in args, got %v", args)
	}
}

func TestBuildArgs_MetadataKeep(t *testing.T) {
	opts := EncodeOptions{
		Input:  "input.mp4",
		Output: "output.mp4",
	}
	opts.Preset.Codec = "h264"
	opts.Preset.AudioCodec = "aac"
	opts.Preset.AudioBitrate = "128k"

	args := BuildArgs(opts)
	found := false
	for i, arg := range args {
		if arg == "-map_metadata" && i+1 < len(args) && args[i+1] == "0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected -map_metadata 0 in args for default (keep) metadata")
	}
}
