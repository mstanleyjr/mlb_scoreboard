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

	leagueMap := BuildLeagueDivisionLookup(leagues, divisions)

	scheduleFetch := func(ctx context.Context) (statsapi.ScheduleRestObject, error) {
		return mlbClient.GetMLBSchedule(ctx, NATIONALS_TEAM_ID)
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

	loopInterval := 1 * time.Second

	pages := make([]func(scoreboardInfo ScoreboardInformation), 0)
	for _, league := range leagueMap {
		for divisionIndex := range league.Divisions {
			li := *league.League.Id
			di := divisionIndex
			pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
				DivisionStandingsDisplay(scoreboardInfo, li, di, controller)
			})
		}
	}
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		NextMatchupDisplay(scoreboardInfo, controller)
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
			Schedule:                scheduleRes.Latest(),
			Standings:               standingsMap,
			NationalLeagueStandings: nationalLeagueDivisionStandingRes.Latest(),
			AmericanLeagueStandings: americanLeagueDivisionStandingRes.Latest(),
		}

		activeGame, err := FindActiveGame(scoreboardInfo)
		if err != nil {
			println("Error finding active game: ", err.Error())
		}

		if !IsDataLoaded(scoreboardInfo) {
			LoadingScreen()
		} else if activeGame != nil {
			fmt.Println("Active game found! Displaying live game data for game ID: ", activeGame)
			ActiveGameDisplay(ctx, *activeGame, mlbClient, controller)
		} else {
			pages[pageIndex](scoreboardInfo)
			pageIndex++
			if pageIndex == len(pages) {
				pageIndex = 0
			}
		}
		// If it's within 1 hour of a game start we'll need to switch to something else

		select {
		case <-ctx.Done():
			fmt.Println("StartScoreboard: shutting down during sleep")
			return
		case <-time.After(loopInterval):
		}
	}

	// So I think if it's not within 2 hours of a nats game
	// We cycle through league standings and at the bottom have next Nats game
	// Maybe we also do a nats game for potential pitching matchups and time etc
	// And a last game where we do Winner losser RHEL, total game time (final final/10)

	// Then for game time we should draw that out but ultimately will just keep hitting the endpoint for the next play (that's been completed)
	// Once we get that, we can draw the play
	// Maybe at the top we have score RHEL BSO Pitcher/Batter and Inning
	// The bottom we could have the last couple of plays? That'd be cool if there's enough room but that stuff I really can't do until I get the board going

	// So we have to goroutines that give us the data except the live game data and we can just hit that one a lot and if there's a new completed play then update the board?

}
