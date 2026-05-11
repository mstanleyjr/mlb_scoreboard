package scoreboard

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

// MockCanvas is a simple PixelCanvas implementation for use when a real canvas isn't available.
// It's useful for testing and for display functions that manage their own canvas lifecycle.
type MockCanvas struct {
	w, h int
	pix  []color.RGBA
}

func NewMockCanvas(w, h int) *MockCanvas {
	return &MockCanvas{w: w, h: h, pix: make([]color.RGBA, w*h)}
}

func (m *MockCanvas) Set(x, y int, c color.Color) {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return
	}
	r, g, b, a := c.RGBA()
	m.pix[y*m.w+x] = color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func (m *MockCanvas) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.w, m.h)
}

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

	// Publish state so render loop draws this page.
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
	if nextGame.GamePk == nil {
		println("Next game has no gamePk.")
		return
	}

	// Source of truth for this page is the direct game payload.
	foundNextGame, err := client.GetLiveGame(ctx, *nextGame.GamePk)
	if err != nil {
		println("Error getting live game data for next game: ", err.Error())
		return
	}

	fmt.Printf("foundNextGame: %+v\n", foundNextGame)

	scheduleTeams := map[string]statsapi.BaseballScheduleItemTeamRestObject{}
	if nextGame.Teams != nil {
		scheduleTeams = *nextGame.Teams
	}
	liveTeams := map[string]statsapi.BaseballTeamRestObject{}
	if foundNextGame.GameData != nil && foundNextGame.GameData.Teams != nil {
		liveTeams = *foundNextGame.GameData.Teams
	}

	homeSchedule := scheduleTeams["home"]
	awaySchedule := scheduleTeams["away"]
	homeLive := liveTeams["home"]
	awayLive := liveTeams["away"]

	var homeLivePitcher, awayLivePitcher *statsapi.BaseballPersonRestObject
	if foundNextGame.GameData != nil && foundNextGame.GameData.ProbablePitchers != nil {
		homeLivePitcher = foundNextGame.GameData.ProbablePitchers.Home
		awayLivePitcher = foundNextGame.GameData.ProbablePitchers.Away
	}

	homePitcher := buildProbablePitcher(ctx, client, "home", homeLivePitcher, homeSchedule.ProbablePitcher)
	awayPitcher := buildProbablePitcher(ctx, client, "away", awayLivePitcher, awaySchedule.ProbablePitcher)
	fmt.Printf("awayPticher %+v\n\n", awayPitcher)

	homeWins, homeLosses := resolveRecord(homeLive.Record, homeSchedule.LeagueRecord)
	awayWins, awayLosses := resolveRecord(awayLive.Record, awaySchedule.LeagueRecord)

	venue := ""
	if foundNextGame.GameData != nil && foundNextGame.GameData.Venue != nil {
		venue = chooseString(foundNextGame.GameData.Venue.Name, nil, "")
	}
	if venue == "" && nextGame.Venue != nil {
		venue = chooseString(nextGame.Venue.Name, nil, "")
	}

	gameTime := time.Time{}
	if foundNextGame.GameData != nil && foundNextGame.GameData.Datetime != nil && foundNextGame.GameData.Datetime.DateTime != nil {
		gameTime = *foundNextGame.GameData.Datetime.DateTime
	} else if nextGame.GameDate != nil {
		gameTime = *nextGame.GameDate
	}

	gameTypeCode := ""
	if foundNextGame.GameData != nil && foundNextGame.GameData.Game != nil && foundNextGame.GameData.Game.Type != nil {
		gameTypeCode = *foundNextGame.GameData.Game.Type
	} else if nextGame.GameType != nil {
		gameTypeCode = *nextGame.GameType
	}
	gameType := info.GameTypeMap[gameTypeCode]

	displayInfo := ScoreboardNextMatchup{
		Venue: venue,
		AwayTeam: ScoreboardNextMatchupTeam{
			Name:            chooseString(awayLive.Name, scheduleTeamName(awaySchedule), "TBD"),
			ShortName:       chooseString(awayLive.Abbreviation, scheduleTeamAbbreviation(awaySchedule), ""),
			ProbablePitcher: awayPitcher,
			Record: ScoreboardWinLossRecord{
				Wins:   awayWins,
				Losses: awayLosses,
			},
		},
		HomeTeam: ScoreboardNextMatchupTeam{
			Name:            chooseString(homeLive.Name, scheduleTeamName(homeSchedule), "TBD"),
			ShortName:       chooseString(homeLive.Abbreviation, scheduleTeamAbbreviation(homeSchedule), ""),
			ProbablePitcher: homePitcher,
			Record: ScoreboardWinLossRecord{
				Wins:   homeWins,
				Losses: homeLosses,
			},
		},
		DateTime: gameTime,
		GameType: gameType,
	}
	fmt.Printf("displayinfo %+v\n", displayInfo)

	DisplayMutex.Lock()
	CurrentDisplayType = DisplayTypeNextMatchup
	CurrentDisplayData = displayInfo
	DisplayMutex.Unlock()

	DisplayLoop(1*time.Second, time.Second*8, controller)
	fmt.Println("Finished displaying next matchup.")
}

