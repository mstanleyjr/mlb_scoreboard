package scoreboard

import (
	"context"
	"fmt"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

// We'll probably have to make a client at some point

// Def need a struct

type ScoreboardInformation struct {
	LeagueMap               map[int32]League
	Schedule                *statsapi.ScheduleRestObject
	NationalLeagueStandings *statsapi.StandingsRestObject
	AmericanLeagueStandings *statsapi.StandingsRestObject
}

func LoadingScreen() {
	println("Loading scoreboard data...")
}

func DivisionStandingsDisplay(info ScoreboardInformation, leagueID int32, divisionIndex int, controller *DisplayController) {
	leagueMap := info.LeagueMap
	fmt.Println("Displaying standings for League: ", *leagueMap[leagueID].League.Name, " ; Division : ", *leagueMap[leagueID].Divisions[divisionIndex].Name)
	DisplayLoop(1*time.Second, time.Second*5, controller) // TODO: We'll need to turn down the checkInterval
	fmt.Println("Finished displaying division standings.")
}

func NextMatchupDisplay(info ScoreboardInformation, controller *DisplayController) {
	fmt.Println("Displaying next matchup from schedule...")
	DisplayLoop(1*time.Second, time.Second*5, controller)
	fmt.Println("Finished displaying next matchup.")
}

func ActiveGameDisplay(ctx context.Context, game statsapi.BaseballScheduleItemRestObject, m *statsapi.MLBClient, controller *DisplayController) {

	teams := *game.Teams
	homeTeam := teams["home"].Team.Name
	awayTeam := teams["away"].Team.Name
	fmt.Println("Displaying active game: ", homeTeam, " vs ", awayTeam)

	liveGame, err := m.GetLiveGame(ctx, *game.GamePk)
	if err != nil {
		println("Error getting live game data: ", err.Error())
		return
	}

	// Looping duration
	callInterval := time.Second

	currStartTime := liveGame.GameData.Datetime.DateTime
	currEndTime := currStartTime.Add(callInterval)

	// Ugh I gotta do the same shit again.
	for *liveGame.GameData.Status.AbstractGameCode == "L" {
		// LiveGame Process and display
		DisplayLoop(100*time.Microsecond, callInterval, controller)
		liveGame, err = m.GetLiveGameDiffPatch(ctx, *game.GamePk, currStartTime, &currEndTime)
		if err != nil {
			println("Error getting live game data: ", err.Error())
			return
		}
	}

	//

	// Loop the live game?
	// Gotta figure that out

	// So we loop it and keep the current at bat until it is over?
	// Man if we had previous play result in there we wouldn't have to loop?

	// So if the current at bat in the found data changes, we can then print the result
	// And then change to the new at bat

	// So I wonder if I
	// I'll need to need a slightly different loop that isn't time based here so we can
	// So it would be the same check but a different conditional
}
