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
			name: "grand slam becomes gs notation",
			play: &statsapi.BaseballPlayRestObject{
				Result: &statsapi.Result{
					EventType:   ptr(string(statsapi.EventTypeHomeRun)),
					Event:       ptr("Home Run"),
					Description: ptr("Batter hits a grand slam."),
					Rbi:         int32Ptr(4),
				},
			},
			want: "GS, 4 RBI",
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
			want: "",
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

func TestShortLastPlayDescriptionPitchingChange(t *testing.T) {
	ptr := func(s string) *string { return &s }
	play := &statsapi.BaseballPlayRestObject{
		Result: &statsapi.Result{
			EventType:   ptr(string(statsapi.EventTypePitchingSubstitution)),
			Event:       ptr("Pitching Substitution"),
			Description: ptr("Pitching Change: Joe Smith replaces John Doe."),
		},
	}

	if got := shortLastPlayDescription(play); got != "Pitching Change" {
		t.Fatalf("shortLastPlayDescription() = %q, want %q", got, "Pitching Change")
	}
}

func TestShortLastPlayDescriptionKeepsNormalDescription(t *testing.T) {
	ptr := func(s string) *string { return &s }
	play := &statsapi.BaseballPlayRestObject{
		Result: &statsapi.Result{
			EventType:   ptr(string(statsapi.EventTypeSingle)),
			Event:       ptr("Single"),
			Description: ptr("Line drive single to center."),
		},
	}

	if got := shortLastPlayDescription(play); got != "Line drive single to center." {
		t.Fatalf("shortLastPlayDescription() = %q, want %q", got, "Line drive single to center.")
	}
}

