package scoreboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

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
		foundId := divisionRecords.Division.Id
		if *foundId == *divisionID {
			standings = divisionRecords.TeamRecords
			break
		}
	}

	if standings == nil {
		println("No standings for division: ", divisionID)
		return
	}

	var divisionTeams []ScoreboardDivisionTeam
	for _, standing := range *standings {
		rank, err := strconv.Atoi(*standing.DivisionRank)
		if err != nil {
			println("Error converting rank")
		}

		divisionTeams = append(divisionTeams, ScoreboardDivisionTeam{
			Name: *standing.Team.Name,
			Rank: rank,
			Record: ScoreboardWinLossRecord{
				Wins:   int(*standing.Wins),
				Losses: int(*standing.Losses),
			},
			GamesBack: *standing.DivisionGamesBack,
		})
	}

	displayInfo := ScoreboardDivision{
		LeagueName: *leagueMap[leagueID].Divisions[divisionIndex].NameShort,
		Teams:      divisionTeams,
	}

	fmt.Printf("displayinfo %+v\n", displayInfo)

	DisplayLoop(1*time.Second, time.Second*8, controller)
	fmt.Println("Finished displaying division standings.")
}

func NextMatchupDisplay(ctx context.Context, info ScoreboardInformation, client *statsapi.MLBClient, controller *DisplayController) {
	fmt.Println("Displaying next matchup from schedule...")
	nextGame, err := FindNextScheduledGame(info)
	if err != nil {
		println("Error finding next game: ", err.Error())
		return
	}
	if nextGame == nil {
		println("No next game found.")
		return
	}

	teams := *nextGame.Teams

	// TODO: We could use some validations here for nils
	homeTeamName := teams["home"].Team.Name
	awayTeamName := teams["away"].Team.Name
	awayWins := teams["away"].LeagueRecord.Wins
	awayLosses := teams["away"].LeagueRecord.Losses
	homeWins := teams["home"].LeagueRecord.Wins
	homeLosses := teams["home"].LeagueRecord.Losses
	homePitcher := pitcherName(teams["home"].ProbablePitcher)
	awayPitcher := pitcherName(teams["away"].ProbablePitcher)
	venue := *nextGame.Venue.Name

	gameTime := *nextGame.GameDate

	gameType := info.GameTypeMap[*nextGame.GameType]
	displayInfo := ScoreboardNextMatchup{
		Venue: venue,
		AwayTeam: ScoreboardNextMatchupTeam{
			Name:            *awayTeamName,
			ProbablePitcher: awayPitcher,
			Record: ScoreboardWinLossRecord{
				Wins:   int(*awayWins),
				Losses: int(*awayLosses),
			},
		},
		HomeTeam: ScoreboardNextMatchupTeam{
			Name:            *homeTeamName,
			ProbablePitcher: homePitcher,
			Record: ScoreboardWinLossRecord{
				Wins:   int(*homeWins),
				Losses: int(*homeLosses),
			},
		},
		DateTime: gameTime,
		GameType: gameType,
	}
	fmt.Printf("displayinfo %+v\n", displayInfo)
	DisplayLoop(1*time.Second, time.Second*8, controller)
	fmt.Println("Finished displaying next matchup.")
}

func LastMatchupDisplay(ctx context.Context, info ScoreboardInformation, client *statsapi.MLBClient, controller *DisplayController) {
	fmt.Println("Displaying last completed matchup from schedule...")
	lastGame, err := FindLastCompletedGame(info)
	if err != nil {
		println("Error finding last completed game: ", err.Error())
		return
	}
	if lastGame == nil {
		println("No last completed game found.")
		return
	}

	status := *lastGame.Status.DetailedState

	game, err := client.GetLiveGame(ctx, *lastGame.GamePk)
	if err != nil {
		println("Error getting live game data for last completed game: ", err.Error())
		return
	}

	homeTeam, err := getScoreboardLiveGameTeams(game, true)
	if err != nil {
		println("Error getting home team data: ", err.Error())
		return
	}
	awayTeam, err := getScoreboardLiveGameTeams(game, false)
	if err != nil {
		println("Error getting away team data: ", err.Error())
		return
	}
	homeTeamWinner := homeTeam.Runs > awayTeam.Runs

	venue := *lastGame.Venue.Name
	gameTime := *lastGame.GameDate
	fmt.Println("Last game: ", awayTeam.Name, " vs ", homeTeam.Name, " at ", venue, " on ", gameTime)
	fmt.Println("Final Score: ", awayTeam.Name, awayTeam.Runs, " - ", homeTeam.Name, homeTeam.Runs)

	gameType := info.GameTypeMap[*lastGame.GameType]

	finalInning := 0
	if game.LiveData.Linescore.CurrentInning != nil {
		finalInning = int(*game.LiveData.Linescore.CurrentInning)
	}
	displayInfo := ScoreboardLastMatchup{
		FinalInning: finalInning,
		Venue:       *lastGame.Venue.Name,
		AwayTeam: ScoreboardLastMatchupTeam{
			Team:     awayTeam,
			Winner:   !homeTeamWinner,
			Decision: findPitcherDecision(game, !homeTeamWinner),
		},
		HomeTeam: ScoreboardLastMatchupTeam{
			Team:     homeTeam,
			Winner:   homeTeamWinner,
			Decision: findPitcherDecision(game, homeTeamWinner),
		},
		DateTime:   *lastGame.GameDate,
		GameType:   gameType,
		GameStatus: status,
	}
	fmt.Printf("displayinfo %+v\n", displayInfo)

	DisplayLoop(1*time.Second, time.Second*8, controller)
	fmt.Println("Finished displaying last completed matchup.")
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
	callInterval := time.Second * 1

	var gameType string
	if game.GameType != nil {
		gameType = getGameTypeFromLookup(*game.GameType)
		if gameType == "Regular Season" {
			// For regular season we don't show the game type since it's just implied
			gameType = ""
		}
	}

	for *liveGame.GameData.Status.AbstractGameCode == "L" {
		gameInfo, err := getLiveGameInfo(liveGame)
		if err != nil {
			println("Error getting live game info: ", err.Error())
		}

		gameInfo.GameType = gameType

		DisplayLoop(1*time.Second, callInterval, controller)
		latestInfo, err := m.GetLiveGame(ctx, *game.GamePk)
		if err != nil {
			println("Error getting live game data: ", err.Error())
			return
		}
		liveGame = latestInfo

		fmt.Printf("Displaying active game: %+v\n", liveGame.GameData.Datetime)
		fmt.Printf("gameInfo:  %+v\n\n", gameInfo)
	}

	println("Finishing the ball game")
}

func LiveLookInDisplay(ctx context.Context, info ScoreboardInformation, m *statsapi.MLBClient, controller *DisplayController) {
	gameIds := FindAllActiveGameIds(info)

	callInterval := time.Second * 10
	for _, gameId := range gameIds {
		liveGame, err := m.GetLiveGame(ctx, gameId)
		if err != nil {
			println("Error getting live game data: ", err.Error())
			continue
		}

		gameInfo, err := getLiveGameInfo(liveGame)
		if err != nil {
			println("Error getting live game info: ", err.Error())
			continue
		}

		fmt.Printf("Live look-in for game ID %d: %+v\n", gameId, gameInfo)
		DisplayLoop(1*time.Second, callInterval, controller)
	}

	println("Finishing live look-in display")
}
