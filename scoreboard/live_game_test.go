package scoreboard

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"
)

type strictCanvas struct {
	w, h int
	pix  []color.RGBA
}

func newStrictCanvas(w, h int) *strictCanvas {
	return &strictCanvas{w: w, h: h, pix: make([]color.RGBA, w*h)}
}

func (c *strictCanvas) Set(x, y int, col color.Color) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		panic(fmt.Sprintf("out of bounds Set(%d,%d)", x, y))
	}
	r, g, b, a := col.RGBA()
	c.pix[y*c.w+x] = color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func (c *strictCanvas) Bounds() image.Rectangle {
	return image.Rect(0, 0, c.w, c.h)
}

func mockPixel(canvas *MockCanvas, x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= canvas.w || y >= canvas.h {
		return color.RGBA{}
	}
	return canvas.pix[y*canvas.w+x]
}

func TestSetLiveGameTimingDefaults(t *testing.T) {
	SetLiveGameTiming(0, 0)
	if liveGameHoldDuration != 2200*time.Millisecond || liveGameSlideDuration != 400*time.Millisecond {
		t.Fatalf("expected default live game timing, got hold=%s slide=%s", liveGameHoldDuration, liveGameSlideDuration)
	}
}

func TestMeasureTightText5x8WidthShrinksNotationTracking(t *testing.T) {
	notation := "GDP6-4"
	if got, want := measureTightText5x8Width(notation), measureText5x8Width(notation); got >= want {
		t.Fatalf("expected tight tracking width %d to be smaller than normal %d", got, want)
	}
}

func TestMeasureLiveGameStatusTextWidthTightensExtraInningsSpacing(t *testing.T) {
	game := ScoreboardLiveGame{Inning: 10, HalfInning: "bottom", Balls: 3, Strikes: 2}
	if got, wantMax := measureLiveGameStatusTextWidth(game), measureText5x8Width("B10 3-2")-1; got > wantMax {
		t.Fatalf("expected tightened status width <= %d, got %d", wantMax, got)
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
		LastPlayNotation: "2B, 2 RBI",
		LastPlayRBIs:     2,
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     3,
		NextPanelIndex: -1,
	})

	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if countColorInBand(canvas, orange, 29, 63) == 0 {
		t.Fatalf("expected last play text to render")
	}
	if mockPixel(canvas, 40, 54) != white || mockPixel(canvas, 40, 36) != white {
		t.Fatalf("expected white offset advancement path for a double")
	}
	if mockPixel(canvas, 50, 45) != black {
		t.Fatalf("expected old outside runner marker to be removed")
	}
	if mockPixel(canvas, 55, 34) != orange || mockPixel(canvas, 55, 41) != orange {
		t.Fatalf("expected RBI dots on right side")
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

func TestDrawLiveGameFrameLastPlayPanelClipsRightEdge(t *testing.T) {
	canvas := newStrictCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:         ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:         ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		LastPlayNotation: "RBI single to center",
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     2,
		NextPanelIndex: 3,
		SlideOffset:    0,
	})
}

func TestDrawLiveGameFrameLastPlayPanelLookingStrikeoutUsesBackwardK(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:         ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:         ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		LastPlayNotation: "K",
		LastPlay:         "Called strike three looking.",
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     3,
		NextPanelIndex: -1,
	})

	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if mockPixel(canvas, 29, 44) != black {
		t.Fatalf("expected mirrored K to leave left row-1 pixel empty")
	}
	if mockPixel(canvas, 32, 44) != orange || mockPixel(canvas, 33, 44) != orange {
		t.Fatalf("expected mirrored K row to render on the right-shifted columns")
	}
}

func TestDrawLiveGameFrameLastPlayPanelSwingingStrikeoutKeepsNormalK(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	game := ScoreboardLiveGame{
		AwayTeam:         ScoreboardLiveGameTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam:         ScoreboardLiveGameTeam{Name: "Washington Nationals", ShortName: "WSH"},
		LastPlayNotation: "K",
		LastPlay:         "Strikeout swinging.",
	}

	DrawLiveGameFrame(canvas, ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     3,
		NextPanelIndex: -1,
	})

	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}
	if mockPixel(canvas, 29, 44) != orange || mockPixel(canvas, 30, 44) != orange {
		t.Fatalf("expected standard K shape for non-looking strikeout")
	}
}
