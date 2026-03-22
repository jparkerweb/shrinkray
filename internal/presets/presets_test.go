package presets

import "testing"

func TestAll_Returns18Presets_Phase1(t *testing.T) {
	all := All()
	if len(all) != 18 {
		t.Errorf("expected 18 presets (6 quality + 5 purpose + 7 platform), got %d", len(all))
	}
}

func TestAll_ReturnsACopy(t *testing.T) {
	all1 := All()
	all2 := All()

	// Modifying one should not affect the other
	all1[0].Name = "modified"
	if all2[0].Name == "modified" {
		t.Error("All() should return a copy, not a reference to the internal slice")
	}
}

func TestLookup_ExactKey(t *testing.T) {
	p, found := Lookup("balanced")
	if !found {
		t.Fatal("expected to find 'balanced' preset")
	}
	if p.Key != "balanced" {
		t.Errorf("expected key 'balanced', got '%s'", p.Key)
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	p, found := Lookup("BALANCED")
	if !found {
		t.Fatal("expected to find 'BALANCED' preset (case-insensitive)")
	}
	if p.Key != "balanced" {
		t.Errorf("expected key 'balanced', got '%s'", p.Key)
	}
}

func TestLookup_TagMatch(t *testing.T) {
	p, found := Lookup("recommended")
	if !found {
		t.Fatal("expected to find preset with tag 'recommended'")
	}
	if p.Key != "balanced" {
		t.Errorf("expected balanced preset for tag 'recommended', got '%s'", p.Key)
	}
}

func TestLookup_FuzzyNameMatch(t *testing.T) {
	p, found := Lookup("ultra")
	if !found {
		t.Fatal("expected to find preset matching 'ultra'")
	}
	if p.Key != "ultra" {
		t.Errorf("expected ultra preset, got '%s'", p.Key)
	}
}

func TestLookup_FuzzyDescriptionMatch(t *testing.T) {
	p, found := Lookup("archival")
	if !found {
		t.Fatal("expected to find preset matching 'archival' in description")
	}
	if p.Key != "ultra" {
		t.Errorf("expected ultra preset for 'archival', got '%s'", p.Key)
	}
}

func TestLookup_NotFound(t *testing.T) {
	_, found := Lookup("nonexistent-preset-xyz")
	if found {
		t.Error("expected no match for nonexistent preset")
	}
}

func TestLookup_EmptyQuery(t *testing.T) {
	_, found := Lookup("")
	if found {
		t.Error("expected no match for empty query")
	}
}

func TestByCategory_Quality(t *testing.T) {
	quality := ByCategory(CategoryQuality)
	if len(quality) != 6 {
		t.Errorf("expected 6 quality presets, got %d", len(quality))
	}
	for _, p := range quality {
		if p.Category != CategoryQuality {
			t.Errorf("preset %s has category %s, expected quality", p.Key, p.Category)
		}
	}
}

func TestByCategory_Purpose(t *testing.T) {
	purpose := ByCategory(CategoryPurpose)
	if len(purpose) != 5 {
		t.Errorf("expected 5 purpose presets, got %d", len(purpose))
	}
}

func TestPreset_RequiredFields(t *testing.T) {
	for _, p := range All() {
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
			if p.Category == "" {
				t.Error("preset Category is empty")
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
		})
	}
}

func TestPreset_LosslessCRF(t *testing.T) {
	p, found := Lookup("lossless")
	if !found {
		t.Fatal("lossless preset not found")
	}
	if p.CRF != 0 {
		t.Errorf("lossless CRF expected 0, got %d", p.CRF)
	}
	if p.AudioCodec != "copy" {
		t.Errorf("lossless audio expected copy, got %s", p.AudioCodec)
	}
}

func TestPreset_TinyResolution(t *testing.T) {
	p, found := Lookup("tiny")
	if !found {
		t.Fatal("tiny preset not found")
	}
	if p.Resolution == "" {
		t.Error("tiny preset should have a resolution cap")
	}
}

func TestRegister_CustomPreset(t *testing.T) {
	before := len(All())

	Register(Preset{
		Key:       "test-custom",
		Name:      "Test Custom",
		Category:  CategoryQuality,
		Codec:     "h264",
		Container: "mp4",
	})

	after := len(All())
	if after != before+1 {
		t.Errorf("expected %d presets after register, got %d", before+1, after)
	}

	p, found := Lookup("test-custom")
	if !found {
		t.Error("custom preset not found after register")
	}
	if p.Key != "test-custom" {
		t.Errorf("expected key 'test-custom', got '%s'", p.Key)
	}

	// Clean up: remove the test preset from registry
	var cleaned []Preset
	for _, p := range registry {
		if p.Key != "test-custom" {
			cleaned = append(cleaned, p)
		}
	}
	registry = cleaned
}
