package scoreboard

import (
	"fmt"
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

func DisplayLoop(condCheckInterval time.Duration, displayDuration time.Duration, controller *DisplayController) {
	fmt.Println("DisplayLoop")

	ticker := time.NewTicker(condCheckInterval)
	done := time.After(displayDuration)

	// So I could these functions display and then watch for the interrupt to unlock and allow the other
	count := 0
	for {
		controller.Mu.Lock()
		select {
		case <-done:
			ticker.Stop()
			return
		case <-ticker.C:
			// Check the Cond and unlock and return
			// Then the other can lock and do whatever
			for controller.Paused {
				fmt.Println("Paused, waiting...")
				controller.Cond.Wait()
			}
			count++
			if count%6 == 0 {
				fmt.Print("\n")
			}
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
