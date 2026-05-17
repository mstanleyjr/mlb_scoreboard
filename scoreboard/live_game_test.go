package scoreboard

import (
	"image/color"
	"testing"
)

func TestDrawLiveGameFrameStaticTopAndRHELPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:   ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL", Runs: 3, Hits: 8, Errors: 1, LOB: 6},
		HomeTeam:   ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH", Runs: 13, Hits: 15, Errors: 0, LOB: 7},
		Bases:      ScoreboardLiveGameBases{First: true, Second: false, Third: true},
		Inning:     7,
		HalfInning: "bottom",
		Outs:       2,
		Balls:      2,
		Strikes:    1,
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     0,
		NextPanelIndex: -1,
	})

	awayTeam := GetTeamColor(game.AwayTeam.Name)
	homeTeam := GetTeamColor(game.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	baseFill := color.RGBA{R: 200, G: 100, B: 0, A: 255}

	if countColorInBand(canvas, awayTeam, 0, 8) == 0 {
		t.Fatalf("expected away abbreviation to render in team color")
	}
	if countColorInBand(canvas, homeTeam, 0, 8) == 0 {
		t.Fatalf("expected home abbreviation to render in team color")
	}
	if countColorInBand(canvas, white, 9, 17) == 0 {
		t.Fatalf("expected score line to render")
	}
	if countColorInBand(canvas, grey, 18, 25) == 0 {
		t.Fatalf("expected inning/count status line to render")
	}
	if countColorInBand(canvas, baseFill, 8, 18) == 0 {
		t.Fatalf("expected occupied bases to render in static top")
	}
	if countColorInBand(canvas, grey, 28, 62) == 0 {
		t.Fatalf("expected RHEL labels to render")
	}
}

func TestDrawLiveGameFrameBatterPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam: ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam: ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		CurrentBatter: ScoreboardLiveGameBatter{
			FullName:             "CJ Abrams",
			LastName:             "Abrams",
			CurrentPosition:      "SS",
			SeasonBattingAverage: ".287",
			SeasonOPS:            ".812",
			Summary:              "2-3, HR",
		},
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     1,
		NextPanelIndex: -1,
	})

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, white, 28, 35) == 0 {
		t.Fatalf("expected batter name to render")
	}
	if countColorInBand(canvas, grey, 37, 53) == 0 {
		t.Fatalf("expected batter detail lines to render")
	}
	if countColorInBand(canvas, white, 55, 62) == 0 {
		t.Fatalf("expected batter summary to render")
	}
}

func TestDrawLiveGameFramePitcherPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam: ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam: ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		CurrentPitcher: ScoreboardPitcher{
			FullName: "Dylan Tate",
			LastName: "Tate",
			Hand:     "R",
			ERA:      "3.42",
			Wins:     2,
			Losses:   1,
		},
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     2,
		NextPanelIndex: -1,
	})

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, white, 28, 35) == 0 {
		t.Fatalf("expected pitcher name to render")
	}
	if countColorInBand(canvas, grey, 37, 53) == 0 {
		t.Fatalf("expected pitcher detail lines to render")
	}
	if countColorInBand(canvas, white, 55, 62) == 0 {
		t.Fatalf("expected pitcher tertiary line to render")
	}
}

func TestDrawLiveGameFrameLastPlayPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:         ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:         ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		LastPlayNotation: "RBI single to center",
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     3,
		NextPanelIndex: -1,
	})

	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	if countColorInBand(canvas, orange, 28, 62) == 0 {
		t.Fatalf("expected last play text to render")
	}
}
