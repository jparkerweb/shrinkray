package presets

import (
	"os"
	"path/filepath"
	"testing"
)

// ---- Preset catalog tests ----

func TestAll_Returns18Presets(t *testing.T) {
	all := All()
	if len(all) != 18 {
		t.Errorf("expected 18 total presets, got %d", len(all))
		for _, p := range all {
			t.Logf("  %s (%s)", p.Key, p.Category)
		}
	}
}

func TestByCategory_Quality_Returns6(t *testing.T) {
	quality := ByCategory(CategoryQuality)
	if len(quality) != 6 {
		t.Errorf("expected 6 quality presets, got %d", len(quality))
	}
}

func TestByCategory_Purpose_Returns5(t *testing.T) {
	purpose := ByCategory(CategoryPurpose)
	if len(purpose) != 5 {
		t.Errorf("expected 5 purpose presets, got %d", len(purpose))
	}
}

func TestByCategory_Platform_Returns7(t *testing.T) {
	platform := ByCategory(CategoryPlatform)
	if len(platform) != 7 {
		t.Errorf("expected 7 platform presets, got %d", len(platform))
	}
}

func TestPurposePresets_ValidFields(t *testing.T) {
	purpose := ByCategory(CategoryPurpose)
	for _, p := range purpose {
		t.Run(p.Key, func(t *testing.T) {
			if p.Key == "" {
				t.Error("preset Key is empty")
			}
			if p.Name == "" {
				t.Error("preset Name is empty")
			}
			if p.Description == "" {
				t.Error("preset Description is empty")
			}
			if p.Codec == "" {
				t.Error("preset Codec is empty")
			}
			if p.Container == "" {
				t.Error("preset Container is empty")
			}
			if p.Icon == "" {
				t.Error("preset Icon is empty")
			}
			if len(p.Tags) == 0 {
				t.Error("preset Tags is empty")
			}
			if p.Category != CategoryPurpose {
				t.Errorf("expected category purpose, got %s", p.Category)
			}
		})
	}
}

func TestPlatformPresets_ValidFields(t *testing.T) {
	platform := ByCategory(CategoryPlatform)
	for _, p := range platform {
		t.Run(p.Key, func(t *testing.T) {
			if p.Key == "" {
				t.Error("preset Key is empty")
			}
			if p.Name == "" {
				t.Error("preset Name is empty")
			}
			if p.Description == "" {
				t.Error("preset Description is empty")
			}
			if p.Codec == "" {
				t.Error("preset Codec is empty")
			}
			if p.Container == "" {
				t.Error("preset Container is empty")
			}
			if p.Icon == "" {
				t.Error("preset Icon is empty")
			}
			if len(p.Tags) == 0 {
				t.Error("preset Tags is empty")
			}
			if p.Category != CategoryPlatform {
				t.Errorf("expected category platform, got %s", p.Category)
			}
		})
	}
}

func TestPurposePreset_Web(t *testing.T) {
	p, found := Lookup("web")
	if !found {
		t.Fatal("web preset not found")
	}
	if p.Codec != "h264" {
		t.Errorf("expected h264, got %s", p.Codec)
	}
	if p.CRF != 23 {
		t.Errorf("expected CRF 23, got %d", p.CRF)
	}
	if p.Resolution != "1920x1080" {
		t.Errorf("expected 1920x1080, got %s", p.Resolution)
	}
}

func TestPurposePreset_Email(t *testing.T) {
	p, found := Lookup("email")
	if !found {
		t.Fatal("email preset not found")
	}
	if p.TargetSizeMB != 20 {
		t.Errorf("expected target 20MB, got %.0f", p.TargetSizeMB)
	}
	if !p.TwoPass {
		t.Error("email preset should use two-pass")
	}
}

func TestPurposePreset_Archive(t *testing.T) {
	p, found := Lookup("archive")
	if !found {
		t.Fatal("archive preset not found")
	}
	if p.Codec != "h265" {
		t.Errorf("expected h265, got %s", p.Codec)
	}
	if p.AudioCodec != "copy" {
		t.Errorf("expected audio copy, got %s", p.AudioCodec)
	}
}

func TestPlatformPreset_Discord(t *testing.T) {
	p, found := Lookup("discord")
	if !found {
		t.Fatal("discord preset not found")
	}
	if p.TargetSizeMB != 10 {
		t.Errorf("expected target 10MB, got %.0f", p.TargetSizeMB)
	}
	if !p.TwoPass {
		t.Error("discord preset should use two-pass")
	}
	if p.Resolution != "1280x720" {
		t.Errorf("expected 1280x720, got %s", p.Resolution)
	}
}

func TestPlatformPreset_DiscordNitro(t *testing.T) {
	p, found := Lookup("discord-nitro")
	if !found {
		t.Fatal("discord-nitro preset not found")
	}
	if p.TargetSizeMB != 50 {
		t.Errorf("expected target 50MB, got %.0f", p.TargetSizeMB)
	}
}

func TestPlatformPreset_WhatsApp(t *testing.T) {
	p, found := Lookup("whatsapp")
	if !found {
		t.Fatal("whatsapp preset not found")
	}
	if p.MaxFPS != 30 {
		t.Errorf("expected MaxFPS 30, got %d", p.MaxFPS)
	}
	if p.TargetSizeMB != 16 {
		t.Errorf("expected target 16MB, got %.0f", p.TargetSizeMB)
	}
}

func TestPlatformPreset_YouTube(t *testing.T) {
	p, found := Lookup("youtube")
	if !found {
		t.Fatal("youtube preset not found")
	}
	if p.Codec != "h265" {
		t.Errorf("expected h265, got %s", p.Codec)
	}
	if p.CRF != 18 {
		t.Errorf("expected CRF 18, got %d", p.CRF)
	}
	if p.Resolution != "" {
		t.Errorf("youtube should keep source resolution, got %s", p.Resolution)
	}
}

