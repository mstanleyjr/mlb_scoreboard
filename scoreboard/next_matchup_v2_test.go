package scoreboard

import (
	"image/color"
	"testing"
	"time"
)

func TestSetNextMatchupTimingDefaults(t *testing.T) {
	SetNextMatchupTiming(0, 0)
	if nextMatchupHoldDuration != 2500*time.Millisecond || nextMatchupSlideDuration != 500*time.Millisecond {
		t.Fatalf("expected default matchup timing, got hold=%s slide=%s", nextMatchupHoldDuration, nextMatchupSlideDuration)
	}
}

func TestNextMatchupPitcherDetailLine(t *testing.T) {
	if got := nextMatchupPitcherDetailLine(ScoreboardPitcher{Hand: "R", ERA: "3.12"}); got != "RHP 3.12" {
		t.Fatalf("expected pitcher detail line, got %q", got)
	}
	if got := nextMatchupPitcherDetailLine(ScoreboardPitcher{}); got != "TBD" {
		t.Fatalf("expected fallback pitcher detail line, got %q", got)
	}
}

func TestFont5x8IncludesColon(t *testing.T) {
	if _, ok := font5x8[':']; !ok {
		t.Fatal("expected 5x8 font to include colon glyph")
	}
	if _, ok := font5x8['@']; !ok {
		t.Fatal("expected 5x8 font to include at-sign glyph")
	}
}

func TestMeasureText5x8SegmentsWidthPreservesInteriorSpaces(t *testing.T) {
	segments := []text5x8Segment{
		{text: "BAL", col: color.RGBA{R: 255, A: 255}},
		{text: " @ ", col: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{text: "WSH", col: color.RGBA{B: 255, A: 255}},
	}
	if got, want := measureText5x8SegmentsWidth(segments), measureText5x8Width("BAL @ WSH"); got != want {
		t.Fatalf("expected segmented width %d to match full-text width %d", got, want)
	}
}

func TestDrawNextMatchupAwayPanel(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardNextMatchup{
		AwayTeam: ScoreboardNextMatchupTeam{
			Name:      "New York Mets",
			ShortName: "NYM",
			ProbablePitcher: ScoreboardPitcher{
				FullName: "Tylor Megill",
				LastName: "Megill",
				Hand:     "R",
				ERA:      "3.12",
			},
			Record: ScoreboardWinLossRecord{Wins: 22, Losses: 19},
		},
	}

	DrawNextMatchupFrame(canvas, ScoreboardNextMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     1,
		NextPanelIndex: -1,
	})

	awayHeader := color.RGBA{R: 120, G: 180, B: 255, A: 255}
	awayTeam := GetTeamColor(matchup.AwayTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 160, G: 160, B: 160, A: 255}

	if countColorInBand(canvas, awayHeader, 1, 8) == 0 {
		t.Fatalf("expected away header pixels to render")
	}
	if countColorInBand(canvas, awayTeam, 14, 24) == 0 {
		t.Fatalf("expected away abbreviation to render in team color")
	}
	if countColorInBand(canvas, white, 14, 34) == 0 {
		t.Fatalf("expected away team record and pitcher name to render")
	}
	if countColorInBand(canvas, grey, 40, 48) == 0 {
		t.Fatalf("expected away pitcher detail line to render")
	}
}

func TestDrawNextMatchupOverviewTwoLineVenueSpacing(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	matchup := ScoreboardNextMatchup{
		AwayTeam: ScoreboardNextMatchupTeam{Name: "Baltimore Orioles", ShortName: "BAL"},
		HomeTeam: ScoreboardNextMatchupTeam{Name: "Washington Nationals", ShortName: "WSH"},
		DateTime: time.Date(2026, 5, 17, 19, 5, 0, 0, time.Local),
		Venue:    "Nationals Park",
	}

	DrawNextMatchupFrame(canvas, ScoreboardNextMatchupFrame{
		Matchup:        matchup,
		PanelIndex:     0,
		NextPanelIndex: -1,
	})

	awayTeam := GetTeamColor(matchup.AwayTeam.Name)
	homeTeam := GetTeamColor(matchup.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 160, G: 160, B: 160, A: 255}
	if countColorInBand(canvas, awayTeam, 35, 42) == 0 {
		t.Fatalf("expected away abbreviation in matchup line to render in team color")
	}
	if countColorInBand(canvas, homeTeam, 35, 42) == 0 {
		t.Fatalf("expected home abbreviation in matchup line to render in team color")
	}
	if countColorInBand(canvas, white, 35, 42) == 0 {
		t.Fatalf("expected separator in matchup line to render in its new band")
	}
	if countColorInBand(canvas, grey, 47, 63) == 0 {
		t.Fatalf("expected two-line venue footer to render below matchup line")
	}
}

func TestDrawNextMatchupUsesBlackBackground(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	DrawNextMatchupFrame(canvas, ScoreboardNextMatchupFrame{
		Matchup:        ScoreboardNextMatchup{},
		PanelIndex:     0,
		NextPanelIndex: -1,
	})

	background := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if got := canvas.pix[63*canvas.w+63]; got != background {
		t.Fatalf("expected blank background pixel to be %v, got %v", background, got)
	}
}

func TestEasedSlideOffset(t *testing.T) {
	if got := easedSlideOffset(1, 10, 64); got <= 0 || got >= 7 {
		t.Fatalf("expected eased first step to be small but positive, got %d", got)
	}
	if got := easedSlideOffset(5, 10, 64); got != 32 {
		t.Fatalf("expected midpoint eased offset to be half distance, got %d", got)
	}
	if got := easedSlideOffset(10, 10, 64); got != 64 {
		t.Fatalf("expected final eased offset to reach full distance, got %d", got)
	}
}

func countColorInBand(canvas *MockCanvas, want color.RGBA, yStart, yEnd int) int {
	count := 0
	for y := yStart; y <= yEnd && y < canvas.h; y++ {
		if y < 0 {
			continue
		}
		for x := 0; x < canvas.w; x++ {
			if canvas.pix[y*canvas.w+x] == want {
				count++
			}
		}
	}
	return count
}
