package scoreboard

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"sort"
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

func StandingsRollupDisplay(ctx context.Context, info ScoreboardInformation, controller *DisplayController, maxDuration time.Duration) {
	_ = ctx

	displayInfo := buildStandingsRollup(info)
	if len(displayInfo.Sections) == 0 {
		fmt.Println("No standings sections available for rollup.")
		return
	}

	DisplayMutex.Lock()
	CurrentDisplayType = DisplayTypeDivisionStandings
	CurrentDisplayData = displayInfo
	DisplayMutex.Unlock()

	scrollDivisionStandings(displayInfo, controller, standingsRollupDuration(displayInfo, maxDuration), 100*time.Millisecond)
	fmt.Println("Finished displaying standings rollup.")
}

func buildStandingsRollup(info ScoreboardInformation) ScoreboardDivision {
	sections := make([]ScoreboardStandingsSection, 0)
	for _, leagueID := range []int32{AMERICAN_LEAGUE_ID, NATIONAL_LEAGUE_ID} {
		league, exists := info.LeagueMap[leagueID]
		if !exists {
			continue
		}

		divisions := append([]statsapi.DivisionRestObject(nil), league.Divisions...)
		sort.SliceStable(divisions, func(i, j int) bool {
			iTitle := chooseString(divisions[i].NameShort, divisions[i].Name, "")
			jTitle := chooseString(divisions[j].NameShort, divisions[j].Name, "")
			return iTitle < jTitle
		})

		standings, ok := info.Standings[leagueID]
		if !ok {
			continue
		}

		for _, division := range divisions {
			teams := standingsTeamsForDivision(standings, division)
			if len(teams) == 0 {
				continue
			}
			sections = append(sections, ScoreboardStandingsSection{
				Title: chooseString(division.NameShort, division.Name, ""),
				Teams: teams,
			})
		}
	}

	return ScoreboardDivision{Sections: sections}
}

func standingsTeamsForDivision(standings statsapi.StandingsRestObject, division statsapi.DivisionRestObject) []ScoreboardDivisionTeam {
	if standings.Records == nil || division.Id == nil {
		return nil
	}

	for _, divisionRecords := range *standings.Records {
		if divisionRecords.Division.Id == nil || *divisionRecords.Division.Id != *division.Id || divisionRecords.TeamRecords == nil {
			continue
		}

		teams := make([]ScoreboardDivisionTeam, 0, len(*divisionRecords.TeamRecords))
		for _, standing := range *divisionRecords.TeamRecords {
			if standing.Team == nil || standing.Team.Name == nil {
				continue
			}

			rank := 0
			if standing.DivisionRank != nil {
				if parsedRank, err := strconv.Atoi(*standing.DivisionRank); err == nil {
					rank = parsedRank
				}
			}

			wins := 0
			if standing.Wins != nil {
				wins = int(*standing.Wins)
			}
			losses := 0
			if standing.Losses != nil {
				losses = int(*standing.Losses)
			}
			gamesBack := ""
			if standing.DivisionGamesBack != nil {
				gamesBack = *standing.DivisionGamesBack
			}

			teams = append(teams, ScoreboardDivisionTeam{
				Name:      *standing.Team.Name,
				ShortName: chooseString(standing.Team.TeamName, standing.Team.ClubName, chooseString(standing.Team.Abbreviation, standing.Team.Name, "")),
				Rank:      rank,
				Record: ScoreboardWinLossRecord{
					Wins:   wins,
					Losses: losses,
				},
				GamesBack: gamesBack,
			})
		}

		sort.SliceStable(teams, func(i, j int) bool {
			if teams[i].Rank != teams[j].Rank {
				return teams[i].Rank < teams[j].Rank
			}
			return teams[i].Name < teams[j].Name
		})
		return teams
	}

	return nil
}

func scrollDivisionStandings(displayInfo ScoreboardDivision, controller *DisplayController, displayDuration time.Duration, frameInterval time.Duration) {
	if frameInterval <= 0 {
		frameInterval = 100 * time.Millisecond
	}

	totalFrames := int(displayDuration / frameInterval)
	if totalFrames < 1 {
		totalFrames = 1
	}
	totalFrames++

	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	for frame := 0; frame < totalFrames; frame++ {
		controller.Mu.Lock()
		for controller.Paused {
			controller.Cond.Wait()
		}
		controller.Mu.Unlock()

		frameDisplay := displayInfo
		frameDisplay.ScrollOffset = divisionStandingsScrollOffset(displayInfo, frame, totalFrames)

		DisplayMutex.Lock()
		CurrentDisplayType = DisplayTypeDivisionStandings
		CurrentDisplayData = frameDisplay
		DisplayMutex.Unlock()

		if frame == totalFrames-1 {
			return
		}

		<-ticker.C
	}
}

