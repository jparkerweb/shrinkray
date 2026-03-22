package engine

import "testing"

func TestDetectHDR_SDR(t *testing.T) {
	info := &VideoInfo{
		ColorTransfer: "bt709",
		Codec:         "h264",
	}

	isHDR, format := DetectHDR(info)
	if isHDR {
		t.Error("expected SDR for bt709 transfer")
	}
	if format != "SDR" {
		t.Errorf("expected format SDR, got %s", format)
	}
}

func TestDetectHDR_HDR10(t *testing.T) {
	info := &VideoInfo{
		ColorTransfer: "smpte2084",
		Codec:         "hevc",
	}

	isHDR, format := DetectHDR(info)
	if !isHDR {
		t.Error("expected HDR for smpte2084 transfer")
	}
	if format != "HDR10" {
		t.Errorf("expected format HDR10, got %s", format)
	}
}

func TestDetectHDR_HLG(t *testing.T) {
	info := &VideoInfo{
		ColorTransfer: "arib-std-b67",
		Codec:         "hevc",
	}

	isHDR, format := DetectHDR(info)
	if !isHDR {
		t.Error("expected HDR for arib-std-b67 transfer")
	}
	if format != "HLG" {
		t.Errorf("expected format HLG, got %s", format)
	}
}

func TestDetectHDR_DolbyVision(t *testing.T) {
	info := &VideoInfo{
		ColorTransfer: "",
		Codec:         "dvhe",
	}

	isHDR, format := DetectHDR(info)
	if !isHDR {
		t.Error("expected HDR for Dolby Vision codec")
	}
	if format != "Dolby Vision" {
		t.Errorf("expected format Dolby Vision, got %s", format)
	}
}

func TestDetectHDR_Nil(t *testing.T) {
	isHDR, format := DetectHDR(nil)
	if isHDR {
		t.Error("expected SDR for nil input")
	}
	if format != "SDR" {
		t.Errorf("expected format SDR, got %s", format)
	}
}
