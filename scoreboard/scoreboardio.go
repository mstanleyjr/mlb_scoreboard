package scoreboard

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

type League struct {
	League    statsapi.LeagueRestObject
	Divisions []statsapi.DivisionRestObject
}

type ScoreboardInformation struct {
	LeagueMap               map[int32]League
	TeamSchedule            *statsapi.ScheduleRestObject
	TodaySchedule           *statsapi.ScheduleRestObject
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
	GameStatus  string
	Venue       string
	GameType    string
	FinalInning int
}

type ScoreboardLastMatchupTeam struct {
	Team     ScoreboardLiveGameTeam
	Winner   bool
	Decision string // Winning or losing pitcher
}

type ScoreboardLiveGame struct {
	AwayTeam              ScoreboardLiveGameTeam
	HomeTeam              ScoreboardLiveGameTeam
	Bases                 ScoreboardLiveGameBases
	Inning                int
	HalfInning            string
	Outs                  int
	Balls                 int
	Strikes               int
	CurrentPitcher        string
	CurrentBatter         string
	CurrentBatterPosition string
	Venue                 string
	GameType              string
	LastPlay              string
	LastPlayRBIs          int
}

type ScoreboardLiveGameTeam struct {
	Name      string
	ShortName string
	Record    ScoreboardWinLossRecord
	Runs      int
	Hits      int
	Errors    int
	LOB       int
}

type ScoreboardLiveGameBases struct {
	First  bool
	Second bool
	Third  bool
}

func BuildLeagueDivisionLookup(leagues statsapi.LeagueResponseObject, divisions statsapi.DivisionsRestObject) map[int32]League {
	lookup := make(map[int32]League)
	for _, league := range leagues.Leagues {
		if *league.Id == AMERICAN_LEAGUE_ID || *league.Id == NATIONAL_LEAGUE_ID {
			lookup[*league.Id] = League{
				League: league}
		}
	}

	for _, division := range *divisions.Divisions {
		leagueID := *division.League.Id
		if league, exists := lookup[leagueID]; exists {
			league.Divisions = append(league.Divisions, division)
			lookup[leagueID] = league
		}
	}
	return lookup
}

func BuildGameTypeLookup(gametypes []statsapi.GameTypeEnum) map[string]string {
	lookup := make(map[string]string)
	for _, gameType := range gametypes {
		lookup[*gameType.Id] = *gameType.Description
	}
	return lookup
}

func IsDataLoaded(info ScoreboardInformation) bool {
	return info.AmericanLeagueStandings != nil || info.NationalLeagueStandings != nil || info.TeamSchedule != nil
}

func FindActiveTeamGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.TeamSchedule == nil || len(*info.TeamSchedule.Dates) == 0 {
		return nil, nil
	}

	for _, date := range *info.TeamSchedule.Dates {
		for _, game := range *date.Games {
			if *game.Status.AbstractGameCode == "L" {
				return &game, nil
			}
		}
	}

	return nil, nil
}

func FindAllActiveGameIds(info ScoreboardInformation) []int32 {
	activeGameIds := make([]int32, 0)
	if info.TodaySchedule == nil || len(*info.TodaySchedule.Dates) == 0 {
		return activeGameIds
	}

	for _, date := range *info.TodaySchedule.Dates {
		for _, game := range *date.Games {
			if *game.Status.AbstractGameCode == "L" {
				activeGameIds = append(activeGameIds, *game.GamePk)
			}
		}
	}

	return activeGameIds
}

func FindNextScheduledGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.TeamSchedule == nil || len(*info.TeamSchedule.Dates) == 0 {
		return nil, nil
	}
	return nextScheduledGame(info)
}

func FindLastCompletedGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.TeamSchedule == nil || len(*info.TeamSchedule.Dates) == 0 {
		return nil, nil
	}
	return previousCompletedGame(info)
}

func DisplayLoop(condCheckInterval time.Duration, displayDuration time.Duration, controller *DisplayController) {
	fmt.Println("DisplayLoop")

	ticker := time.NewTicker(condCheckInterval)
	done := time.After(displayDuration)

	// So I could these functions display and then watch for the interrupt to unlock and allow the other
	count := 0
	for {
		select {
		case <-done:
			ticker.Stop()
			fmt.Print("\n")
			return
		case <-ticker.C:
			controller.Mu.Lock()
			// Check the Cond and unlock and return
			// Then the other can lock and do whatever
			for controller.Paused {
				fmt.Println("Paused, waiting...")
				controller.Cond.Wait()
			}
			count++
			fmt.Print(count)
			controller.Mu.Unlock()
		}
	}
}

func nextScheduledGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.TeamSchedule == nil || len(*info.TeamSchedule.Dates) == 0 {
		return nil, nil
	}

	dates := *info.TeamSchedule.Dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Date.Time.Before(dates[j].Date.Time)
	})

	// So these dates and games are in order, so I can just do the next P game
	for _, date := range *info.TeamSchedule.Dates {
		for _, game := range *date.Games {
			if *game.Status.AbstractGameCode == "P" {
				return &game, nil
			}
		}
	}

	return nil, nil
}

func previousCompletedGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.TeamSchedule == nil || len(*info.TeamSchedule.Dates) == 0 {
		return nil, nil
	}
	dates := *info.TeamSchedule.Dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Date.Time.After(dates[j].Date.Time)
	})

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		println("Error loading location: ", err.Error())
		return nil, err
	}

	currTime := time.Now().In(loc)

	for _, date := range *info.TeamSchedule.Dates {
		dt := date.Date.Time

		if dt.Equal(currTime) || dt.Before(currTime) {
			if len(*date.Games) > 0 {
				// Reverse iterate to find the last completed game
				for i := len(*date.Games) - 1; i >= 0; i-- {
					game := (*date.Games)[i]
					if *game.Status.AbstractGameCode == "F" || *game.Status.AbstractGameCode == "O" {
						return &game, nil
					}
				}
			}
		}
	}

	return nil, nil
}

func pitcherName(pitcher *statsapi.BaseballPersonRestObject) string {
	if pitcher == nil || pitcher.FullName == nil {
		return "TBD"
	}
	return *pitcher.FullName
}

func findPitcherDecision(game statsapi.BaseballGameRestObject, winner bool) string {
	if game.LiveData.Decisions == nil {
		return "TBD"
	}
	if winner {
		if game.LiveData.Decisions.Winner != nil {
			return *game.LiveData.Decisions.Winner.FullName
		}
	} else {
		if game.LiveData.Decisions.Loser != nil {
			return *game.LiveData.Decisions.Loser.FullName
		}
	}

	return "TBD"
}

func getLiveGameInfo(game statsapi.BaseballGameRestObject) (ScoreboardLiveGame, error) {
	if game.LiveData == nil {
		return ScoreboardLiveGame{}, fmt.Errorf("no live data available")
	}

	homeTeam, err := getScoreboardLiveGameTeams(game, true)
	if err != nil {
		return ScoreboardLiveGame{}, err
	}
	awayTeam, err := getScoreboardLiveGameTeams(game, false)
	if err != nil {
		return ScoreboardLiveGame{}, err
	}

	var inning, outs, balls, strikes, batterId, lastPlayRBIs int
	var halfInning, currentPitcher, currentBatter, currentBatterPosition, venue, lastPlay string
	var bases ScoreboardLiveGameBases

	if game.LiveData.Linescore != nil {

		if game.LiveData.Linescore.Outs != nil {
			outs = int(*game.LiveData.Linescore.Outs)
		}
		if game.LiveData.Linescore.Balls != nil {
			balls = int(*game.LiveData.Linescore.Balls)
		}
		if game.LiveData.Linescore.Strikes != nil {
			strikes = int(*game.LiveData.Linescore.Strikes)
		}

		if game.LiveData.Linescore.CurrentInning != nil {
			inning = int(*game.LiveData.Linescore.CurrentInning)
		}

		// TODO: Might be better done in the display logic.
		if game.LiveData.Linescore.InningHalf != nil {
			halfInning = *game.LiveData.Linescore.InningHalf
			if outs == 3 {
				if strings.ToLower(halfInning) == "top" {
					halfInning = "Mid"
				} else if strings.ToLower(halfInning) == "bottom" {
					halfInning = "End"
				}
			}
		}

		bases = ScoreboardLiveGameBases{
			First:  game.LiveData.Linescore.Offense.First != nil,
			Second: game.LiveData.Linescore.Offense.Second != nil,
			Third:  game.LiveData.Linescore.Offense.Third != nil,
		}

		if game.LiveData.Plays.CurrentPlay.Matchup.Batter != nil {
			currentBatter = *game.LiveData.Plays.CurrentPlay.Matchup.Batter.FullName
			batterId = int(*game.LiveData.Plays.CurrentPlay.Matchup.Batter.Id)
			teams := *game.LiveData.Boxscore.Teams
			if teams != nil {
				if strings.ToLower(halfInning) == "top" {
					players := *teams["away"].Players
					idKey := "ID" + fmt.Sprint(batterId)

					person, exists := players[idKey]
					if exists && person.Position != nil {
						currentBatterPosition = *person.Position.Abbreviation
					}
				} else {
					players := *teams["home"].Players
					idKey := "ID" + fmt.Sprint(batterId)

					person, exists := players[idKey]
					if exists && person.Position != nil {
						currentBatterPosition = *person.Position.Abbreviation
					}
				}
			}
		}

		// TODO: Make a mapping of position ids to names so I can just pull that instead of doing this rigamarole every time.
		// Until then just the abbreviation is fine

		if game.LiveData.Plays.CurrentPlay.Matchup.Pitcher != nil {
			currentPitcher = *game.LiveData.Plays.CurrentPlay.Matchup.Pitcher.FullName
		}

		if game.GameData.Venue != nil && game.GameData.Venue.Name != nil {
			venue = *game.GameData.Venue.Name
		}

		if game.LiveData.Plays.CurrentPlay.AtBatIndex != nil {
			if *game.LiveData.Plays.CurrentPlay.AtBatIndex > 0 {
				// The current at bat starts as soon as the last one ends, so we can use the at bat index to know when to update the last play
				lastPlayIndex := int(*game.LiveData.Plays.CurrentPlay.AtBatIndex)
				if game.LiveData.Plays.AllPlays != nil && len(*game.LiveData.Plays.AllPlays) > lastPlayIndex {
					lastPlayData := (*game.LiveData.Plays.AllPlays)[lastPlayIndex]

					if lastPlayData.Result.Event == nil {
						lastPlayIndex = lastPlayIndex - 1
						lastPlayData = (*game.LiveData.Plays.AllPlays)[lastPlayIndex]
					}

					if lastPlayData.Result != nil && lastPlayData.Result.Description != nil {
						lastPlay = *lastPlayData.Result.Description
						if lastPlayData.Result.Rbi != nil {
							lastPlayRBIs = int(*lastPlayData.Result.Rbi)
						}
					}
				}
			}
		}

		// I'll need what base is occupied for the next play and then can do some fun stuff with that, maybe even a little animation for the next play
		// Interesting, we don't have the last play strikeout

		// Kind of excited to deal with the play events later, I'll update that then with a G6-2 kind of thing
		// I htink for that one it's a current play deal when the result pops up
		//Then use the credits
	}

	return ScoreboardLiveGame{
		AwayTeam:              awayTeam,
		HomeTeam:              homeTeam,
		Bases:                 bases,
		Inning:                inning,
		HalfInning:            halfInning,
		Outs:                  outs,
		Balls:                 balls,
		Strikes:               strikes,
		CurrentPitcher:        currentPitcher,
		CurrentBatter:         currentBatter,
		CurrentBatterPosition: currentBatterPosition,
		Venue:                 venue,
		LastPlay:              lastPlay,
		LastPlayRBIs:          lastPlayRBIs,
	}, nil
}

