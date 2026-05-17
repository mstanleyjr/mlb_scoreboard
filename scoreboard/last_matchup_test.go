package scoreboard

import (
	"image/color"
	"testing"
	"time"
)

func TestSetLastMatchupTimingDefaults(t *testing.T) {
	SetLastMatchupTiming(0, 0)
	if lastMatchupHoldDuration != 2500*time.Millisecond || lastMatchupSlideDuration != 500*time.Millisecond {
		t.Fatalf("expected default last matchup timing, got hold=%s slide=%s", lastMatchupHoldDuration, lastMatchupSlideDuration)
	}
}

func TestDrawLastMatchupFrameStaticTopAndRecordsPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardLastMatchup{
		AwayTeam: ScoreboardLastMatchupTeam{
			Team: ScoreboardLiveGameTeam{
				Name:      "Baltimore Orioles",
				ShortName: "BAL",
				Record:    ScoreboardWinLossRecord{Wins: 20, Losses: 25},
				Runs:      3,
			},
		},
		HomeTeam: ScoreboardLastMatchupTeam{
			Team: ScoreboardLiveGameTeam{
				Name:      "Washington Nationals",
				ShortName: "WSH",
				Record:    ScoreboardWinLossRecord{Wins: 23, Losses: 23},
				Runs:      13,
			},
			Winner: true,
		},
		FinalInning: 9,
		DateTime:    time.Date(2026, 5, 17, 19, 5, 0, 0, time.Local),
		Venue:       "Nationals Park",
	}

	DrawLastMatchupFrame(canvas, ScoreboardLastMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     0,
		NextPanelIndex: -1,
	})

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, white, 3, 10) == 0 {
		t.Fatalf("expected static matchup line to render")
	}
	if countColorInBand(canvas, white, 14, 21) == 0 {
		t.Fatalf("expected static score line to render")
	}
	if countColorInBand(canvas, yellow, 14, 21) == 0 {
		t.Fatalf("expected final status tag to render next to the score")
	}
	if countColorInBand(canvas, white, 28, 35) == 0 {
		t.Fatalf("expected records to render in records panel")
	}
	if countColorInBand(canvas, white, 37, 44) == 0 {
		t.Fatalf("expected date to render on first slide")
	}
	if countColorInBand(canvas, grey, 46, 62) == 0 {
		t.Fatalf("expected venue to render on first slide")
	}
}

func TestDrawLastMatchupFrameRHELPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardLastMatchup{
		AwayTeam: ScoreboardLastMatchupTeam{
			Team: ScoreboardLiveGameTeam{
				Name:      "Baltimore Orioles",
				ShortName: "BAL",
				Runs:      3,
				Hits:      8,
				Errors:    1,
				LOB:       6,
			},
		},
		HomeTeam: ScoreboardLastMatchupTeam{
			Team: ScoreboardLiveGameTeam{
				Name:      "Washington Nationals",
				ShortName: "WSH",
				Runs:      13,
				Hits:      15,
				Errors:    0,
				LOB:       7,
			},
		},
	}

	DrawLastMatchupFrame(canvas, ScoreboardLastMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     1,
		NextPanelIndex: -1,
	})

	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	awayTeam := GetTeamColor(matchup.AwayTeam.Team.Name)
	homeTeam := GetTeamColor(matchup.HomeTeam.Team.Name)
	if countColorInBand(canvas, grey, 28, 62) == 0 {
		t.Fatalf("expected stat labels to render")
	}
	if countColorInBand(canvas, awayTeam, 28, 62) == 0 {
		t.Fatalf("expected away values to render in away color")
	}
	if countColorInBand(canvas, homeTeam, 28, 62) == 0 {
		t.Fatalf("expected home values to render in home color")
	}
}

func TestDrawLastMatchupFramePitchingPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardLastMatchup{
		AwayTeam:               ScoreboardLastMatchupTeam{Team: ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"}},
		HomeTeam:               ScoreboardLastMatchupTeam{Team: ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"}},
		WinningPitcherLastName: "PARKER",
		LosingPitcherLastName:  "IRVIN",
		SavePitcherLastName:    "FINNEGAN",
	}

	DrawLastMatchupFrame(canvas, ScoreboardLastMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     2,
		NextPanelIndex: -1,
	})

	green := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	red := color.RGBA{R: 220, G: 60, B: 60, A: 255}
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	if countColorInBand(canvas, green, 28, 35) == 0 {
		t.Fatalf("expected winning pitcher line to render")
	}
	if countColorInBand(canvas, red, 37, 44) == 0 {
		t.Fatalf("expected losing pitcher line to render")
	}
	if countColorInBand(canvas, white, 46, 53) == 0 {
		t.Fatalf("expected save pitcher line to render")
	}
}
