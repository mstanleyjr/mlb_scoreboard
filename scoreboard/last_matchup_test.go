package scoreboard

import (
	"image/color"
	"testing"
	"time"
)

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
	}

	DrawLastMatchupFrame(canvas, ScoreboardLastMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     0,
		NextPanelIndex: -1,
	})

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	awayHeader := color.RGBA{R: 120, G: 180, B: 255, A: 255}
	homeHeader := color.RGBA{R: 120, G: 255, B: 140, A: 255}
	if countColorInBand(canvas, white, 8, 15) == 0 {
		t.Fatalf("expected static matchup line to render")
	}
	if countColorInBand(canvas, white, 20, 27) == 0 {
		t.Fatalf("expected static score line to render")
	}
	if countColorInBand(canvas, yellow, 20, 27) == 0 {
		t.Fatalf("expected final status tag to render next to the score")
	}
	if countColorInBand(canvas, awayHeader, 37, 44) == 0 {
		t.Fatalf("expected away label to render in records panel")
	}
	if countColorInBand(canvas, homeHeader, 37, 44) == 0 {
		t.Fatalf("expected home label to render in records panel")
	}
	if countColorInBand(canvas, white, 50, 58) == 0 {
		t.Fatalf("expected records to render in records panel")
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
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	awayTeam := GetTeamColor(matchup.AwayTeam.Team.Name)
	homeTeam := GetTeamColor(matchup.HomeTeam.Team.Name)
	if countColorInBand(canvas, grey, 37, 44) == 0 {
		t.Fatalf("expected RHEL header row to render")
	}
	if countColorInBand(canvas, awayTeam, 46, 53) == 0 {
		t.Fatalf("expected away abbreviation to render in team color")
	}
	if countColorInBand(canvas, homeTeam, 55, 62) == 0 {
		t.Fatalf("expected home abbreviation to render in team color")
	}
	if countColorInBand(canvas, white, 46, 62) == 0 {
		t.Fatalf("expected stat values to render in white")
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
	if countColorInBand(canvas, green, 37, 44) == 0 {
		t.Fatalf("expected winning pitcher line to render")
	}
	if countColorInBand(canvas, red, 37, 44) == 0 {
		t.Fatalf("expected losing pitcher line to render")
	}
	if countColorInBand(canvas, white, 50, 57) == 0 {
		t.Fatalf("expected save pitcher line to render")
	}
}

func TestDrawLastMatchupFrameContextPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardLastMatchup{
		AwayTeam: ScoreboardLastMatchupTeam{Team: ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL", Runs: 3}},
		HomeTeam: ScoreboardLastMatchupTeam{Team: ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH", Runs: 13}},
		DateTime: time.Date(2026, 5, 17, 19, 5, 0, 0, time.Local),
		Venue:    "Nationals Park",
	}

	DrawLastMatchupFrame(canvas, ScoreboardLastMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     3,
		NextPanelIndex: -1,
	})

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, white, 37, 44) == 0 {
		t.Fatalf("expected date line to render")
	}
	if countColorInBand(canvas, grey, 46, 62) == 0 {
		t.Fatalf("expected venue footer to render")
	}
}