func divisionStandingsScrollOffset(displayInfo ScoreboardDivision, frameIndex int, totalFrames int) int {
	maxOffset := divisionStandingsMaxScrollOffset(displayInfo)
	if maxOffset == 0 || totalFrames <= 1 {
		return 0
	}

	holdFrames := totalFrames / 8
	if holdFrames > 10 {
		holdFrames = 10
	}
	if holdFrames*2 >= totalFrames {
		holdFrames = (totalFrames - 1) / 2
	}

	if frameIndex < holdFrames {
		return 0
	}
	if frameIndex >= totalFrames-holdFrames {
		return maxOffset
	}

	scrollFrames := totalFrames - (2 * holdFrames)
	if scrollFrames <= 1 {
		return maxOffset
	}

	progressNum := frameIndex - holdFrames
	progressDen := scrollFrames - 1
	return (maxOffset*progressNum + progressDen/2) / progressDen
}

func standingsRollupDuration(displayInfo ScoreboardDivision, maxDuration time.Duration) time.Duration {
	maxOffset := divisionStandingsMaxScrollOffset(displayInfo)
	duration := 8*time.Second + time.Duration(maxOffset)*30*time.Millisecond
	if maxDuration > 0 && duration > maxDuration {
		return maxDuration
	}
	return duration
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

	homePitcher := buildProbablePitcher(ctx, info, client, "home", homeLivePitcher, homeSchedule.ProbablePitcher)
	awayPitcher := buildProbablePitcher(ctx, info, client, "away", awayLivePitcher, awaySchedule.ProbablePitcher)

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

func buildProbablePitcher(ctx context.Context, info ScoreboardInformation, client *statsapi.MLBClient, side string, livePitcher *statsapi.BaseballPersonRestObject, schedulePitcher *statsapi.BaseballPersonRestObject) ScoreboardPitcher {
	pitcher := livePitcher
	if pitcher == nil {
		pitcher = schedulePitcher
	}
	if pitcher == nil {
		return ScoreboardPitcher{FullName: "TBD"}
	}

	pitcherSource := pitcher
	if foundPitcher := playerLookup(info, pitcher.Id); foundPitcher != nil {
		pitcherSource = foundPitcher
	}

	pitcherName := chooseString(pitcherSource.FullName, pitcher.FullName, "TBD")
	pitcherLastName := chooseString(pitcherSource.LastName, pitcher.LastName, "")
	pitchHand := getPitchHandFromPerson(pitcherSource)

	if pitcher.Id == nil {
		return ScoreboardPitcher{FullName: pitcherName, LastName: pitcherLastName, Hand: pitchHand}
	}

	pitcherStats, err := client.GetMLBPLayerStats(ctx, *pitcher.Id)
	if err != nil {
		println("Error getting "+side+" pitcher stats: ", err.Error())
		return ScoreboardPitcher{FullName: pitcherName, LastName: pitcherLastName, Hand: pitchHand}
	}

	era, wins, losses, saves := GetScoreboardPitcherStats(pitcherStats)
	return ScoreboardPitcher{
		FullName: pitcherName,
		LastName: pitcherLastName,
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

	winner := findPitcherDecision(info, game, PitchingDecisionWin)
	loser := findPitcherDecision(info, game, PitchingDecisionLoss)
	save := findPitcherDecision(info, game, PitchingDecisionSave)

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

func ActiveGameDisplay(ctx context.Context, game statsapi.BaseballScheduleItemRestObject, info ScoreboardInformation, m *statsapi.MLBClient, controller *DisplayController) {
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
				batter := getScoreboardLiveGameBatter(info, batterStats, liveGame, gameInfo)

				fmt.Printf("Batter stats: %+v\n", batter)
				gameInfo.CurrentBatter = batter

				pitcherStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentPitcherId)
				if err != nil {
					println("Error getting pitcher stats: ", err.Error())
				}
				pitcher := getScoreboardLiveGamePitcher(info, pitcherStats, liveGame)
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
			batter := getScoreboardLiveGameBatter(info, batterStats, liveGame, gameInfo)

			fmt.Printf("Batter stats: %+v\n", batter)
			gameInfo.CurrentBatter = batter

			pitcherStats, err := m.GetMLBPLayerStats(ctx, gameInfo.CurrentPitcherId)
			if err != nil {
				println("Error getting pitcher stats: ", err.Error())
			}
			pitcher := getScoreboardLiveGamePitcher(info, pitcherStats, liveGame)
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
