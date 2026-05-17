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
	LiveGame          LiveGameConfig          `json:"live_game"`
	NextMatchup       NextMatchupConfig       `json:"next_matchup"`
	LastMatchup       LastMatchupConfig       `json:"last_matchup"`
	Termsim           TermsimConfig           `json:"termsim"`
}

type DivisionStandingsConfig struct {
	Monochrome      bool `json:"monochrome"`
	GreenBackground bool `json:"green_background"`
}

type LiveGameConfig struct {
	HoldMS  int `json:"hold_ms"`
	SlideMS int `json:"slide_ms"`
}

type NextMatchupConfig struct {
	HoldMS  int `json:"hold_ms"`
	SlideMS int `json:"slide_ms"`
}

type LastMatchupConfig struct {
	HoldMS  int `json:"hold_ms"`
	SlideMS int `json:"slide_ms"`
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
		LiveGame: LiveGameConfig{
			HoldMS:  2200,
			SlideMS: 400,
		},
		NextMatchup: NextMatchupConfig{
			HoldMS:  2500,
			SlideMS: 500,
		},
		LastMatchup: LastMatchupConfig{
			HoldMS:  2500,
			SlideMS: 500,
		},
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
