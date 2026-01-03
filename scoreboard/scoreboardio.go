package scoreboard

import "github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"

type League struct {
	League    statsapi.LeagueRestObject
	Divisions []statsapi.DivisionRestObject
}

func BuildLeagueDivisionLookup(leagues statsapi.LeagueResponseObject, divisions statsapi.DivisionsRestObject) map[int32]League {
	lookup := make(map[int32]League)
	for _, league := range leagues.Leagues {
		if *league.Id == AMERICAN_LEAGUE_ID || *league.Id == NATIONAL_LEAGUE_ID {
			lookup[*league.Id] = League{
				League: league}
		}
	}

	for _, division := range *divisions.Divisions {
		leagueID := *division.League.Id
		if league, exists := lookup[leagueID]; exists {
			league.Divisions = append(league.Divisions, division)
			lookup[leagueID] = league
		}
	}
	return lookup
}

func IsDataLoaded(info ScoreboardInformation) bool {
	return info.AmericanLeagueStandings != nil || info.NationalLeagueStandings != nil || info.Schedule != nil
}
