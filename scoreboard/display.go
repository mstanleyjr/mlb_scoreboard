package scoreboard

import (
	"fmt"

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

func DivisionStandingsDisplay(info ScoreboardInformation, leagueID int32, divisionIndex int) {
	leagueMap := info.LeagueMap
	fmt.Println("Displaying standings for League: ", *leagueMap[leagueID].League.Name, " ; Division : ", *leagueMap[leagueID].Divisions[divisionIndex].Name)
}

func NextMatchupDisplay(info ScoreboardInformation) {
	fmt.Println("Displaying next matchup from schedule...")
}
