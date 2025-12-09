package main

import (
	"context"
	"fmt"

	"github.com/mstanleyjr/mlb_scoreboard/client/oapi/statsapi"
)

const CLIENT_SERVER = "https://statsapi.mlb.com"
const MLB_SPORTS_ID = 1

func main() {

	println("Starting MLB Scoreboard")
	ctx, _ := context.WithCancel(context.Background())

	mlbClient, err := statsapi.NewClient(CLIENT_SERVER)
	if err != nil {
		println(err.Error())
		return
	}

	schedule, err := mlbClient.Schedule(ctx, &statsapi.ScheduleParams{
		CalendarTypes:        nil,
		EventTypes:           nil,
		ScheduleEventTypes:   nil,
		TeamId:               nil,
		LeagueId:             nil,
		SportId:              MLB_SPORTS_ID,
		GamePk:               nil,
		GamePks:              nil,
		EventIds:             nil,
		VenueIds:             nil,
		PerformerIds:         nil,
		GameTypes:            nil,
		GameType:             nil,
		Season:               nil,
		Seasons:              nil,
		Date:                 nil,
		StartDate:            nil,
		EndDate:              nil,
		Timecode:             nil,
		UseLatestGames:       nil,
		OpponentId:           nil,
		PublicFacing:         nil,
		Fields:               nil,
		UsingPrivateEndpoint: false,
	})
	if err != nil {
		println(err.Error())
		return
	}

	fmt.Println("Schedule retrieved: ", schedule)
	fmt.Println("req", schedule.StatusCode)
	fmt.Println("body", schedule.Body)

	// make a config object
	// make a client
	// Should I bother with tests here?
	// Future me might like it and they won't be that expensive to mock
	// then if I put this up for future stuff it won't be a tragedy.

}
