package scoreboard

import (
	"testing"
	"time"
)

func TestDivisionStandingsMaxScrollOffset(t *testing.T) {
	compact := ScoreboardDivision{
		Sections: []ScoreboardStandingsSection{
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}}},
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
			{Title: "AL EAST", Teams: []ScoreboardDivisionTeam{{Rank: 1}, {Rank: 2}}},
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
