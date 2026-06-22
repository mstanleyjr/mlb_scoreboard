package scoreboard

import (
	"image/color"
	"testing"
)

func findColorBoundsInRows(canvas *MockCanvas, want color.RGBA, startY, endY int) (int, int, bool) {
	minX := canvas.w
	maxX := -1
	found := false
	for y := startY; y <= endY; y++ {
		if y < 0 || y >= canvas.h {
			continue
		}
		for x := 0; x < canvas.w; x++ {
			if canvas.pix[y*canvas.w+x] != want {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			found = true
		}
	}
	return minX, maxX, found
}

func TestTrimText5x8ToWidthPreservesWhiteSox(t *testing.T) {
	const teamAreaWidth = 64 - divisionStandingsTeamX - 1

	if got := trimText5x8ToWidth("WHITE SOX", teamAreaWidth); got != "WHITE SOX" {
		t.Fatalf("expected WHITE SOX to fit in %dpx, got %q", teamAreaWidth, got)
	}
}

func TestFormatDivisionGamesBackPreservesHalfGames(t *testing.T) {
	if got := trimToChars(formatDivisionGamesBack("11.5"), divisionStandingsMaxGBChars); got != "11.5" {
		t.Fatalf("expected 11.5 games back to be preserved, got %q", got)
	}
}

func TestFormatDivisionGamesBackPreservesWholeGames(t *testing.T) {
	if got := formatDivisionGamesBack("5.0"); got != "5.0" {
		t.Fatalf("expected 5.0 games back to be preserved, got %q", got)
	}
}

func TestDivisionStandingsColorsMonochrome(t *testing.T) {
	SetDivisionStandingsMonochrome(true)
	SetDivisionStandingsGreenBackground(false)
	t.Cleanup(func() {
		SetDivisionStandingsMonochrome(false)
		SetDivisionStandingsGreenBackground(false)
	})

	palette := divisionStandingsColors()
	want := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	if palette.title != want || palette.team != want || palette.record != want || palette.gamesBack != want || palette.rule != want || palette.line != want {
		t.Fatalf("expected monochrome palette to use %+v across standings colors, got %+v", want, palette)
	}
}

func TestDivisionStandingsColorsGreenBackground(t *testing.T) {
	SetDivisionStandingsMonochrome(false)
	SetDivisionStandingsGreenBackground(true)
	t.Cleanup(func() {
		SetDivisionStandingsMonochrome(false)
		SetDivisionStandingsGreenBackground(false)
	})

	palette := divisionStandingsColors()
	want := color.RGBA{R: 22, G: 67, B: 22, A: 255}
	if palette.background != want {
		t.Fatalf("expected green background %+v, got %+v", want, palette.background)
	}
}

func TestTrimText5x8ToWidthPreservesLongSingleWordNames(t *testing.T) {
	const teamAreaWidth = 64 - divisionStandingsTeamX - 1

	for _, name := range []string{"NATIONALS", "CARDINALS"} {
		if got := trimText5x8ToWidth(name, teamAreaWidth); got != name {
			t.Fatalf("expected %s to fit in %dpx, got %q", name, teamAreaWidth, got)
		}
	}
}

func TestDrawDivisionStandingsAddsHeaderSeparator(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	SetDivisionStandingsMonochrome(false)
	SetDivisionStandingsGreenBackground(false)
	division := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{
				Title: "AL CENTRAL",
				Teams: []ScoreboardDivisionTeam{
					{
						Name:      "Chicago White Sox",
						ShortName: "White Sox",
						Rank:      1,
						Record:    ScoreboardWinLossRecord{Wins: 10, Losses: 5},
						GamesBack: "0.0",
					},
				},
			},
		},
	}

	DrawDivisionStandings(canvas, division)

	ruleY := divisionStandingsTopPadding + divisionStandingsTitleH + divisionStandingsTitleRuleGap
	want := color.RGBA{R: 70, G: 70, B: 70, A: 255}
	got := canvas.pix[ruleY*canvas.w+1]
	if got != want {
		t.Fatalf("expected divider pixel at (1,%d) to be %+v, got %+v", ruleY, want, got)
	}
	if got := canvas.pix[(ruleY+1)*canvas.w+divisionStandingsLineX]; got != (color.RGBA{R: 90, G: 90, B: 90, A: 255}) {
		t.Fatalf("expected first vertical line pixel below header at (%d,%d), got %+v", divisionStandingsLineX, ruleY+1, got)
	}
}

func TestDrawDivisionStandingsMonochromeUsesLightGray(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	SetDivisionStandingsMonochrome(true)
	SetDivisionStandingsGreenBackground(false)
	t.Cleanup(func() {
		SetDivisionStandingsMonochrome(false)
		SetDivisionStandingsGreenBackground(false)
	})

	division := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{
				Title: "AL CENTRAL",
				Teams: []ScoreboardDivisionTeam{
					{
						Name:      "Chicago White Sox",
						ShortName: "White Sox",
						Rank:      1,
						Record:    ScoreboardWinLossRecord{Wins: 10, Losses: 5},
						GamesBack: "0.0",
					},
				},
			},
		},
	}

	DrawDivisionStandings(canvas, division)

	want := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	ruleY := divisionStandingsTopPadding + divisionStandingsTitleH + divisionStandingsTitleRuleGap
	if got := canvas.pix[ruleY*canvas.w+1]; got != want {
		t.Fatalf("expected divider pixel to be %+v, got %+v", want, got)
	}
	if got := canvas.pix[(ruleY+1)*canvas.w+divisionStandingsLineX]; got != want {
		t.Fatalf("expected first vertical line pixel below header to be %+v, got %+v", want, got)
	}
	nameY := divisionStandingsTopPadding + divisionStandingsTitleH + divisionStandingsTitleGap
	if got := canvas.pix[nameY*canvas.w+divisionStandingsTeamX]; got != want {
		t.Fatalf("expected team pixel to be %+v, got %+v", want, got)
	}
}