func chooseString(primary *string, fallback *string, defaultValue string) string {
	if primary != nil && *primary != "" {
		return *primary
	}
	if fallback != nil && *fallback != "" {
		return *fallback
	}
	return defaultValue
}

func scheduleTeamName(t statsapi.BaseballScheduleItemTeamRestObject) *string {
	if t.Team == nil {
		return nil
	}
	return t.Team.Name
}

func scheduleTeamAbbreviation(t statsapi.BaseballScheduleItemTeamRestObject) *string {
	if t.Team == nil {
		return nil
	}
	return t.Team.Abbreviation
}

func resolveRecord(liveRecord *statsapi.TeamStandingsRecordRestObject, scheduleRecord *statsapi.WinLossRecordRestObject) (wins int, losses int) {
	if liveRecord != nil {
		if liveRecord.Wins != nil {
			wins = int(*liveRecord.Wins)
		}
		if liveRecord.Losses != nil {
			losses = int(*liveRecord.Losses)
		}
	}
	if scheduleRecord != nil {
		if liveRecord == nil || liveRecord.Wins == nil {
			if scheduleRecord.Wins != nil {
				wins = int(*scheduleRecord.Wins)
			}
		}
		if liveRecord == nil || liveRecord.Losses == nil {
			if scheduleRecord.Losses != nil {
				losses = int(*scheduleRecord.Losses)
			}
		}
	}
	return wins, losses
}

func buildProbablePitcher(ctx context.Context, client *statsapi.MLBClient, side string, livePitcher *statsapi.BaseballPersonRestObject, schedulePitcher *statsapi.BaseballPersonRestObject) ScoreboardPitcher {
	pitcher := livePitcher
	if pitcher == nil {
		pitcher = schedulePitcher
	}
	if pitcher == nil {
		return ScoreboardPitcher{FullName: "TBD"}
	}

	pitcherName := chooseString(pitcher.FullName, nil, "TBD")
	pitchHand := getPitchHandFromPerson(pitcher)
	if pitchHand == "" && pitcher.Id != nil {
		if fullPitcher, err := client.GetMLBPlayer(ctx, *pitcher.Id); err == nil {
			pitchHand = getPitchHandFromPerson(&fullPitcher)
		}
	}

	if pitcher.Id == nil {
		return ScoreboardPitcher{FullName: pitcherName, Hand: pitchHand}
	}

	pitcherStats, err := client.GetMLBPLayerStats(ctx, *pitcher.Id)
	if err != nil {
		println("Error getting "+side+" pitcher stats: ", err.Error())
		return ScoreboardPitcher{FullName: pitcherName, Hand: pitchHand}
	}

	era, wins, losses, saves := GetScoreboardPitcherStats(pitcherStats)
	return ScoreboardPitcher{
		FullName: pitcherName,
		Hand:     pitchHand,
		ERA:      era,
		Wins:     wins,
		Losses:   losses,
		Saves:    saves,
	}
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
		DateTime:               *lastGame.GameDate,
		GameStatus:             status,
		Venue:                  *lastGame.Venue.Name,
		GameType:               gameType,
		FinalInning:            finalInning,
		WinningPitcherLastName: winner,
		LosingPitcherLastName:  loser,
		SavePitcherLastName:    save,
	}
	fmt.Printf("displayinfo %+v\n", displayInfo)

	DisplayMutex.Lock()
	CurrentDisplayType = DisplayTypeLastMatchup
	CurrentDisplayData = displayInfo
	DisplayMutex.Unlock()

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

		// Keep display state updated; renderer loop draws from CurrentDisplayData.
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

		DisplayMutex.Lock()
		CurrentDisplayType = DisplayTypeLiveGame
		CurrentDisplayData = gameInfo
		DisplayMutex.Unlock()

		DisplayLoop(1*time.Second, callInterval, controller)
	}

	println("Finishing live look-in display")
}
