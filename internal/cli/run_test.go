package cli

import (
	"math"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	registerRunFlags(cmd)
	return cmd
}

func TestValidateFlags_HWAccelMutualExclusion(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--hw-accel", "--no-hw-accel"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for --hw-accel + --no-hw-accel, got nil")
	}
	if err.Error() != "--hw-accel and --no-hw-accel are mutually exclusive" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateFlags_HWAccelAlone(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--hw-accel"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFlags_NoHWAccelAlone(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--no-hw-accel"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFlags_NeitherHWFlag(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFlags_AudioCodecInvalid(t *testing.T) {
	for _, val := range []string{"mp3", "flac"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-codec", val})
		_ = cmd.Execute()

		err := validateFlags(cmd)
		if err == nil {
			t.Fatalf("expected error for --audio-codec=%s, got nil", val)
		}
		if err.Error() != "--audio-codec must be one of: aac, opus, copy, none" {
			t.Fatalf("unexpected error message for %s: %v", val, err)
		}
	}
}

func TestValidateFlags_AudioCodecValid(t *testing.T) {
	for _, val := range []string{"aac", "opus", "copy", "none"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-codec", val})
		_ = cmd.Execute()

		if err := validateFlags(cmd); err != nil {
			t.Fatalf("unexpected error for --audio-codec=%s: %v", val, err)
		}
	}
}

func TestValidateFlags_AudioBitrateInvalid(t *testing.T) {
	for _, val := range []string{"128", "fast", "128kb"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-bitrate", val})
		_ = cmd.Execute()

		err := validateFlags(cmd)
		if err == nil {
			t.Fatalf("expected error for --audio-bitrate=%s, got nil", val)
		}
		if err.Error() != "--audio-bitrate must be in format like 128k" {
			t.Fatalf("unexpected error message for %s: %v", val, err)
		}
	}
}

func TestValidateFlags_AudioBitrateValid(t *testing.T) {
	for _, val := range []string{"64k", "128k", "256k"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-bitrate", val})
		_ = cmd.Execute()

		if err := validateFlags(cmd); err != nil {
			t.Fatalf("unexpected error for --audio-bitrate=%s: %v", val, err)
		}
	}
}

func TestValidateFlags_AudioChannelsInvalid(t *testing.T) {
	for _, val := range []string{"surround", "5.1"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-channels", val})
		_ = cmd.Execute()

		err := validateFlags(cmd)
		if err == nil {
			t.Fatalf("expected error for --audio-channels=%s, got nil", val)
		}
		if err.Error() != "--audio-channels must be one of: stereo, mono, source" {
			t.Fatalf("unexpected error message for %s: %v", val, err)
		}
	}
}

func TestValidateFlags_AudioChannelsValid(t *testing.T) {
	for _, val := range []string{"stereo", "mono", "source"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--audio-channels", val})
		_ = cmd.Execute()

		if err := validateFlags(cmd); err != nil {
			t.Fatalf("unexpected error for --audio-channels=%s: %v", val, err)
		}
	}
}

func TestParseTargetSize(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"25mb", 25.0},
		{"1gb", 1024.0},
		{"2.5gb", 2560.0},
		{"500MB", 500.0},
	}
	for _, tt := range tests {
		got, err := parseTargetSize(tt.input)
		if err != nil {
			t.Errorf("parseTargetSize(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("parseTargetSize(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseTargetSize_Errors(t *testing.T) {
	tests := []string{"25", "mb", "25kb", "abc", "", "-5mb"}
	for _, input := range tests {
		_, err := parseTargetSize(input)
		if err == nil {
			t.Errorf("parseTargetSize(%q) expected error, got nil", input)
		}
	}
}

func TestValidateFlags_SpeedPresetInvalid(t *testing.T) {
	for _, val := range []string{"turbo", "fastest", "invalid"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--speed-preset", val})
		_ = cmd.Execute()

		err := validateFlags(cmd)
		if err == nil {
			t.Fatalf("expected error for --speed-preset=%s, got nil", val)
		}
		if err.Error() != "--speed-preset must be one of: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow, placebo" {
			t.Fatalf("unexpected error message for %s: %v", val, err)
		}
	}
}

func TestValidateFlags_SpeedPresetValid(t *testing.T) {
	for _, val := range []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow", "placebo"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--speed-preset", val})
		_ = cmd.Execute()

		if err := validateFlags(cmd); err != nil {
			t.Fatalf("unexpected error for --speed-preset=%s: %v", val, err)
		}
	}
}

func TestValidateFlags_FPSInvalid(t *testing.T) {
	for _, val := range []string{"0", "-1"} {
		cmd := newTestCommand()
		cmd.SetArgs([]string{"--fps", val})
		_ = cmd.Execute()

		err := validateFlags(cmd)
		if err == nil {
			t.Fatalf("expected error for --fps=%s, got nil", val)
		}
		if err.Error() != "--fps must be a positive integer" {
			t.Fatalf("unexpected error message for --fps=%s: %v", val, err)
		}
	}
}

func TestValidateFlags_FPSValid(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--fps", "30"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error for --fps=30: %v", err)
	}
}

func TestValidateFlags_TargetSizeInvalid(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--target-size", "25kb"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for --target-size=25kb, got nil")
	}
}

func TestValidateFlags_TargetSizeValid(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--target-size", "25mb"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error for --target-size=25mb: %v", err)
	}
}

func TestValidateFlags_OverwritePlusSkipExisting(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--overwrite", "--skip-existing"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for --overwrite + --skip-existing, got nil")
	}
	if err.Error() != "--overwrite, --auto-rename, and --skip-existing are mutually exclusive" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateFlags_AutoRenamePlusSkipExisting(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--auto-rename", "--skip-existing"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for --auto-rename + --skip-existing, got nil")
	}
	if err.Error() != "--overwrite, --auto-rename, and --skip-existing are mutually exclusive" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateFlags_OverwritePlusAutoRename(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--overwrite", "--auto-rename"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for --overwrite + --auto-rename, got nil")
	}
	if err.Error() != "--overwrite, --auto-rename, and --skip-existing are mutually exclusive" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateFlags_AllThreeConflictFlags(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--overwrite", "--auto-rename", "--skip-existing"})
	_ = cmd.Execute()

	err := validateFlags(cmd)
	if err == nil {
		t.Fatal("expected error for all three conflict flags, got nil")
	}
	if err.Error() != "--overwrite, --auto-rename, and --skip-existing are mutually exclusive" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateFlags_OverwriteAlone(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--overwrite"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error for --overwrite alone: %v", err)
	}
}

func TestValidateFlags_AutoRenameAlone(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--auto-rename"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error for --auto-rename alone: %v", err)
	}
}

func TestValidateFlags_SkipExistingAlone(t *testing.T) {
	cmd := newTestCommand()
	cmd.SetArgs([]string{"--skip-existing"})
	_ = cmd.Execute()

	if err := validateFlags(cmd); err != nil {
		t.Fatalf("unexpected error for --skip-existing alone: %v", err)
	}
}
