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
		"hardware": {
			"brightness": 80
		},
		"live_game": {
			"hold_ms": 1800,
			"last_play_panel_v2": true
		},
		"next_matchup": {
			"hold_ms": 3000
		},
		"last_matchup": {
			"slide_ms": 700
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
	if cfg.Hardware.Brightness != 80 {
		t.Fatalf("expected hardware brightness 80, got %+v", cfg.Hardware)
	}
	if cfg.LiveGame.HoldMS != 1800 || cfg.LiveGame.SlideMS != 400 || !cfg.LiveGame.LastPlayPanelV2 {
		t.Fatalf("expected live game config to merge with defaults, got %+v", cfg.LiveGame)
	}
	if cfg.NextMatchup.HoldMS != 3000 || cfg.NextMatchup.SlideMS != 500 {
		t.Fatalf("expected next matchup config to merge with defaults, got %+v", cfg.NextMatchup)
	}
	if cfg.LastMatchup.HoldMS != 2500 || cfg.LastMatchup.SlideMS != 700 {
		t.Fatalf("expected last matchup config to merge with defaults, got %+v", cfg.LastMatchup)
	}
	if cfg.Termsim.FPS != 10 {
		t.Fatalf("expected termsim fps 10, got %d", cfg.Termsim.FPS)
	}
	if cfg.Termsim.Empty != " " || cfg.Termsim.RulerY != -1 {
		t.Fatalf("expected defaults to remain for unspecified termsim fields, got %+v", cfg.Termsim)
	}
}