func getScoreboardLiveGameTeams(game statsapi.BaseballGameRestObject, homeTeam bool) (ScoreboardLiveGameTeam, error) {
	if game.GameData == nil || game.LiveData == nil || game.GameData.Teams == nil {
		return ScoreboardLiveGameTeam{}, fmt.Errorf("no game data or live data available")
	}
	teams := *game.GameData.Teams

	teamString := "away"
	if homeTeam {
		teamString = "home"
	}

	teamData := teams[teamString]
	if teamData.TeamName == nil || teamData.Abbreviation == nil {
		return ScoreboardLiveGameTeam{}, fmt.Errorf("incomplete team data for %s", teamString)
	}

	linescore := game.LiveData.Linescore.Teams.Away
	if homeTeam {
		linescore = game.LiveData.Linescore.Teams.Home
	}

	var name string
	if teamData.Name != nil {
		name = *teamData.Name
	}

	var shortName string
	if teamData.Abbreviation != nil {
		shortName = *teamData.Abbreviation
	}

	var record ScoreboardWinLossRecord
	if teamData.Record.LeagueRecord != nil {
		record = ScoreboardWinLossRecord{
			Wins:   int(*teamData.Record.LeagueRecord.Wins),
			Losses: int(*teamData.Record.LeagueRecord.Losses),
		}
	}

	var runs, hits, errors, lob int
	if linescore.Runs != nil {
		runs = int(*linescore.Runs)
	}
	if linescore.Hits != nil {
		hits = int(*linescore.Hits)
	}
	if linescore.Errors != nil {
		errors = int(*linescore.Errors)
	}
	if linescore.LeftOnBase != nil {
		lob = int(*linescore.LeftOnBase)
	}

	return ScoreboardLiveGameTeam{
		Name:      name,
		ShortName: shortName,
		Record:    record,
		Runs:      runs,
		Hits:      hits,
		Errors:    errors,
		LOB:       lob,
	}, nil
}

func getGameTypeFromLookup(lookup string) string {
	switch lookup {
	case "S":
		return "Spring Training"
	case "R":
		return "Regular Season"
	case "F":
		return "Wild Card Game"
	case "D":
		return "Division Series"
	case "L":
		return "League Championship Series"
	case "W":
		return "World Series"
	case "C":
		return "Championship"
	case "N":
		return "Nineteenth Century Series"
	case "P":
		return "Playoffs"
	case "A":
		return "All-Star Game"
	case "I":
		return "Intrasquad"
	case "E":
		return "Exhibition"
	default:
		return ""
	}
}
