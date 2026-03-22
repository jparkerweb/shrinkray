package engine

import (
	"runtime"
	"strings"
	"testing"
)

func TestOpenFolderCommand_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	name, args := OpenFolderCommand("/tmp/output.mp4")
	if name != "explorer.exe" {
		t.Errorf("expected explorer.exe, got %s", name)
	}
	if len(args) != 1 || !strings.HasPrefix(args[0], "/select,") {
		t.Errorf("expected [/select,path], got %v", args)
	}
}

func TestOpenFolderCommand_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}
	name, args := OpenFolderCommand("/tmp/output.mp4")
	if name != "open" {
		t.Errorf("expected open, got %s", name)
	}
	if len(args) != 2 || args[0] != "-R" {
		t.Errorf("expected [-R path], got %v", args)
	}
}

func TestOpenFolderCommand_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}
	name, args := OpenFolderCommand("/tmp/output.mp4")
	if name != "xdg-open" {
		t.Errorf("expected xdg-open, got %s", name)
	}
	if len(args) != 1 || args[0] != "/tmp" {
		t.Errorf("expected [/tmp], got %v", args)
	}
}

func TestOpenFolderCommand_ReturnsCorrectPlatform(t *testing.T) {
	// This test just verifies OpenFolderCommand doesn't panic
	name, args := OpenFolderCommand("/some/path/video.mp4")
	if name == "" {
		t.Error("expected non-empty command name")
	}
	if len(args) == 0 {
		t.Error("expected non-empty args")
	}
}
