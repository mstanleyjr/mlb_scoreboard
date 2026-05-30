package statsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const CLIENT_SERVER = "https://statsapi.mlb.com"
const MLB_SPORTS_ID int32 = 1
const MLB_SEASON = "2026"
const TIMECODE_LAYOUT = "20060102_150405"
const playerLookupRefreshInterval = 24 * time.Hour

type MLBClient struct {
	client         *Client
	playerLookupMu sync.RWMutex
	playerLookup   map[int32]BaseballPersonRestObject
}

type LeagueResponseObject struct {
	Leagues []LeagueRestObject `json:"leagues"`
}

func NewMLBClient() (*MLBClient, error) {
	mlbClient, err := NewClient(CLIENT_SERVER)
	if err != nil {
		return nil, err
	}
	return &MLBClient{
		client:       mlbClient,
		playerLookup: make(map[int32]BaseballPersonRestObject),
	}, nil
}

func decodeJSON[T any](resp *http.Response) (T, error) {
	var value T
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

func firstPerson(people PeopleRestObject, playerID int32) (BaseballPersonRestObject, error) {
	if people.People == nil || len(*people.People) == 0 {
		return BaseballPersonRestObject{}, fmt.Errorf("no player returned for playerID %d", playerID)
	}

	return (*people.People)[0], nil
}

func (m *MLBClient) playerFromLookup(playerID int32) (BaseballPersonRestObject, bool) {
	m.playerLookupMu.RLock()
	player, ok := m.playerLookup[playerID]
	m.playerLookupMu.RUnlock()
	if !ok {
		return BaseballPersonRestObject{}, false
	}
	return player, true
}

func (m *MLBClient) storePlayerInLookup(playerID int32, player BaseballPersonRestObject) {
	m.playerLookupMu.Lock()
	m.playerLookup[playerID] = player
	m.playerLookupMu.Unlock()
}

func buildPlayerLookup(players []BaseballPersonRestObject) map[int32]BaseballPersonRestObject {
	lookup := make(map[int32]BaseballPersonRestObject, len(players))
	for _, player := range players {
		if player.Id == nil {
			continue
		}
		lookup[*player.Id] = player
	}
	return lookup
}

func clonePlayerLookup(src map[int32]BaseballPersonRestObject) map[int32]BaseballPersonRestObject {
	cloned := make(map[int32]BaseballPersonRestObject, len(src))
	for id, player := range src {
		cloned[id] = player
	}
	return cloned
}

func (m *MLBClient) PlayerLookupSnapshot() map[int32]BaseballPersonRestObject {
	m.playerLookupMu.RLock()
	defer m.playerLookupMu.RUnlock()
	return clonePlayerLookup(m.playerLookup)
}

func (m *MLBClient) GetMLBPlayers(ctx context.Context) ([]BaseballPersonRestObject, error) {
	seasonID := MLB_SEASON
	accent := false

	resp, err := m.client.SportPlayers(ctx, MLB_SPORTS_ID, &SportPlayersParams{
		Season: &seasonID,
		Accent: &accent,
	})
	if err != nil {
		println("Error getting players: ", err.Error())
		return nil, err
	}

	people, err := decodeJSON[PeopleRestObject](resp)
	if err != nil {
		println("Error decoding players response: ", err.Error())
		return nil, err
	}
	if people.People == nil {
		return nil, fmt.Errorf("no players returned for sportID %d", MLB_SPORTS_ID)
	}

	return *people.People, nil
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

	teams, err := decodeJSON[TeamsRestObject](resp)
	if err != nil {
		println("Error decoding teams response: ", err.Error())
		return TeamsRestObject{}, err
	}

	return teams, nil
}

func (m *MLBClient) GetMLBTeamSchedule(ctx context.Context, teamID int32) (ScheduleRestObject, error) {
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

	schedule, err := decodeJSON[ScheduleRestObject](resp)
	if err != nil {
		println("Error decoding schedule response: ", err.Error())
		return ScheduleRestObject{}, err
	}

	return schedule, nil
}

func (m *MLBClient) GetMLBGamesForCurrentDay(ctx context.Context) (ScheduleRestObject, error) {
	fmt.Println("Getting MLB Games for Date")
	var sportID *[]int32
	sportID = &[]int32{MLB_SPORTS_ID}

	// Time in PT for West Coast Games.
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		println("Error loading location: ", err.Error())
		return ScheduleRestObject{}, err
	}
	currTime := time.Now().In(loc)

	var reqDate *openapi_types.Date
	reqDate = &openapi_types.Date{
		Time: currTime,
	}

	resp, err := m.client.Schedule(ctx, &ScheduleParams{
		SportId: sportID,
		Date:    reqDate,
	})
	if err != nil {
		println("Error getting games for date: ", err.Error())
		return ScheduleRestObject{}, err
	}

	schedule, err := decodeJSON[ScheduleRestObject](resp)
	if err != nil {
		println("Error decoding games for date response: ", err.Error())
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

	standings, err := decodeJSON[StandingsRestObject](resp)
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
	divisions, err := decodeJSON[DivisionsRestObject](resp)
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

	leagues, err := decodeJSON[LeagueResponseObject](resp)
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

	liveGame, err := decodeJSON[BaseballGameRestObject](resp)
	if err != nil {
		println("Error decoding live game response: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	if liveGame.GamePk == nil {
		return BaseballGameRestObject{}, fmt.Errorf("live game data is nil")
	}

	return liveGame, nil
}

func (m *MLBClient) GetLiveGameDiffPatch(ctx context.Context, gamePk int32, startTime time.Time, endTime time.Time) (BaseballGameRestObject, error) {
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

	liveGamePatch, err := decodeJSON[BaseballGameRestObject](resp)
	if err != nil {
		println("Error decoding live game diff patch response: ", err.Error())
		return BaseballGameRestObject{}, err
	}

	return liveGamePatch, nil
}

func (m *MLBClient) GetMLBGameTypes(ctx context.Context) ([]GameTypeEnum, error) {
	fmt.Println("Getting MLB Game Types")
	resp, err := m.client.GameTypes(ctx, &GameTypesParams{})
	if err != nil {
		println("Error getting game types: ", err.Error())
		return nil, err
	}

	gameTypes, err := decodeJSON[[]GameTypeEnum](resp)
	if err != nil {
		println("Error decoding game types response: ", err.Error())
		return nil, err
	}
	return gameTypes, nil
}

func (m *MLBClient) GetMLBPLayerStats(ctx context.Context, playerID int32) (PlayerStatsResponse, error) {
	seasonId := MLB_SEASON

	resp, err := m.client.Stats3(ctx, playerID, &Stats3Params{
		Stats:  []StatType{StatTypeSEASON},
		Group:  &[]StatGroup{StatGroupHITTING, StatGroupPITCHING},
		Season: &seasonId,
	})
	if err != nil {
		println("Error getting player stats: ", err.Error())
		return PlayerStatsResponse{}, err
	}

	stats, err := decodeJSON[PlayerStatsResponse](resp)
	if err != nil {
		println("Error decoding player stats response: ", err.Error())
		return PlayerStatsResponse{}, err
	}
	return stats, nil
}

func (m *MLBClient) GetMLBPlayer(ctx context.Context, playerID int32) (BaseballPersonRestObject, error) {
	accent := false
	resp, err := m.client.Person(ctx, playerID, &PersonParams{Accent: &accent})
	if err != nil {
		println("Error getting player data: ", err.Error())
		return BaseballPersonRestObject{}, err
	}

	people, err := decodeJSON[PeopleRestObject](resp)
	if err != nil {
		println("Error decoding player response: ", err.Error())
		return BaseballPersonRestObject{}, err
	}

	player, err := firstPerson(people, playerID)
	if err != nil {
		return BaseballPersonRestObject{}, err
	}

	return player, nil
}
