package scoreboard

import (
	"testing"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

func ptr(s string) *string { return &s }

func assertNoPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", name, r)
		}
	}()
	fn()
}

func TestScorekeepingLastPlayNotation(t *testing.T) {
	ptr := func(s string) *string { return &s }
	pos := func(code string) *statsapi.BaseballPosition {
		return &statsapi.BaseballPosition{Code: ptr(code), Abbreviation: ptr(code)}
	}

	tests := []struct {
		name string
		play *statsapi.BaseballPlayRestObject
		want string
	}{
		{
			name: "nil play",
			play: nil,
			want: "",
		},
		{
			name: "direct enum home run",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeHomeRun)),
					Event:     ptr("Home Run"),
					Rbi:       int32Ptr(2),
				},
			},
			want: "HR, 2 RBI",
		},
		{
			name: "direct enum strikeout",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeStrikeout)),
					Event:     ptr("Strikeout"),
				},
			},
			want: "K",
		},
		{
			name: "walk with rbi gets suffix",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeWalk)),
					Event:     ptr("Walk"),
					Rbi:       int32Ptr(1),
				},
			},
			want: "BB, 1 RBI",
		},
		{
			name: "groundout with fielding credits becomes G4-3",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeFieldOut)),
					Event:     ptr("Groundout"),
					Rbi:       int32Ptr(1),
				},
				Runners: &[]statsapi.RunnerMovement{
					{
						Credits: &[]statsapi.PlayCreditRestObject{
							{Credit: ptr("f_assist"), Position: pos("4")},
							{Credit: ptr("f_putout"), Position: pos("3")},
						},
					},
				},
			},
			want: "G4-3, 1 RBI",
		},
		{
			name: "flyout with putout credit becomes F8",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeFieldOut)),
					Event:     ptr("Flyout"),
				},
				Credits: &[]statsapi.PlayCreditRestObject{
					{Credit: ptr("f_putout"), Position: pos("8")},
				},
			},
			want: "F8",
		},
		{
			name: "sac fly carries rbi suffix",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeSacFly)),
					Event:     ptr("Sac Fly"),
					Rbi:       int32Ptr(1),
				},
			},
			want: "SF, 1 RBI",
		},
		{
			name: "lineout with putout credit becomes L5",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeFieldOut)),
					Event:     ptr("Lineout"),
				},
				Credits: &[]statsapi.PlayCreditRestObject{
					{Credit: ptr("f_putout"), Position: pos("5")},
				},
			},
			want: "L5",
		},
		{
			name: "popout with putout credit becomes P2",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeFieldOut)),
					Event:     ptr("Popout"),
				},
				Credits: &[]statsapi.PlayCreditRestObject{
					{Credit: ptr("f_putout"), Position: pos("2")},
				},
			},
			want: "P2",
		},
		{
			name: "force out with sequence becomes FO4-6",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType: ptr(string(statsapi.EventTypeForceOut)),
					Event:     ptr("Force Out"),
				},
				Credits: &[]statsapi.PlayCreditRestObject{
					{Credit: ptr("f_assist"), Position: pos("4")},
					{Credit: ptr("f_putout"), Position: pos("6")},
				},
			},
			want: "FO4-6",
		},
		{
			name: "description fallback when no notation is recognized",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType:   ptr(string(statsapi.EventTypeFieldOut)),
					Event:       ptr("Something Unexpected"),
					Description: ptr("Something Unexpected"),
				},
			},
			want: "Something Unexpected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorekeepingLastPlayNotation(tt.play)
			if got != tt.want {
				t.Fatalf("scorekeepingLastPlayNotation() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetScoreboardLiveGameBatter_NilSafe(t *testing.T) {
	assertNoPanic(t, "empty game", func() {
		got := getScoreboardLiveGameBatter(statsapi.PlayerStatsResponse{}, statsapi.BaseballGameRestObject{}, ScoreboardLiveGame{})
		if got != (ScoreboardLiveGameBatter{}) {
			t.Fatalf("expected zero value batter, got %+v", got)
		}
	})

	assertNoPanic(t, "partial current play", func() {
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Batter: &statsapi.BaseballPersonRestObject{
								FullName: ptr("Edgar Quero"),
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGameBatter(statsapi.PlayerStatsResponse{}, game, ScoreboardLiveGame{HalfInning: "top", CurrentBatterId: 700337})
		if got.FullName != "Edgar Quero" || got.LastName != "" {
			t.Fatalf("expected full name only with empty last name, got %+v", got)
		}
	})
}

func TestGetScoreboardLiveGamePitcher_NilSafe(t *testing.T) {
	assertNoPanic(t, "empty game", func() {
		got := getScoreboardLiveGamePitcher(statsapi.PlayerStatsResponse{}, statsapi.BaseballGameRestObject{})
		if got.FullName != "" || got.LastName != "" || got.Hand != "" || got.ERA != "0.00" || got.Wins != 0 || got.Losses != 0 || got.Saves != 0 {
			t.Fatalf("expected default pitcher values, got %+v", got)
		}
	})

	assertNoPanic(t, "partial current play", func() {
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Pitcher: &statsapi.BaseballPersonRestObject{
								FullName: ptr("Brad Lord"),
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGamePitcher(statsapi.PlayerStatsResponse{}, game)
		if got.FullName != "Brad Lord" || got.LastName != "" {
			t.Fatalf("expected full name only with empty last name, got %+v", got)
		}
	})
}

func int32Ptr(v int32) *int32 { return &v }
