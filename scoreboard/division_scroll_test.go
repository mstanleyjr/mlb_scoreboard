package scoreboard

import (
	"testing"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

func TestDivisionStandingsMaxScrollOffset(t *testing.T) {
	compact := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}}},
		},
	}
	if got := divisionStandingsMaxScrollOffset(compact); got != 0 {
		t.Fatalf("expected no scroll for compact rollup, got %d", got)
	}

	tall := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
			{Title: "AL CENTRAL", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
		},
	}
	if got := divisionStandingsMaxScrollOffset(tall); got <= 0 {
		t.Fatalf("expected positive scroll for tall rollup, got %d", got)
	}
}

func TestDivisionStandingsScrollOffset(t *testing.T) {
	displayInfo := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
			{Title: "AL CENTRAL", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
		},
	}
	totalFrames := 24
	maxOffset := divisionStandingsMaxScrollOffset(displayInfo)

	if got := divisionStandingsScrollOffset(displayInfo, 0, totalFrames); got != 0 {
		t.Fatalf("expected initial offset 0, got %d", got)
	}

	last := 0
	for frame := 0; frame < totalFrames; frame++ {
		got := divisionStandingsScrollOffset(displayInfo, frame, totalFrames)
		if got < last {
			t.Fatalf("expected monotonic offsets, frame %d had %d after %d", frame, got, last)
		}
		last = got
	}

	if got := divisionStandingsScrollOffset(displayInfo, totalFrames-1, totalFrames); got != maxOffset {
		t.Fatalf("expected final offset %d, got %d", maxOffset, got)
	}
}

func TestStandingsRollupDuration(t *testing.T) {
	compact := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}}},
		},
	}
	if got := standingsRollupDuration(compact, 0); got != 12*time.Second {
		t.Fatalf("expected compact standings duration 12s, got %s", got)
	}

	tall := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
			{Title: "AL CENTRAL", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}, {Rank: 3}, {Rank: 4}, {Rank: 5}}},
		},
	}
	if got := standingsRollupDuration(tall, 0); got <= 12*time.Second {
		t.Fatalf("expected tall standings duration above base duration, got %s", got)
	}
}

func TestStandingsTeamsForDivisionIncludesWildCardGamesBack(t *testing.T) {
	divisionID := int32(200)
	teamName := "Washington Nationals"
	teamShort := "Nationals"
	divisionGB := "2.5"
	wildCardGB := "1.0"
	divisionRank := "3"
	wins := int32(40)
	losses := int32(35)

	teams := standingsTeamsForDivision(statsapi.StandingsRestObject{
		Records: &[]statsapi.TeamStandingsRecordContainerRestObject{
			{
				Division: &statsapi.DivisionRestObject{Id: &divisionID},
				TeamRecords: &[]statsapi.TeamStandingsRecordRestObject{
					{
						DivisionRank:      &divisionRank,
						DivisionGamesBack: &divisionGB,
						WildCardGamesBack: &wildCardGB,
						Wins:              &wins,
						Losses:            &losses,
						Team: &statsapi.BaseballTeamRestObject{
							Name:     &teamName,
							TeamName: &teamShort,
						},
					},
				},
			},
		},
	}, statsapi.DivisionRestObject{Id: &divisionID})

	if len(teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(teams))
	}
	if teams[0].WildCardGamesBack != wildCardGB {
		t.Fatalf("expected wildcard games back %q, got %q", wildCardGB, teams[0].WildCardGamesBack)
	}
	if teams[0].GamesBack != divisionGB {
		t.Fatalf("expected division games back %q, got %q", divisionGB, teams[0].GamesBack)
	}
}
