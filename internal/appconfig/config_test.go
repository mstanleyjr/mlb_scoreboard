package appconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingConfigUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected missing config to fall back to defaults, got error %v", err)
	}

	if cfg != DefaultConfig() {
		t.Fatalf("expected default config %+v, got %+v", DefaultConfig(), cfg)
	}
}

func TestLoadConfigMergesWithDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mlb_scoreboard.json")
	if err := os.WriteFile(path, []byte(`{
		"division_standings": {
			"monochrome": true,
			"green_background": true
		},
		"termsim": {
			"fps": 10
		}
	}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected config to load, got error %v", err)
	}

	if !cfg.DivisionStandings.Monochrome || !cfg.DivisionStandings.GreenBackground {
		t.Fatalf("expected division standings config to load, got %+v", cfg.DivisionStandings)
	}
	if cfg.Termsim.FPS != 10 {
		t.Fatalf("expected termsim fps 10, got %d", cfg.Termsim.FPS)
	}
	if cfg.Termsim.Empty != " " || cfg.Termsim.RulerY != -1 {
		t.Fatalf("expected defaults to remain for unspecified termsim fields, got %+v", cfg.Termsim)
	}
}
