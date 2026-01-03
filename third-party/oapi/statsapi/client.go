package statsapi

import (
	"context"
	"encoding/json"
	"fmt"
)

const CLIENT_SERVER = "https://statsapi.mlb.com"
const MLB_SPORTS_ID int32 = 1
const MLB_SEASON = "2026"

type MLBClient struct {
	client *Client
}

type LeagueResponseObject struct {
	Leagues []LeagueRestObject `json:"leagues"`
}

func NewMLBClient() (*MLBClient, error) {
	mlbClient, err := NewClient(CLIENT_SERVER)
	if err != nil {
		return nil, err
	}
	return &MLBClient{client: mlbClient}, nil
}

func (m *MLBClient) GetMLBTeams(ctx context.Context) (TeamsRestObject, error) {
	fmt.Println("Getting MLB Teams")
	sportID := new(int32)
	*sportID = MLB_SPORTS_ID

	resp, err := m.client.Teams(ctx, &TeamsParams{
		SportId: sportID,
	})
	if err != nil {
		println("Error getting teams: ", err.Error())
		return TeamsRestObject{}, err
	}

	var teams TeamsRestObject
	err = json.NewDecoder(resp.Body).Decode(&teams)
	if err != nil {
		println("Error decoding teams response: ", err.Error())
		return TeamsRestObject{}, err
	}

	return teams, nil
}

func (m *MLBClient) GetMLBSchedule(ctx context.Context, teamID int32) (ScheduleRestObject, error) {
	fmt.Println("Getting MLB Schedule")
	var sportID *[]int32
	sportID = &[]int32{MLB_SPORTS_ID}

	var internalTeamID *[]int32
	internalTeamID = &[]int32{teamID}

	var seasonId *[]string
	seasonId = &[]string{MLB_SEASON}

	resp, err := m.client.Schedule(ctx, &ScheduleParams{
		TeamId:  internalTeamID,
		SportId: sportID,
		Season:  seasonId,
	})
	if err != nil {
		println("Error getting schedule: ", err.Error())
		return ScheduleRestObject{}, err
	}

	var schedule ScheduleRestObject

	err = json.NewDecoder(resp.Body).Decode(&schedule)
	if err != nil {
		println("Error decoding schedule response: ", err.Error())
		return ScheduleRestObject{}, err
	}

	return schedule, nil
}

func (m *MLBClient) GetMLBDivisionStandings(ctx context.Context, leagueID int32) (StandingsRestObject, error) {
	fmt.Println("Getting MLB DivisionStandings")
	var internalLeagueID *[]int32
	internalLeagueID = &[]int32{leagueID}

	seasonID := new(string)
	*seasonID = MLB_SEASON

	resp, err := m.client.Standings(ctx, "byDivision", &StandingsParams{
		LeagueId: internalLeagueID,
		Season:   seasonID,
	})
	if err != nil {
		println("Error getting standings: ", err.Error())
		return StandingsRestObject{}, err
	}

	var standings StandingsRestObject
	err = json.NewDecoder(resp.Body).Decode(&standings)
	if err != nil {
		println("Error decoding standings response: ", err.Error())
		return StandingsRestObject{}, err
	}
	return standings, nil
}

func (m *MLBClient) GetMLBDivisions(ctx context.Context) (DivisionsRestObject, error) {
	fmt.Println("Getting MLB Divisions")
	sportID := new(int32)
	*sportID = MLB_SPORTS_ID

	resp, err := m.client.Divisions(ctx, &DivisionsParams{
		SportId: sportID,
	})
	if err != nil {
		println("Error getting divisions: ", err.Error())
		return DivisionsRestObject{}, err
	}
	var divisions DivisionsRestObject
	err = json.NewDecoder(resp.Body).Decode(&divisions)
	if err != nil {
		println("Error decoding divisions response: ", err.Error())
		return DivisionsRestObject{}, err
	}
	return divisions, nil
}

func (m *MLBClient) GetMLBLeague(ctx context.Context) (LeagueResponseObject, error) {
	fmt.Println("Getting MLB Leagues")
	sportID := new(int32)
	*sportID = MLB_SPORTS_ID

	resp, err := m.client.League(ctx, &LeagueParams{
		SportId: sportID,
	})
	if err != nil {
		println("Error getting leagues: ", err.Error())
		return LeagueResponseObject{}, err
	}

	var leagues LeagueResponseObject
	err = json.NewDecoder(resp.Body).Decode(&leagues)
	if err != nil {
		println("Error decoding leagues response: ", err.Error())
		return LeagueResponseObject{}, err
	}
	return leagues, nil
}
