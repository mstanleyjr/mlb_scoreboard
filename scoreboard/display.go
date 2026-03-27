package scoreboard

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

// Global display state for LED matrix
var (
	CurrentDisplayType DisplayType
	CurrentDisplayData interface{}
	DisplayMutex       sync.RWMutex
)

type DisplayType int

const (
	DisplayTypeLoading DisplayType = iota
	DisplayTypeDivisionStandings
	DisplayTypeLiveGame
	DisplayTypeNextMatchup
	DisplayTypeLastMatchup
)

func LoadingScreen() {
	println("Loading scoreboard data...")

	// Set display state for LED matrix
	DisplayMutex.Lock()
	CurrentDisplayType = DisplayTypeLoading
	CurrentDisplayData = nil
	DisplayMutex.Unlock()
}

func DivisionStandingsDisplay(info ScoreboardInformation, leagueID int32, divisionIndex int, controller *DisplayController) {
	leagueMap := info.LeagueMap

	// Safety check: ensure league exists
	league, exists := leagueMap[leagueID]
	if !exists {
		fmt.Println("League not found:", leagueID)
		return
	}

	// Safety check: ensure league has name
	if league.League.Name == nil {
		fmt.Println("League name is nil")
		return
	}

	// Safety check: ensure divisions exist and divisionIndex is valid
	if league.Divisions == nil || divisionIndex >= len(league.Divisions) {
		fmt.Println("Divisions not loaded or invalid divisionIndex:", divisionIndex)
		return
	}

	division := league.Divisions[divisionIndex]
	if division.Name == nil || division.Id == nil {
		fmt.Println("Division data incomplete")
		return
	}

	fmt.Println("Displaying standings for League: ", *league.League.Name, " ; Division : ", *division.Name)

	divisionID := division.Id
	leagueStandingsByDivision := info.Standings[leagueID].Records

	// Safety check: ensure standings data exists
	if leagueStandingsByDivision == nil {
		fmt.Println("Standings data not loaded for league:", leagueID)
		return
	}

	var standings *[]statsapi.TeamStandingsRecordRestObject
	for _, divisionRecords := range *leagueStandingsByDivision {
		foundId := divisionRecords.Division.Id
		if foundId != nil && *foundId == *divisionID {
			standings = divisionRecords.TeamRecords
			break
		}
	}

	if standings == nil {
		fmt.Println("No standings for division: ", *divisionID)
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
		LeagueName: *division.NameShort,
		Teams:      divisionTeams,
	}

	fmt.Printf("displayinfo %+v\n", displayInfo)

	// Set display state for LED matrix
	DisplayMutex.Lock()
	CurrentDisplayType = DisplayTypeDivisionStandings
	CurrentDisplayData = displayInfo
	DisplayMutex.Unlock()

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

	homeTeamName := teams["home"].Team.Name
	awayTeamName := teams["away"].Team.Name
	awayWins := teams["away"].LeagueRecord.Wins
	awayLosses := teams["away"].LeagueRecord.Losses
	homeWins := teams["home"].LeagueRecord.Wins
	homeLosses := teams["home"].LeagueRecord.Losses

	var homePitcher, awayPitcher ScoreboardPitcher

	if teams["home"].ProbablePitcher != nil {
		homePitcherId := teams["home"].ProbablePitcher.Id
		homePitcherName := teams["home"].ProbablePitcher.FullName
		homePitchHand := teams["home"].ProbablePitcher.PitchHand.Code
		homePitcherStats, err := client.GetMLBPLayerStats(ctx, *homePitcherId)
		if err != nil {
			println("Error getting home pitcher stats: ", err.Error())
		}

		homePitcherERA, homePitcherWins, homePitcherLosses, homePitcherSaves := GetScoreboardPitcherStats(homePitcherStats)
		homePitcher = ScoreboardPitcher{
			Name:   *homePitcherName,
			Hand:   *homePitchHand,
			ERA:    homePitcherERA,
			Wins:   homePitcherWins,
			Losses: homePitcherLosses,
			Saves:  homePitcherSaves,
		}
	} else {
		homePitcher = ScoreboardPitcher{
			Name: "TBD",
		}
	}

	if teams["away"].ProbablePitcher != nil {
		awayPitcherId := teams["away"].ProbablePitcher.Id
		awayPitcherName := teams["away"].ProbablePitcher.FullName
		awayPitchHand := teams["away"].ProbablePitcher.PitchHand.Code

		awayPitcherStats, err := client.GetMLBPLayerStats(ctx, *awayPitcherId)
		if err != nil {
			println("Error getting away pitcher stats: ", err.Error())
		}

		awayPitcherERA, awayPitcherWins, awayPitcherLosses, awayPitcherSaves := GetScoreboardPitcherStats(awayPitcherStats)

		awayPitcher = ScoreboardPitcher{
			Name:   *awayPitcherName,
			Hand:   *awayPitchHand,
			ERA:    awayPitcherERA,
			Wins:   awayPitcherWins,
			Losses: awayPitcherLosses,
			Saves:  awayPitcherSaves,
		}
	} else {
		awayPitcher = ScoreboardPitcher{
			Name: "TBD",
		}
	}

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

	winner := findPitcherDecision(game, PitchingDecisionWin)
	loser := findPitcherDecision(game, PitchingDecisionLoss)
	save := findPitcherDecision(game, PitchingDecisionSave)

	displayInfo := ScoreboardLastMatchup{
		AwayTeam: ScoreboardLastMatchupTeam{
			Team:   awayTeam,
			Winner: !homeTeamWinner,
		},
		HomeTeam: ScoreboardLastMatchupTeam{
			Team:   homeTeam,
			Winner: homeTeamWinner,
		},
		DateTime:       *lastGame.GameDate,
		GameStatus:     status,
		Venue:          *lastGame.Venue.Name,
		GameType:       gameType,
		FinalInning:    finalInning,
		WinningPitcher: winner,
		LosingPitcher:  loser,
		SavePitcher:    save,
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

	var currentBatterId, currentPitcherId int32

	for *liveGame.GameData.Status.AbstractGameCode == "L" {
		gameInfo, err := getLiveGameInfo(liveGame)
		if err != nil {
			println("Error getting live game info: ", err.Error())
		}

		gameInfo.GameType = gameType

		// Set display state for LED matrix
		DisplayMutex.Lock()
		CurrentDisplayType = DisplayTypeLiveGame
		CurrentDisplayData = gameInfo
		DisplayMutex.Unlock()

		if currentBatterId != gameInfo.CurrentBatterId || currentPitcherId != gameInfo.CurrentPitcherId {
			if gameInfo.CurrentBatterId != 0 && gameInfo.CurrentPitcherId != 0 {
				// Make the api calls
				currentPitcherId = gameInfo.CurrentPitcherId

				batterStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentBatterId)
				if err != nil {
					println("Error getting batter stats: ", err.Error())
				}
				currentBatterId = gameInfo.CurrentBatterId
				batter := getScoreboardLiveGameBatter(batterStats, liveGame, gameInfo)

				fmt.Printf("Batter stats: %+v\n", batter)
				gameInfo.CurrentBatter = batter

				pitcherStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentPitcherId)
				if err != nil {
					println("Error getting pitcher stats: ", err.Error())
				}
				pitcher := getScoreboardLiveGamePitcher(pitcherStats, liveGame)
				fmt.Printf("Pitcher stats: %+v\n", pitcher)
				gameInfo.CurrentPitcher = pitcher
			}
		}

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

		if gameInfo.CurrentBatterId != 0 && gameInfo.CurrentPitcherId != 0 {
			batterStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentBatterId)
			if err != nil {
				println("Error getting batter stats: ", err.Error())
			}
			batter := getScoreboardLiveGameBatter(batterStats, liveGame, gameInfo)

			fmt.Printf("Batter stats: %+v\n", batter)
			gameInfo.CurrentBatter = batter

			pitcherStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentPitcherId)
			if err != nil {
				println("Error getting pitcher stats: ", err.Error())
			}
			pitcher := getScoreboardLiveGamePitcher(pitcherStats, liveGame)
			fmt.Printf("Pitcher stats: %+v\n", pitcher)
			gameInfo.CurrentPitcher = pitcher
		}

		fmt.Printf("Live look-in for game ID %d: %+v\n", gameId, gameInfo)
		DisplayLoop(1*time.Second, callInterval, controller)
	}

	println("Finishing live look-in display")
}
