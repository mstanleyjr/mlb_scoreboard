package scoreboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

// We'll probably have to make a client at some point

// Def need a struct

type ScoreboardInformation struct {
	LeagueMap               map[int32]League
	Schedule                *statsapi.ScheduleRestObject
	Standings               map[int32]statsapi.StandingsRestObject
	NationalLeagueStandings *statsapi.StandingsRestObject
	AmericanLeagueStandings *statsapi.StandingsRestObject
}

type ScoreboardDivision struct {
	LeagueName string
	Teams      []ScoreboardDivisionTeam
}

type ScoreboardDivisionTeam struct {
	Rank      int
	Wins      int
	Losses    int
	GamesBack string
}

func LoadingScreen() {
	println("Loading scoreboard data...")
}

func DivisionStandingsDisplay(info ScoreboardInformation, leagueID int32, divisionIndex int, controller *DisplayController) {
	leagueMap := info.LeagueMap
	fmt.Println("Displaying standings for League: ", *leagueMap[leagueID].League.Name, " ; Division : ", *leagueMap[leagueID].Divisions[divisionIndex].Name)

	divisionID := leagueMap[leagueID].Divisions[divisionIndex].Id
	leagueStandingsByDivision := info.Standings[leagueID].Records
	var standings *[]statsapi.TeamStandingsRecordRestObject
	for _, divisionRecords := range *leagueStandingsByDivision {
		if divisionRecords.Division.Id == divisionID {
			standings = divisionRecords.TeamRecords
			break
		}
	}

	// TODO: Fix the client for this.

	if standings == nil {
		println("No standings for division: ", divisionID)
		return
	}

	// Make a struct of what we want to pass in?
	// stuff like AL East
	//				1. Blue Jays W/L/DivGB
	var divisionTeams []ScoreboardDivisionTeam
	for _, standing := range *standings {
		rank, err := strconv.Atoi(*standing.DivisionRank)
		if err != nil {
			println("Error converting rank")
		}

		divisionTeams = append(divisionTeams, ScoreboardDivisionTeam{
			Rank:      rank,
			Wins:      int(*standing.Wins),
			Losses:    int(*standing.Losses),
			GamesBack: *standing.DivisionGamesBack,
		})
	}

	displayInfo := ScoreboardDivision{
		LeagueName: *leagueMap[leagueID].Divisions[divisionIndex].NameShort,
		Teams:      divisionTeams,
	}

	fmt.Printf("displayinfo %+v\n", displayInfo)

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

	for *liveGame.GameData.Status.AbstractGameCode == "L" {
		// LiveGame Process and display
		DisplayLoop(100*time.Microsecond, callInterval, controller)
		liveGame, err = m.GetLiveGameDiffPatch(ctx, *game.GamePk, currStartTime, &currEndTime)
		if err != nil {
			println("Error getting live game data: ", err.Error())
			return
		}
	}

	println("Finishing the ball game")
}
