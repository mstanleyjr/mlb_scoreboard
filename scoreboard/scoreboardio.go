package scoreboard

import (
	"fmt"
	"sort"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/third-party/oapi/statsapi"
)

type League struct {
	League    statsapi.LeagueRestObject
	Divisions []statsapi.DivisionRestObject
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
	return info.AmericanLeagueStandings != nil || info.NationalLeagueStandings != nil || info.Schedule != nil
}

func FindActiveGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
		return nil, nil
	}

	games, err := todayGames(info)
	if err != nil {
		return nil, err
	}
	if games == nil || len(games) == 0 {
		return nil, nil
	}

	for _, game := range games {
		if *game.Status.AbstractGameCode == "L" {
			return &game, nil
		}
	}

	return nil, nil
}

func FindNextScheduledGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
		return nil, nil
	}
	return nextScheduledGame(info)
}

func FindLastCompletedGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
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

func todayGames(info ScoreboardInformation) ([]statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
		return nil, nil
	}

	currentUTCTime := time.Now().UTC()
	for _, date := range *info.Schedule.Dates {
		dt := date.Date.Time
		// TODO: const the layout
		currentUTCDate := currentUTCTime.Format("2006-01-02")
		dtDate := dt.Format("2006-01-02")

		if currentUTCDate == dtDate {
			return *date.Games, nil
		}
	}

	return nil, nil
}

func nextScheduledGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
		return nil, nil
	}

	dates := *info.Schedule.Dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Date.Time.Before(dates[j].Date.Time)
	})

	currentUTCTime := time.Now().UTC()
	for _, date := range *info.Schedule.Dates {
		dt := date.Date.Time

		// First equal or after current date
		if dt.Equal(currentUTCTime) || dt.After(currentUTCTime) {
			if len(*date.Games) > 0 {
				for _, game := range *date.Games {
					if *game.Status.AbstractGameCode == "P" {
						return &game, nil
					}
				}
			}
		}
	}

	return nil, nil
}

func previousCompletedGame(info ScoreboardInformation) (*statsapi.BaseballScheduleItemRestObject, error) {
	if info.Schedule == nil || len(*info.Schedule.Dates) == 0 {
		return nil, nil
	}
	dates := *info.Schedule.Dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Date.Time.After(dates[j].Date.Time)
	})
	currentUTCTime := time.Now().UTC()

	for _, date := range *info.Schedule.Dates {
		dt := date.Date.Time

		// First before current date
		if dt.Equal(currentUTCTime) || dt.Before(currentUTCTime) {
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
