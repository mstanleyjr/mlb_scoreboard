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
	GameTypeMap             map[string]string
}

type ScoreboardDivision struct {
	LeagueName string
	Teams      []ScoreboardDivisionTeam
}

type ScoreboardDivisionTeam struct {
	Name      string
	Rank      int
	Record    ScoreboardWinLossRecord
	GamesBack string
}

type ScoreboardWinLossRecord struct {
	Wins   int
	Losses int
}

type ScoreboardNextMatchupTeam struct {
	Name            string
	ProbablePitcher string
	Record          ScoreboardWinLossRecord
}

type ScoreboardNextMatchup struct {
	AwayTeam ScoreboardNextMatchupTeam
	HomeTeam ScoreboardNextMatchupTeam
	DateTime time.Time
	Venue    string
	GameType string
}

type ScoreboardLastMatchup struct {
	AwayTeam    ScoreboardLastMatchupTeam
	HomeTeam    ScoreboardLastMatchupTeam
	DateTime    time.Time
	Venue       string
	GameType    string
	FinalInning int
}

type ScoreboardLastMatchupTeam struct {
	Name     string
	Record   ScoreboardWinLossRecord
	Runs     int
	Hits     int
	Errors   int
	LOB      int
	Winner   bool
	Decision string // Winning or losing pitcher
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

	game, err := client.GetLiveGame(ctx, *lastGame.GamePk)
	if err != nil {
		println("Error getting live game data for last completed game: ", err.Error())
		return
	}

	teams := *game.GameData.Teams
	homeTeamName := *teams["home"].TeamName
	awayTeamName := *teams["away"].TeamName
	homeTeamRuns := *game.LiveData.Linescore.Teams.Home.Runs
	awayTeamRuns := *game.LiveData.Linescore.Teams.Away.Runs
	homeTeamHits := *game.LiveData.Linescore.Teams.Home.Hits
	awayTeamHits := *game.LiveData.Linescore.Teams.Away.Hits
	homeTeamErrors := *game.LiveData.Linescore.Teams.Home.Errors
	awayTeamErrors := *game.LiveData.Linescore.Teams.Away.Errors
	homeTeamLOB := *game.LiveData.Linescore.Teams.Home.LeftOnBase
	awayTeamLOB := *game.LiveData.Linescore.Teams.Away.LeftOnBase

	homeTeamWinner := homeTeamRuns > awayTeamRuns

	// Lol there's no flag for who won the game so just have to compare run

	// Want to get pitchers of record too
	// Runs Hits Errors LOB
	// Want if it was Spring Training or Regular Season or Post or whatever

	venue := *lastGame.Venue.Name
	gameTime := *lastGame.GameDate
	fmt.Println("Last game: ", awayTeamName, " vs ", homeTeamName, " at ", venue, " on ", gameTime)
	fmt.Println("Final Score: ", awayTeamName, awayTeamRuns, " - ", homeTeamName, homeTeamRuns)

	gameType := info.GameTypeMap[*lastGame.GameType]

	displayInfo := ScoreboardLastMatchup{
		FinalInning: int(*game.LiveData.Linescore.CurrentInning),
		Venue:       *lastGame.Venue.Name,
		AwayTeam: ScoreboardLastMatchupTeam{
			Name: awayTeamName,
			Record: ScoreboardWinLossRecord{
				Wins:   int(*teams["away"].Record.LeagueRecord.Wins),
				Losses: int(*teams["away"].Record.LeagueRecord.Losses),
			},
			Runs:     int(awayTeamRuns),
			Hits:     int(awayTeamHits),
			Errors:   int(awayTeamErrors),
			LOB:      int(awayTeamLOB),
			Winner:   !homeTeamWinner,
			Decision: findPitcherDecision(game, !homeTeamWinner),
		},
		HomeTeam: ScoreboardLastMatchupTeam{
			Name: homeTeamName,
			Record: ScoreboardWinLossRecord{
				Wins:   int(*teams["home"].Record.LeagueRecord.Wins),
				Losses: int(*teams["home"].Record.LeagueRecord.Losses),
			},
			Runs:     int(homeTeamRuns),
			Hits:     int(homeTeamHits),
			Errors:   int(homeTeamErrors),
			LOB:      int(homeTeamLOB),
			Winner:   homeTeamWinner,
			Decision: findPitcherDecision(game, homeTeamWinner),
		},
		DateTime: *lastGame.GameDate,
		GameType: gameType,
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
