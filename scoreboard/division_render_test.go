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

func TestDrawDivisionStandingsAddsHeaderSeparator(t *testing.T) {
	canvas := NewMockCanvas(64, 64)
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
