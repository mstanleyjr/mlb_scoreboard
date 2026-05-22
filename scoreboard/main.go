package scoreboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

const NATIONALS_TEAM_ID = 120
const AMERICAN_LEAGUE_ID = 103
const NATIONAL_LEAGUE_ID = 104

type DisplayController struct {
	Mu     sync.Mutex
	Cond   *sync.Cond
	Paused bool
}

func StartScoreboard(ctx context.Context, wg *sync.WaitGroup, controller *DisplayController) {
	defer wg.Done()
	println("Starting MLB Scoreboard")

	mlbClient, err := statsapi.NewMLBClient()
	if err != nil {
		println(err.Error())
		return
	}

	leagues, err := mlbClient.GetMLBLeague(ctx)
	if err != nil {
		println("Error getting leagues: ", err.Error())
		return
	}

	divisions, err := mlbClient.GetMLBDivisions(ctx)
	if err != nil {
		println("Error getting divisions: ", err.Error())
		return
	}

	gameTypes, err := mlbClient.GetMLBGameTypes(ctx)
	if err != nil {
		println("Error getting game types: ", err.Error())
		return
	}

	leagueMap := BuildLeagueDivisionLookup(leagues, divisions)
	gameTypeMap := BuildGameTypeLookup(gameTypes)

	playersFetch := func(ctx context.Context) (map[int32]statsapi.BaseballPersonRestObject, error) {
		foundPlayers, err := mlbClient.GetMLBPlayers(ctx)
		if err != nil {
			return nil, err
		}
		return BuildPlayerLookup(foundPlayers), nil
	}
	playerFetchRes := NewResource[map[int32]statsapi.BaseballPersonRestObject]("players", 24*time.Hour, 0, playersFetch)
	playerFetchRes.Start(ctx, wg)

	scheduleFetch := func(ctx context.Context) (statsapi.ScheduleRestObject, error) {
		return mlbClient.GetMLBTeamSchedule(ctx, NATIONALS_TEAM_ID)
	}
	scheduleRes := NewResource[statsapi.ScheduleRestObject]("schedule", 5*time.Minute, 0, scheduleFetch)
	scheduleRes.Start(ctx, wg)

	nationalLeagueDivisionStandingFetch := func(ctx context.Context) (statsapi.StandingsRestObject, error) {
		return mlbClient.GetMLBDivisionStandings(ctx, NATIONAL_LEAGUE_ID)
	}
	nationalLeagueDivisionStandingRes := NewResource[statsapi.StandingsRestObject]("division_standings", 5*time.Minute, 0, nationalLeagueDivisionStandingFetch)
	nationalLeagueDivisionStandingRes.Start(ctx, wg)

	americanLeagueDivisionStandingFetch := func(ctx context.Context) (statsapi.StandingsRestObject, error) {
		return mlbClient.GetMLBDivisionStandings(ctx, AMERICAN_LEAGUE_ID)
	}
	americanLeagueDivisionStandingRes := NewResource[statsapi.StandingsRestObject]("division_standings", 5*time.Minute, 0, americanLeagueDivisionStandingFetch)
	americanLeagueDivisionStandingRes.Start(ctx, wg)

	todayScheduleFetch := func(ctx context.Context) (statsapi.ScheduleRestObject, error) {
		return mlbClient.GetMLBGamesForCurrentDay(ctx)
	}
	todayScheduleRes := NewResource[statsapi.ScheduleRestObject]("today_schedule", 10*time.Second, 0, todayScheduleFetch)
	todayScheduleRes.Start(ctx, wg)

	loopInterval := 1 * time.Second

	batter, err := mlbClient.GetMLBPLayerStats(ctx, 695578)
	if err != nil {
		println("Error getting batter stats: ", err.Error())
	}
	fmt.Printf("Batter stats: %+v\n", batter)

	const standingsRollupMaxDuration = 105 * time.Second

	pages := make([]func(scoreboardInfo ScoreboardInformation), 0)
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		//StandingsRollupDisplay(ctx, scoreboardInfo, controller, standingsRollupMaxDuration)
	})
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		NextMatchupDisplay(ctx, scoreboardInfo, mlbClient, controller)
	})
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		LastMatchupDisplay(ctx, scoreboardInfo, mlbClient, controller)
	})
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		LiveLookInDisplay(ctx, scoreboardInfo, mlbClient, controller)
	})

	pageIndex := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("StartScoreboard: context canceled, exiting")
			return
		case <-scheduleRes.Done():
			return
		case <-nationalLeagueDivisionStandingRes.Done():
		case <-americanLeagueDivisionStandingRes.Done():
		case <-todayScheduleRes.Done():
		default:
		}

		standingsMap := make(map[int32]statsapi.StandingsRestObject)
		if nationalLeagueDivisionStandingRes.Latest() != nil {
			standingsMap[NATIONAL_LEAGUE_ID] = *nationalLeagueDivisionStandingRes.Latest()
		}
		if americanLeagueDivisionStandingRes.Latest() != nil {
			standingsMap[AMERICAN_LEAGUE_ID] = *americanLeagueDivisionStandingRes.Latest()
		}

		scoreboardInfo := ScoreboardInformation{
			LeagueMap:               leagueMap,
			TeamSchedule:            scheduleRes.Latest(),
			TodaySchedule:           todayScheduleRes.Latest(),
			Standings:               standingsMap,
			NationalLeagueStandings: nationalLeagueDivisionStandingRes.Latest(),
			AmericanLeagueStandings: americanLeagueDivisionStandingRes.Latest(),
			GameTypeMap:             gameTypeMap,
			PlayerLookupMap:         playerFetchRes.Latest(),
		}

		activeGame, err := FindActiveTeamGame(scoreboardInfo)
		if err != nil {
			println("Error finding active game: ", err.Error())
		}

		if !IsDataLoaded(scoreboardInfo) {
			LoadingScreen()
		} else if activeGame != nil {
			fmt.Println("Active game found! Displaying live game data for game ID: ", *activeGame)
			ActiveGameDisplay(ctx, *activeGame, scoreboardInfo, mlbClient, controller)
		} else {
			pages[pageIndex](scoreboardInfo)
			pageIndex++
			if pageIndex == len(pages) {
				pageIndex = 0
			}
		}

		select {
		case <-ctx.Done():
			fmt.Println("StartScoreboard: shutting down during sleep")
			return
		case <-time.After(loopInterval):
		}
	}
}
