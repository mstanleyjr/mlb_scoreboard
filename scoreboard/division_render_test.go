package scoreboard

import (
	"image/color"
	"testing"
)

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