func TestGetScoreboardLiveGameBatter_NilSafe(t *testing.T) {
	assertNoPanic(t, "empty game", func() {
		got := getScoreboardLiveGameBatter(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, statsapi.BaseballGameRestObject{}, ScoreboardLiveGame{})
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
		got := getScoreboardLiveGameBatter(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, game, ScoreboardLiveGame{HalfInning: "top", CurrentBatterId: 700337})
		if got.FullName != "Edgar Quero" || got.LastName != "" {
			t.Fatalf("expected full name only with empty last name, got %+v", got)
		}
	})

	assertNoPanic(t, "lookup-backed batter", func() {
		info := ScoreboardInformation{
			PlayerLookupMap: &map[int32]statsapi.BaseballPersonRestObject{
				700337: {Id: int32Ptr(700337), FullName: ptr("Edgar Quero"), LastName: ptr("Quero")},
			},
		}
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Batter: &statsapi.BaseballPersonRestObject{
								Id:       int32Ptr(700337),
								FullName: ptr("Edgar Quero"),
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGameBatter(info, statsapi.PlayerStatsResponse{}, game, ScoreboardLiveGame{HalfInning: "top", CurrentBatterId: 700337})
		if got.FullName != "Edgar Quero" || got.LastName != "Quero" {
			t.Fatalf("expected lookup last name Quero, got %+v", got)
		}
	})

	assertNoPanic(t, "boxscore-backed batter names", func() {
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Batter: &statsapi.BaseballPersonRestObject{
								Id: int32Ptr(700337),
							},
						},
					},
				},
				Boxscore: &statsapi.BaseballBoxscoreRestObject{
					Teams: &map[string]statsapi.BaseballTeamBoxscore{
						"away": {
							Players: &map[string]statsapi.BaseballRosterEntryRestObject{
								"ID700337": {
									Person: &statsapi.BaseballPersonRestObject{
										Id:       int32Ptr(700337),
										FullName: ptr("Edgar Quero"),
										LastName: ptr("Quero"),
									},
									Position: &statsapi.BaseballPosition{Abbreviation: ptr("C")},
									Stats: &statsapi.StatsRestObject{
										Batting: &statsapi.BattingData{
											Hits:    int32Ptr(1),
											AtBats:  int32Ptr(2),
											Summary: ptr("1-2"),
										},
									},
								},
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGameBatter(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, game, ScoreboardLiveGame{CurrentBatterId: 700337})
		if got.FullName != "Edgar Quero" || got.LastName != "Quero" || got.CurrentPosition != "C" || got.GameHits != 1 || got.GameAtBats != 2 || got.Summary != "1-2" {
			t.Fatalf("expected boxscore-backed batter details, got %+v", got)
		}
	})
}

func TestGetScoreboardLiveGamePitcher_NilSafe(t *testing.T) {
	assertNoPanic(t, "empty game", func() {
		got := getScoreboardLiveGamePitcher(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, statsapi.BaseballGameRestObject{})
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
		got := getScoreboardLiveGamePitcher(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, game)
		if got.FullName != "Brad Lord" || got.LastName != "" {
			t.Fatalf("expected full name only with empty last name, got %+v", got)
		}
	})

	assertNoPanic(t, "lookup-backed pitcher", func() {
		info := ScoreboardInformation{
			PlayerLookupMap: &map[int32]statsapi.BaseballPersonRestObject{
				701643: {Id: int32Ptr(701643), FullName: ptr("Brad Lord"), LastName: ptr("Lord")},
			},
		}
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Pitcher: &statsapi.BaseballPersonRestObject{
								Id:       int32Ptr(701643),
								FullName: ptr("Brad Lord"),
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGamePitcher(info, statsapi.PlayerStatsResponse{}, game)
		if got.FullName != "Brad Lord" || got.LastName != "Lord" {
			t.Fatalf("expected lookup last name Lord, got %+v", got)
		}
	})

	assertNoPanic(t, "boxscore-backed pitcher names", func() {
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Plays: &statsapi.BaseballPlayByPlayRestObject{
					CurrentPlay: &statsapi.BaseballPlayRestObject{
						Matchup: &statsapi.Matchup{
							Pitcher: &statsapi.BaseballPersonRestObject{
								Id: int32Ptr(701643),
							},
						},
					},
				},
				Boxscore: &statsapi.BaseballBoxscoreRestObject{
					Teams: &map[string]statsapi.BaseballTeamBoxscore{
						"home": {
							Players: &map[string]statsapi.BaseballRosterEntryRestObject{
								"ID701643": {
									Person: &statsapi.BaseballPersonRestObject{
										Id:       int32Ptr(701643),
										FullName: ptr("Brad Lord"),
										LastName: ptr("Lord"),
										PitchHand: &statsapi.DynamicEnumRestObject{
											Code: ptr("R"),
										},
									},
								},
							},
						},
					},
				},
			},
		}
		got := getScoreboardLiveGamePitcher(ScoreboardInformation{}, statsapi.PlayerStatsResponse{}, game)
		if got.FullName != "Brad Lord" || got.LastName != "Lord" || got.Hand != "R" {
			t.Fatalf("expected boxscore-backed pitcher details, got %+v", got)
		}
	})
}

func TestFindPitcherDecision_NilSafe(t *testing.T) {
	assertNoPanic(t, "missing decision names", func() {
		game := statsapi.BaseballGameRestObject{LiveData: &statsapi.BaseballGameLiveDataRestObject{Decisions: &statsapi.BaseballDecisionRestObject{}}}
		if got := findPitcherDecision(ScoreboardInformation{}, game, PitchingDecisionWin); got != "TBD" {
			t.Fatalf("expected TBD, got %q", got)
		}
	})

	assertNoPanic(t, "lookup-backed winner", func() {
		info := ScoreboardInformation{
			PlayerLookupMap: &map[int32]statsapi.BaseballPersonRestObject{
				701643: {Id: int32Ptr(701643), LastName: ptr("Lord")},
			},
		}
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Decisions: &statsapi.BaseballDecisionRestObject{
					Winner: &statsapi.BaseballPersonRestObject{Id: int32Ptr(701643)},
				},
			},
		}
		if got := findPitcherDecision(info, game, PitchingDecisionWin); got != "Lord" {
			t.Fatalf("expected lookup-backed last name Lord, got %q", got)
		}
	})

	assertNoPanic(t, "payload last name fallback", func() {
		game := statsapi.BaseballGameRestObject{
			LiveData: &statsapi.BaseballGameLiveDataRestObject{
				Decisions: &statsapi.BaseballDecisionRestObject{
					Loser: &statsapi.BaseballPersonRestObject{LastName: ptr("Parker")},
				},
			},
		}
		if got := findPitcherDecision(ScoreboardInformation{}, game, PitchingDecisionLoss); got != "Parker" {
			t.Fatalf("expected payload last name Parker, got %q", got)
		}
	})
}

func TestPlayerLookup(t *testing.T) {
	info := ScoreboardInformation{
		PlayerLookupMap: &map[int32]statsapi.BaseballPersonRestObject{
			123: {Id: int32Ptr(123), FullName: ptr("Brad Lord")},
		},
	}

	got := playerLookup(info, int32Ptr(123))
	if got == nil || got.Id == nil || *got.Id != 123 {
		t.Fatalf("expected player 123, got %+v", got)
	}

	if got := playerLookup(info, int32Ptr(456)); got != nil {
		t.Fatalf("expected nil for missing player, got %+v", got)
	}

	if got := playerLookup(info, nil); got != nil {
		t.Fatalf("expected nil for nil player id, got %+v", got)
	}
}

func int32Ptr(v int32) *int32 { return &v }
