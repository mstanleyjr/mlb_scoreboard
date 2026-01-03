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

func StartScoreboard(ctx context.Context, wg *sync.WaitGroup) {
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
				DivisionStandingsDisplay(scoreboardInfo, li, di)
			})
		}
	}
	pages = append(pages, func(scoreboardInfo ScoreboardInformation) {
		NextMatchupDisplay(scoreboardInfo)
	})

	pageIndex := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("StartScoreboard: context canceled, exiting")
			return
		case <-scheduleRes.Done():
			// TODO: We probably don't need half of this
			// Final snapshot after producer finished
			if p := scheduleRes.Latest(); p != nil {
				fmt.Printf("Final schedule: %+v\n", *p)
			} else {
				fmt.Println("No final schedule available")
			}
			fmt.Println("StartScoreboard: schedule producer finished, exiting")
			return
		case <-nationalLeagueDivisionStandingRes.Done():
			// Final snapshot after division standings producer finished (do not exit)
			if s := nationalLeagueDivisionStandingRes.Latest(); s != nil {
				fmt.Printf("Final National League division standings: %+v\n", *s)
			} else {
				fmt.Println("No final National League division standings available")
			}
			fmt.Println("StartScoreboard: division standings producer finished")
			// continue to wait for schedule or context cancel
		case <-americanLeagueDivisionStandingRes.Done():
			// Final snapshot after division standings producer finished (do not exit)
			if s := americanLeagueDivisionStandingRes.Latest(); s != nil {
				fmt.Printf("Final American League division standings: %+v\n", *s)
			} else {
				fmt.Println("No final American League division standings available")
			}
			fmt.Println("StartScoreboard: division standings producer finished")
			// continue to wait for schedule or context cancel
		default:
		}

		scoreboardInfo := ScoreboardInformation{
			LeagueMap:               leagueMap,
			Schedule:                scheduleRes.Latest(),
			NationalLeagueStandings: nationalLeagueDivisionStandingRes.Latest(),
			AmericanLeagueStandings: americanLeagueDivisionStandingRes.Latest(),
		}

		if !IsDataLoaded(scoreboardInfo) {
			LoadingScreen()
		} else {
			pages[pageIndex](scoreboardInfo)
			time.Sleep(3 * time.Second)

			pageIndex++
			if pageIndex == len(pages) {
				pageIndex = 0
			}
		}
		// If it's within 1 hour of a game start we'll need to switch to something else

		// TODO: Do the standing page here

		// I bet I can make a list of pages to display and just cycle through them.
		// Use that stuff above to help set it up

		// I think I need to have more control here for what "page" we're on

		// THIS is the controller and how often we cycle is up to here

		select {
		case <-ctx.Done():
			fmt.Println("StartScoreboard: shutting down during sleep")
			return
		case <-time.After(loopInterval):
		}
	}

	// That stuff is more dynamic than I thought, so maybe for the schedule we call it every 5 minutes or so
	// Same with standings, we can call that on a similar schedule

	// So might be worth MORE GO ROUTINES that we can communicate with channels or something

	// So we want to get some stuff to hold in memory so we don't have to fuck with calling stuff over and over again
	// I think teams/schedule/standings/types of plays etc

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