// ---- Recommendation tests ----

func TestRecommend_4KSource(t *testing.T) {
	info := &VideoMetadata{
		Width:     3840,
		Height:    2160,
		Framerate: 30,
		Bitrate:   50000000,
		Duration:  60,
		Size:      375000000,
		Codec:     "h264",
	}

	recs := Recommend(info)
	if len(recs) == 0 {
		t.Fatal("expected recommendations for 4K source")
	}

	// 4k-to-1080 should be top recommendation
	found := false
	for _, r := range recs {
		if r.Preset.Key == "4k-to-1080" {
			found = true
			if r.Score < 80 {
				t.Errorf("4k-to-1080 score should be high for 4K source, got %d", r.Score)
			}
			break
		}
	}
	if !found {
		t.Error("expected 4k-to-1080 in recommendations for 4K source")
	}
}

func TestRecommend_ShortClip(t *testing.T) {
	info := &VideoMetadata{
		Width:     1920,
		Height:    1080,
		Framerate: 30,
		Bitrate:   8000000,
		Duration:  30,
		Size:      30000000,
		Codec:     "h264",
	}

	recs := Recommend(info)
	if len(recs) == 0 {
		t.Fatal("expected recommendations for short clip")
	}

	// Short clips should suggest platform presets
	hasPlatform := false
	for _, r := range recs {
		if r.Preset.Category == CategoryPlatform {
			hasPlatform = true
			break
		}
	}
	if !hasPlatform {
		t.Error("expected platform preset in recommendations for short clip")
	}
}

func TestRecommend_AlreadyCompressed(t *testing.T) {
	info := &VideoMetadata{
		Width:     1920,
		Height:    1080,
		Framerate: 30,
		Bitrate:   2000000, // low bitrate for h265
		Duration:  60,
		Size:      15000000,
		Codec:     "hevc",
	}

	recs := Recommend(info)
	if len(recs) == 0 {
		t.Fatal("expected recommendations even for compressed source")
	}

	// Scores should be lower for already-compressed content
	for _, r := range recs {
		if r.Score > 80 {
			t.Errorf("score should be reduced for well-compressed source, got %d for %s", r.Score, r.Preset.Key)
		}
	}
}

func TestRecommend_Returns5Max(t *testing.T) {
	info := &VideoMetadata{
		Width:     1920,
		Height:    1080,
		Framerate: 30,
		Bitrate:   20000000,
		Duration:  300,
		Size:      750000000,
		Codec:     "h264",
	}

	recs := Recommend(info)
	if len(recs) > 5 {
		t.Errorf("expected max 5 recommendations, got %d", len(recs))
	}
}

func TestRecommend_NilInfo(t *testing.T) {
	recs := Recommend(nil)
	if recs != nil {
		t.Error("expected nil recommendations for nil info")
	}
}

// ---- Custom presets tests ----

func TestLoadCustomPresets_NoFile(t *testing.T) {
	dir := t.TempDir()
	result, err := LoadCustomPresets(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for no file, got %d presets", len(result))
	}
}

func TestSaveAndLoadCustomPresets(t *testing.T) {
	dir := t.TempDir()

	preset := Preset{
		Key:          "test-custom-roundtrip",
		Name:         "Test Custom",
		Description:  "A test custom preset",
		Category:     CategoryQuality,
		Codec:        "h264",
		Container:    "mp4",
		CRF:          25,
		AudioCodec:   "aac",
		AudioBitrate: "128k",
		SpeedPreset:  "medium",
		Tags:         []string{"test"},
		Icon:         "\U0001f527",
	}

	// Save
	if err := SaveCustomPreset(dir, preset); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, customPresetsFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("custom presets file not created: %v", err)
	}

	// Load
	loaded, err := LoadCustomPresets(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 loaded preset, got %d", len(loaded))
	}

	if loaded[0].Key != "test-custom-roundtrip" {
		t.Errorf("expected key test-custom-roundtrip, got %s", loaded[0].Key)
	}
	if loaded[0].Codec != "h264" {
		t.Errorf("expected codec h264, got %s", loaded[0].Codec)
	}
	if loaded[0].CRF != 25 {
		t.Errorf("expected CRF 25, got %d", loaded[0].CRF)
	}

	// Clean up: remove the test preset from registry
	var cleaned []Preset
	for _, p := range registry {
		if p.Key != "test-custom-roundtrip" {
			cleaned = append(cleaned, p)
		}
	}
	registry = cleaned
}

func TestSaveCustomPreset_RejectsBuiltinKey(t *testing.T) {
	dir := t.TempDir()
	preset := Preset{
		Key:   "balanced",
		Name:  "Fake balanced",
		Codec: "h264",
		CRF:   23,
	}

	err := SaveCustomPreset(dir, preset)
	if err == nil {
		t.Error("expected error when saving preset with built-in key")
	}
}

func TestSaveCustomPreset_RejectsMissingFields(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name   string
		preset Preset
	}{
		{
			name:   "missing key",
			preset: Preset{Name: "No Key", Codec: "h264", CRF: 23},
		},
		{
			name:   "missing name",
			preset: Preset{Key: "no-name", Codec: "h264", CRF: 23},
		},
		{
			name:   "missing codec",
			preset: Preset{Key: "no-codec", Name: "No Codec", CRF: 23},
		},
		{
			name:   "missing crf and target",
			preset: Preset{Key: "no-crf", Name: "No CRF", Codec: "h264"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveCustomPreset(dir, tt.preset)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
