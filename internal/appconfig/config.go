package appconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const DefaultPath = "mlb_scoreboard.json"

type Config struct {
	DivisionStandings DivisionStandingsConfig `json:"division_standings"`
	Termsim           TermsimConfig           `json:"termsim"`
}

type DivisionStandingsConfig struct {
	Monochrome      bool `json:"monochrome"`
	GreenBackground bool `json:"green_background"`
}

type TermsimConfig struct {
	Empty    string `json:"empty"`
	FPS      int    `json:"fps"`
	ShowGrid bool   `json:"show_grid"`
	RulerY   int    `json:"ruler_y"`
}

func DefaultConfig() Config {
	return Config{
		DivisionStandings: DivisionStandingsConfig{},
		Termsim: TermsimConfig{
			Empty:  " ",
			FPS:    3,
			RulerY: -1,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}
