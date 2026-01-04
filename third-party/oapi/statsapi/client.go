package statsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

const CLIENT_SERVER = "https://statsapi.mlb.com"
const MLB_SPORTS_ID int32 = 1
const MLB_SEASON = "2026"
const TIMECODE_LAYOUT = "YYYYMMDD_HHMMSS"

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

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("booo", err)
	}
	bodyString := string(bodyBytes)
	fmt.Println("Schedule response body: ", bodyString)

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

func (m *MLBClient) GetLiveGame(ctx context.Context, gamePk int32) (BaseballGameRestObject, error) {
	fmt.Println("Getting Live Game Data")
	resp, err := m.client.LiveGameV1(ctx, gamePk, &LiveGameV1Params{})
	if err != nil {
		println("Error getting live game data: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	var liveGame LiveGameV1Response
	err = json.NewDecoder(resp.Body).Decode(&liveGame)
	if err != nil {
		println("Error decoding live game response: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	game := liveGame.ApplicationjsonCharsetUTF8200

	return *game, nil
}

func (m *MLBClient) GetLiveGameDiffPatch(ctx context.Context, gamePk int32, startTime *time.Time, endTime *time.Time) (BaseballGameRestObject, error) {
	fmt.Println("Getting Live Game Data Diff Patch")

	//20250928_190807
	startTimeCode := startTime.Format(TIMECODE_LAYOUT)
	endTimeCode := endTime.Format(TIMECODE_LAYOUT)

	resp, err := m.client.LiveGameDiffPatchV1(ctx, gamePk, &LiveGameDiffPatchV1Params{
		StartTimecode: &startTimeCode,
		EndTimecode:   &endTimeCode,
	})
	if err != nil {
		println("Error getting live game diff patch data: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	var liveGamePatch LiveGameV1Response
	err = json.NewDecoder(resp.Body).Decode(&liveGamePatch)
	if err != nil {
		println("Error decoding live game diff patch response: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	game := liveGamePatch.ApplicationjsonCharsetUTF8200

	return *game, nil
}

//type ScheduleTeam struct {
//	Id   int    `json:"id"`
//	Name string `json:"name"`
//	Link string `json:"link"`
//}
//
//type ScheduleLeagueRecord struct {
//	Wins   int    `json:"wins"`
//	Losses int    `json:"losses"`
//	Pct    string `json:"pct"`
//}
//
//type ScheduleTeamsHomeAway struct {
//	Team         ScheduleTeam         `json:"team"`
//	LeagueRecord ScheduleLeagueRecord `json:"leagueRecord"`
//	SplitSquad   bool                 `json:"splitSquad"`
//	SeriesNumber int                  `json:"seriesNumber"`
//}
//
//type ScheduleTeams struct {
//	Home ScheduleTeamsHomeAway `json:"home"`
//	Away ScheduleTeamsHomeAway `json:"away"`
//}
//
//type ScheduleVenue struct {
//	Id   int    `json:"id"`
//	Name string `json:"name"`
//	Link string `json:"link"`
//}
//
//type ScheduleContent struct {
//	Link string `json:"link"`
//}
//
//type Status struct {
//	AbstractGameState string `json:"abstractGameState"`
//	CodedGameState    string `json:"codedGameState"`
//	DetailedState     string `json:"detailedState"`
//	StatusCode        string `json:"statusCode"`
//	StartTimeTBD      bool   `json:"startTimeTBD"`
//	AbstractGameCode  string `json:"abstractGameCode"`
//}
//
//type ScheduleGame struct {
//	GamePk                 int             `json:"gamePk"`
//	GameGuid               string          `json:"gameGuid"`
//	Link                   string          `json:"link"`
//	GameType               string          `json:"gameType"`
//	Season                 string          `json:"season"`
//	GameDate               time.Time       `json:"gameDate"`
//	OfficialDate           string          `json:"officialDate"`
//	Status                 Status          `json:"status"`
//	Teams                  ScheduleTeams   `json:"teams"`
//	Venue                  ScheduleVenue   `json:"venue"`
//	Content                ScheduleContent `json:"content"`
//	GameNumber             int             `json:"gameNumber"`
//	PublicFacing           bool            `json:"publicFacing"`
//	DoubleHeader           string          `json:"doubleHeader"`
//	GamedayType            string          `json:"gamedayType"`
//	Tiebreaker             string          `json:"tiebreaker"`
//	CalendarEventID        string          `json:"calendarEventID"`
//	SeasonDisplay          string          `json:"seasonDisplay"`
//	DayNight               string          `json:"dayNight"`
//	ScheduledInnings       int             `json:"scheduledInnings"`
//	ReverseHomeAwayStatus  bool            `json:"reverseHomeAwayStatus"`
//	InningBreakLength      int             `json:"inningBreakLength"`
//	GamesInSeries          int             `json:"gamesInSeries"`
//	SeriesGameNumber       int             `json:"seriesGameNumber"`
//	SeriesDescription      string          `json:"seriesDescription"`
//	RecordSource           string          `json:"recordSource"`
//	IfNecessary            string          `json:"ifNecessary"`
//	IfNecessaryDescription string          `json:"ifNecessaryDescription"`
//}
//
//type ScheduleDates struct {
//	Date                 string         `json:"date"`
//	TotalItems           int            `json:"totalItems"`
//	TotalEvents          int            `json:"totalEvents"`
//	TotalGames           int            `json:"totalGames"`
//	TotalGamesInProgress int            `json:"totalGamesInProgress"`
//	Events               []interface{}  `json:"events"`
//	Games                []ScheduleGame `json:"games"`
//}
//
//type ScheduleObject struct {
//	Copyright            string          `json:"copyright"`
//	TotalItems           int             `json:"totalItems"`
//	TotalEvents          int             `json:"totalEvents"`
//	TotalGames           int             `json:"totalGames"`
//	TotalGamesInProgress int             `json:"totalGamesInProgress"`
//	Dates                []ScheduleDates `json:"dates"`
//}
