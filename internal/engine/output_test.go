package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOutput_SuffixMode(t *testing.T) {
	opts := OutputOptions{
		Mode:   OutputModeSuffix,
		Suffix: "_shrunk",
	}

	output, err := ResolveOutput("/tmp/video.mp4", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/tmp/video_shrunk.mp4"
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestResolveOutput_SuffixMode_DefaultSuffix(t *testing.T) {
	opts := OutputOptions{
		Mode: OutputModeSuffix,
	}

	output, err := ResolveOutput("/tmp/video.mp4", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "_shrunk") {
		t.Errorf("expected _shrunk suffix, got %s", output)
	}
}

func TestResolveOutput_ExplicitMode(t *testing.T) {
	opts := OutputOptions{
		Mode:         OutputModeExplicit,
		ExplicitPath: "/tmp/output.mp4",
	}

	output, err := ResolveOutput("/tmp/input.mp4", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != "/tmp/output.mp4" {
		t.Errorf("expected /tmp/output.mp4, got %s", output)
	}
}

func TestResolveOutput_ExplicitMode_NoPath(t *testing.T) {
	opts := OutputOptions{
		Mode: OutputModeExplicit,
	}

	_, err := ResolveOutput("/tmp/input.mp4", opts)
	if err == nil {
		t.Error("expected error for explicit mode with no path")
	}
}

func TestResolveOutput_DirectoryMode(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output")

	opts := OutputOptions{
		Mode:      OutputModeDirectory,
		Directory: outDir,
	}

	output, err := ResolveOutput("/tmp/video.mp4", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(outDir, "video.mp4")
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestResolveOutput_ConflictSkip(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "video_shrunk.mp4")
	if err := os.WriteFile(existingFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	input := filepath.Join(tmpDir, "video.mp4")
	opts := OutputOptions{
		Mode:         OutputModeSuffix,
		Suffix:       "_shrunk",
		ConflictMode: ConflictSkip,
	}

	_, err := ResolveOutput(input, opts)
	if err != ErrSkipExisting {
		t.Errorf("expected ErrSkipExisting, got %v", err)
	}
}

func TestResolveOutput_ConflictAutorename(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "video_shrunk.mp4")
	if err := os.WriteFile(existingFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	input := filepath.Join(tmpDir, "video.mp4")
	opts := OutputOptions{
		Mode:         OutputModeSuffix,
		Suffix:       "_shrunk",
		ConflictMode: ConflictAutorename,
	}

	output, err := ResolveOutput(input, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpDir, "video_shrunk(1).mp4")
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestResolveOutput_DirectoryMode_MirrorsPath(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "compressed")
	baseDir := filepath.Join(tmpDir, "source")

	input := filepath.Join(baseDir, "vacation", "clip.mp4")

	opts := OutputOptions{
		Mode:      OutputModeDirectory,
		Directory: outDir,
		BaseDir:   baseDir,
	}

	output, err := ResolveOutput(input, opts)
	if err != nil {
		t.Fatalf("ResolveOutput failed: %v", err)
	}

	expected := filepath.Join(outDir, "vacation", "clip.mp4")
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}

	// Verify intermediate directory was created
	dir := filepath.Dir(output)
	if _, dirErr := os.Stat(dir); dirErr != nil {
		t.Errorf("expected directory %s to exist", dir)
	}
}

func TestResolveOutput_InplaceMode(t *testing.T) {
	input := "/videos/source.mp4"
	opts := OutputOptions{
		Mode: OutputModeInplace,
	}

	output, err := ResolveOutput(input, opts)
	if err != nil {
		t.Fatalf("ResolveOutput failed: %v", err)
	}

	if output != input {
		t.Errorf("expected %s, got %s", input, output)
	}
}

func TestTempPath(t *testing.T) {
	tp := TempPath("/tmp/output.mp4")
	expected := "/tmp/output.mp4.shrinkray.tmp"
	if tp != expected {
		t.Errorf("expected %s, got %s", expected, tp)
	}
}
