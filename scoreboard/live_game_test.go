package scoreboard

import (
	"image/color"
	"testing"
	"time"
)

func TestSetLiveGameTimingDefaults(t *testing.T) {
	SetLiveGameTiming(0, 0)
	if liveGameHoldDuration != 2200*time.Millisecond || liveGameSlideDuration != 400*time.Millisecond {
		t.Fatalf("expected default live game timing, got hold=%s slide=%s", liveGameHoldDuration, liveGameSlideDuration)
	}
}

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

	if countColorInBand(canvas, awayTeam, 1, 9) == 0 {
		t.Fatalf("expected away abbreviation to render in team color")
	}
	if countColorInBand(canvas, homeTeam, 1, 9) == 0 {
		t.Fatalf("expected home abbreviation to render in team color")
	}
	if countColorInBand(canvas, white, 10, 18) == 0 {
		t.Fatalf("expected score line to render")
	}
	if countColorInBand(canvas, grey, 19, 26) == 0 {
		t.Fatalf("expected inning/count status line to render")
	}
	if countColorInBand(canvas, white, 19, 26) == 0 {
		t.Fatalf("expected filled out circles to render in status line")
	}
	if countColorInBand(canvas, baseFill, 9, 19) == 0 {
		t.Fatalf("expected occupied bases to render in static top")
	}
	if countColorInBand(canvas, grey, 29, 63) == 0 {
		t.Fatalf("expected RHEL labels to render")
	}
}

func TestDrawLiveGameFrameBatterPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:   ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:   ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		HalfInning: "bottom",
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

	batterTeam := GetTeamColor(game.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, batterTeam, 29, 36) == 0 {
		t.Fatalf("expected batter name to render in batting team color")
	}
	if countColorInBand(canvas, grey, 38, 54) == 0 {
		t.Fatalf("expected batter detail lines to render")
	}
	if countColorInBand(canvas, white, 56, 63) == 0 {
		t.Fatalf("expected batter summary to render")
	}
}

func TestDrawLiveGameFramePitcherPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:   ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:   ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		HalfInning: "bottom",
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

	pitcherTeam := GetTeamColor(game.AwayTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	if countColorInBand(canvas, pitcherTeam, 29, 36) == 0 {
		t.Fatalf("expected pitcher name to render in fielding team color")
	}
	if countColorInBand(canvas, grey, 38, 54) == 0 {
		t.Fatalf("expected pitcher detail lines to render")
	}
	if countColorInBand(canvas, white, 56, 63) == 0 {
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

	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	if countColorInBand(canvas, grey, 29, 36) == 0 {
		t.Fatalf("expected last play title to render")
	}
	if countColorInBand(canvas, orange, 38, 63) == 0 {
		t.Fatalf("expected last play text to render")
	}
}

func TestDrawLiveGameFrameLastPlayPanelDoesNotBleedWhenOffscreen(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:         ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:         ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		LastPlayNotation: "RBI single to center",
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     3,
		NextPanelIndex: 0,
		SlideOffset:    64,
	})

	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	if countColorInBand(canvas, orange, 29, 63) != 0 {
		t.Fatalf("expected offscreen last play panel to leave no visible orange pixels")
	}
}