func TestDrawDivisionStandingsGreenBackgroundUsesConfiguredColor(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	SetDivisionStandingsMonochrome(false)
	SetDivisionStandingsGreenBackground(true)
	t.Cleanup(func() {
		SetDivisionStandingsMonochrome(false)
		SetDivisionStandingsGreenBackground(false)
	})

	DrawDivisionStandings(canvas, ScoreboardDivision{})

	want := color.RGBA{R: 22, G: 67, B: 22, A: 255}
	if got := canvas.pix[0]; got != want {
		t.Fatalf("expected background pixel to be %+v, got %+v", want, got)
	}
}

func TestDrawDivisionStandingsSecondRowGetsSameTopSpacing(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	SetDivisionStandingsMonochrome(false)
	SetDivisionStandingsGreenBackground(false)
	division := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{
				Title: "AL CENTRAL",
				Teams: []ScoreboardDivisionTeam{
					{
						Name:      "Chicago White Sox",
						ShortName: "White Sox",
						Rank:      1,
						Record:    ScoreboardWinLossRecord{Wins: 10, Losses: 5},
						GamesBack: "0.0",
					},
					{
						Name:      "Detroit Tigers",
						ShortName: "Tigers",
						Rank:      2,
						Record:    ScoreboardWinLossRecord{Wins: 9, Losses: 6},
						GamesBack: "1.0",
					},
				},
			},
		},
	}

	DrawDivisionStandings(canvas, division)

	secondNameY := divisionStandingsTopPadding + divisionStandingsTitleH + divisionStandingsTitleGap + divisionStandingsRowBlockH + divisionStandingsRowGap
	lineY := secondNameY - divisionStandingsRowGap
	want := color.RGBA{R: 90, G: 90, B: 90, A: 255}
	if got := canvas.pix[lineY*canvas.w+divisionStandingsLineX]; got != want {
		t.Fatalf("expected second vertical line pixel at (%d,%d) to be %+v, got %+v", divisionStandingsLineX, lineY, want, got)
	}
	if got := canvas.pix[secondNameY*canvas.w+divisionStandingsTeamX]; got == (color.RGBA{}) {
		t.Fatalf("expected second team row to render at y=%d", secondNameY)
	}
}

func TestDrawDivisionStandingsSplitsGBLabelsAndValues(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
	SetDivisionStandingsMonochrome(false)
	SetDivisionStandingsGreenBackground(false)
	division := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{
				Title: "AL EAST",
				Teams: []ScoreboardDivisionTeam{
					{
						Name:              "Boston Red Sox",
						ShortName:         "Red Sox",
						Rank:              4,
						Record:            ScoreboardWinLossRecord{Wins: 10, Losses: 5},
						GamesBack:         "28.0",
						WildCardGamesBack: "+1.5",
					},
				},
			},
		},
	}

	DrawDivisionStandings(canvas, division)

	nameY := divisionStandingsTopPadding + divisionStandingsTitleH + divisionStandingsTitleGap
	recordY := nameY + 10
	gbY := nameY + 20
	wildCardGBY := nameY + 30
	palette := divisionStandingsColors()

	recordMinX, _, found := findColorBoundsInRows(canvas, palette.record, recordY, recordY+fontH5x8-1)
	if !found {
		t.Fatalf("expected left-aligned record row to render at y=%d", recordY)
	}
	if recordMinX != divisionStandingsRecordX {
		t.Fatalf("expected record row to start at x=%d, got %d", divisionStandingsRecordX, recordMinX)
	}

	if _, _, found := findColorBoundsInRows(canvas, palette.gamesBack, recordY, recordY+fontH5x8-1); found {
		t.Fatalf("expected GB/WCGB text on their own rows, but found games-back pixels on record row y=%d", recordY)
	}
	gbMinX, gbMaxX, found := findColorBoundsInRows(canvas, palette.gamesBack, gbY, gbY+fontH5x8-1)
	if !found {
		t.Fatalf("expected GB row to render at y=%d", gbY)
	}
	if gbMinX != divisionStandingsRecordX {
		t.Fatalf("expected GB row to start at x=%d, got %d", divisionStandingsRecordX, gbMinX)
	}
	if gbMaxX < divisionStandingsGBRightX-6 {
		t.Fatalf("expected GB value to be right-aligned near x=%d, got max x %d", divisionStandingsGBRightX, gbMaxX)
	}
	wildCardGBMinX, wildCardGBMaxX, found := findColorBoundsInRows(canvas, palette.gamesBack, wildCardGBY, wildCardGBY+fontH5x8-1)
	if !found {
		t.Fatalf("expected WCGB row to render at y=%d", wildCardGBY)
	}
	if wildCardGBMinX != divisionStandingsRecordX {
		t.Fatalf("expected WCGB row to start at x=%d, got %d", divisionStandingsRecordX, wildCardGBMinX)
	}
	if wildCardGBMaxX < divisionStandingsGBRightX-6 {
		t.Fatalf("expected WCGB value to be right-aligned near x=%d, got max x %d", divisionStandingsGBRightX, wildCardGBMaxX)
	}
}
