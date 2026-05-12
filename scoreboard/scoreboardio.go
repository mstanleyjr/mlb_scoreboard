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
	PlayerLookupMap         *map[int32]statsapi.BaseballPersonRestObject
}

type ScoreboardDivision struct {
	LeagueName string
	Teams      []ScoreboardDivisionTeam
}

type ScoreboardDivisionTeam struct {
	Name      string
	ShortName string
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
	ShortName       string
	ProbablePitcher ScoreboardPitcher
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
	AwayTeam               ScoreboardLastMatchupTeam
	HomeTeam               ScoreboardLastMatchupTeam
	DateTime               time.Time
	GameStatus             string
	Venue                  string
	GameType               string
	FinalInning            int
	WinningPitcherLastName string
	LosingPitcherLastName  string
	SavePitcherLastName    string
}

type ScoreboardLastMatchupTeam struct {
	Team   ScoreboardLiveGameTeam
	Winner bool
}

type ScoreboardLiveGame struct {
	AwayTeam         ScoreboardLiveGameTeam
	HomeTeam         ScoreboardLiveGameTeam
	Bases            ScoreboardLiveGameBases
	CurrentBatter    ScoreboardLiveGameBatter
	CurrentPitcher   ScoreboardPitcher
	Inning           int
	HalfInning       string
	Outs             int
	Balls            int
	Strikes          int
	CurrentPitcherId int32
	CurrentBatterId  int32
	Venue            string
	GameType         string
	LastPlayNotation string
	LastPlay         string
	LastPlayRBIs     int
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

type ScoreboardPitcher struct {
	FullName string
	LastName string
	ERA      string
	Hand     string
	Wins     int32
	Losses   int32
	Saves    int32
}

func getPitchHandFromPerson(p *statsapi.BaseballPersonRestObject) string {
	if p != nil && p.PitchHand != nil && p.PitchHand.Code != nil {
		return *p.PitchHand.Code
	}
	return ""
}

type ScoreboardLiveGameBatter struct {
	FullName             string
	LastName             string
	CurrentPosition      string
	SeasonBattingAverage string
	SeasonOPS            string
	GameHits             int32
	GameAtBats           int32
	Summary              string
}

type PitchingDecision string

const (
	PitchingDecisionWin  PitchingDecision = "win"
	PitchingDecisionLoss PitchingDecision = "loss"
	PitchingDecisionSave PitchingDecision = "save"
)

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

func BuildPlayerLookup(players []statsapi.BaseballPersonRestObject) map[int32]statsapi.BaseballPersonRestObject {
	lookup := make(map[int32]statsapi.BaseballPersonRestObject, len(players))
	for _, player := range players {
		if player.Id == nil {
			continue
		}
		lookup[*player.Id] = player
	}
	return lookup
}

func playerLookup(info ScoreboardInformation, playerID *int32) *statsapi.BaseballPersonRestObject {
	if playerID == nil || info.PlayerLookupMap == nil {
		return nil
	}

	player, ok := (*info.PlayerLookupMap)[*playerID]
	if !ok {
		return nil
	}

	return &player
}

func IsDataLoaded(info ScoreboardInformation) bool {
	// Check if basic data is loaded
	hasBasicData := info.AmericanLeagueStandings != nil || info.NationalLeagueStandings != nil || info.TeamSchedule != nil

	// Check if league/divisions data is loaded
	hasLeagueData := len(info.LeagueMap) > 0

	// Check if game types are loaded
	hasGameTypeData := len(info.GameTypeMap) > 0

	return hasBasicData && hasLeagueData && hasGameTypeData
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

// DisplayLoop now handles only timing/pause control.
// Rendering is performed by each display function before entering the loop.
func DisplayLoop(condCheckInterval time.Duration, displayDuration time.Duration, controller *DisplayController) {
	fmt.Println("DisplayLoop")

	ticker := time.NewTicker(condCheckInterval)
	done := time.After(displayDuration)

	count := 0
	for {
		select {
		case <-done:
			ticker.Stop()
			fmt.Print("\n")
			return
		case <-ticker.C:
			controller.Mu.Lock()
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

func findPitcherDecision(info ScoreboardInformation, game statsapi.BaseballGameRestObject, decision PitchingDecision) string {
	if game.LiveData == nil || game.LiveData.Decisions == nil {
		return "TBD"
	}

	var pitcher *statsapi.BaseballPersonRestObject
	switch decision {
	case PitchingDecisionWin:
		pitcher = game.LiveData.Decisions.Winner
	case PitchingDecisionLoss:
		pitcher = game.LiveData.Decisions.Loser
	case PitchingDecisionSave:
		pitcher = game.LiveData.Decisions.Save
	default:
	}

	if pitcher == nil {
		return "TBD"
	}

	if foundPitcher := playerLookup(info, pitcher.Id); foundPitcher != nil && foundPitcher.LastName != nil && *foundPitcher.LastName != "" {
		return *foundPitcher.LastName
	}

	if pitcher.LastName != nil && *pitcher.LastName != "" {
		return *pitcher.LastName
	}

	return "TBD"
}

func scorekeepingLastPlayNotation(play *statsapi.BaseballPlayRestObject) string {
	if play == nil || play.Result == nil {
		return ""
	}

	result := play.Result
	eventType := statsapi.EventType("")
	if result.EventType != nil {
		eventType = statsapi.EventType(strings.ToLower(strings.TrimSpace(*result.EventType)))
	}
	event := ""
	if result.Event != nil {
		event = strings.ToLower(strings.TrimSpace(*result.Event))
	}
	description := ""
	if result.Description != nil {
		description = strings.TrimSpace(*result.Description)
	}

	notation := ""
	switch eventType {
	case statsapi.EventTypeStrikeout, statsapi.EventTypeStrikeoutDoublePlay, statsapi.EventTypeStrikeoutTriplePlay:
		notation = "K"
	case statsapi.EventTypeHomeRun:
		notation = "HR"
	case statsapi.EventTypeGroundedIntoDoublePlay, statsapi.EventTypeDoublePlay:
		if fielding := scorekeepingFieldingNotation(play, "GDP"); fielding != "" {
			notation = fielding
		} else {
			notation = "GDP"
		}
	case statsapi.EventTypeDouble:
		notation = "2B"
	case statsapi.EventTypeSingle:
		notation = "1B"
	case statsapi.EventTypeTriple:
		notation = "3B"
	case statsapi.EventTypeWalk:
		notation = "BB"
	case statsapi.EventTypeHitByPitch:
		notation = "HBP"
	case statsapi.EventTypeSacFly, statsapi.EventTypeSacFlyDoublePlay:
		notation = "SF"
	case statsapi.EventTypeSacBunt, statsapi.EventTypeSacBuntDoublePlay:
		notation = "SH"
	case statsapi.EventTypeStolenBase, statsapi.EventTypeStolenBase2b, statsapi.EventTypeStolenBase3b, statsapi.EventTypeStolenBaseHome:
		notation = "SB"
	case statsapi.EventTypeCaughtStealing, statsapi.EventTypeCaughtStealing2b, statsapi.EventTypeCaughtStealing3b, statsapi.EventTypeCaughtStealingHome:
		notation = "CS"
	}

	if notation != "" {
		return scorekeepingAppendRBIs(notation, result.Rbi)
	}

	switch {
	case strings.Contains(event, "strikeout"):
		notation = "K"
	case strings.Contains(event, "home run"):
		notation = "HR"
	case strings.Contains(event, "double play"):
		if fielding := scorekeepingFieldingNotation(play, "GDP"); fielding != "" {
			notation = fielding
		} else {
			notation = "GDP"
		}
	case strings.Contains(event, "double"):
		notation = "2B"
	case strings.Contains(event, "single"):
		notation = "1B"
	case strings.Contains(event, "triple"):
		notation = "3B"
	case strings.Contains(event, "walk"):
		notation = "BB"
	case strings.Contains(event, "hit by pitch"):
		notation = "HBP"
	case strings.Contains(event, "groundout") || strings.Contains(event, "ground out"):
		if fielding := scorekeepingFieldingNotation(play, "G"); fielding != "" {
			notation = fielding
		} else {
			notation = "GO"
		}
	case strings.Contains(event, "flyout") || strings.Contains(event, "fly out"):
		if fielding := scorekeepingFieldingNotation(play, "F"); fielding != "" {
			notation = fielding
		} else {
			notation = "FO"
		}
	case strings.Contains(event, "lineout") || strings.Contains(event, "line out"):
		if fielding := scorekeepingFieldingNotation(play, "L"); fielding != "" {
			notation = fielding
		} else {
			notation = "LO"
		}
	case strings.Contains(event, "popout") || strings.Contains(event, "pop out"):
		if fielding := scorekeepingFieldingNotation(play, "P"); fielding != "" {
			notation = fielding
		} else {
			notation = "PO"
		}
	case strings.Contains(event, "sac fly"):
		notation = "SF"
	case strings.Contains(event, "sac bunt"):
		notation = "SH"
	case strings.Contains(event, "stolen base"):
		notation = "SB"
	case strings.Contains(event, "caught stealing"):
		notation = "CS"
	case strings.Contains(event, "fielders choice") || strings.Contains(event, "fielder's choice"):
		if fielding := scorekeepingFieldingNotation(play, "FC"); fielding != "" {
			notation = fielding
		} else {
			notation = "FC"
		}
	case strings.Contains(event, "force out") || strings.Contains(event, "forceout"):
		if fielding := scorekeepingFieldingNotation(play, "FO"); fielding != "" {
			notation = fielding
		} else {
			notation = "FO"
		}
	case strings.Contains(event, "other out"):
		if fielding := scorekeepingFieldingNotation(play, "OUT"); fielding != "" {
			notation = fielding
		} else {
			notation = "OUT"
		}
	}

	if notation != "" {
		return scorekeepingAppendRBIs(notation, result.Rbi)
	}

	return description
}

func scorekeepingFieldingNotation(play *statsapi.BaseballPlayRestObject, prefix string) string {
	if play == nil {
		return ""
	}

	positionCode := func(credit *statsapi.PlayCreditRestObject) string {
		if credit == nil || credit.Position == nil {
			return ""
		}
		if credit.Position.Code != nil && strings.TrimSpace(*credit.Position.Code) != "" {
			return strings.TrimSpace(*credit.Position.Code)
		}
		if credit.Position.Abbreviation != nil && strings.TrimSpace(*credit.Position.Abbreviation) != "" {
			return strings.TrimSpace(*credit.Position.Abbreviation)
		}
		return ""
	}

	collect := func(credits *[]statsapi.PlayCreditRestObject) []string {
		if credits == nil {
			return nil
		}
		sequence := make([]string, 0, len(*credits))
		for i := range *credits {
			credit := &(*credits)[i]
			if credit.Credit == nil {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(*credit.Credit)) {
			case "f_assist", "f_putout":
				if pos := positionCode(credit); pos != "" {
					sequence = append(sequence, pos)
				}
			}
		}
		if len(sequence) == 0 {
			return nil
		}
		return sequence
	}

	best := collect(play.Credits)
	if play.Runners != nil {
		for i := range *play.Runners {
			if seq := collect((*play.Runners)[i].Credits); len(seq) > len(best) {
				best = seq
			}
		}
	}

	if len(best) == 0 {
		return ""
	}

	return prefix + strings.Join(best, "-")
}

func scorekeepingAppendRBIs(notation string, rbi *int32) string {
	if notation == "" || rbi == nil || *rbi <= 0 {
		return notation
	}
	return fmt.Sprintf("%s, %d RBI", notation, *rbi)
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

	var inning, outs, balls, strikes, lastPlayRBIs int
	var halfInning, venue, lastPlay, lastPlayNotation string
	var currentPitcherId, currentBatterId int32
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

		if game.LiveData.Plays != nil && game.LiveData.Plays.CurrentPlay != nil && game.LiveData.Plays.CurrentPlay.Matchup != nil {
			if game.LiveData.Plays.CurrentPlay.Matchup.Batter != nil && game.LiveData.Plays.CurrentPlay.Matchup.Batter.Id != nil {
				currentBatterId = *game.LiveData.Plays.CurrentPlay.Matchup.Batter.Id
			}

			if game.LiveData.Plays.CurrentPlay.Matchup.Pitcher != nil && game.LiveData.Plays.CurrentPlay.Matchup.Pitcher.Id != nil {
				currentPitcherId = *game.LiveData.Plays.CurrentPlay.Matchup.Pitcher.Id
			}
		}

		// TODO: Make a mapping of position ids to names so I can just pull that instead of doing this rigamarole every time.
		// Until then just the abbreviation is fine

		if game.GameData.Venue != nil && game.GameData.Venue.Name != nil {
			venue = *game.GameData.Venue.Name
		}

		if game.LiveData.Plays != nil && game.LiveData.Plays.CurrentPlay != nil && game.LiveData.Plays.CurrentPlay.AtBatIndex != nil {
			if *game.LiveData.Plays.CurrentPlay.AtBatIndex > 0 {
				// The current at bat starts as soon as the last one ends, so we can use the at bat index to know when to update the last play
				lastPlayIndex := int(*game.LiveData.Plays.CurrentPlay.AtBatIndex)
				if game.LiveData.Plays.AllPlays != nil && len(*game.LiveData.Plays.AllPlays) > lastPlayIndex {
					lastPlayData := (*game.LiveData.Plays.AllPlays)[lastPlayIndex]

					if lastPlayData.Result == nil || lastPlayData.Result.Event == nil {
						lastPlayIndex = lastPlayIndex - 1
						if lastPlayIndex >= 0 {
							lastPlayData = (*game.LiveData.Plays.AllPlays)[lastPlayIndex]
						}
					}

					if lastPlayData.Result != nil {
						lastPlayNotation = scorekeepingLastPlayNotation(&lastPlayData)
						if lastPlayData.Result.Description != nil {
							lastPlay = *lastPlayData.Result.Description
						}
						if lastPlayData.Result.Rbi != nil {
							lastPlayRBIs = int(*lastPlayData.Result.Rbi)
						}
					}
				}
			}
		}
	}

	return ScoreboardLiveGame{
		AwayTeam:         awayTeam,
		HomeTeam:         homeTeam,
		Bases:            bases,
		Inning:           inning,
		HalfInning:       halfInning,
		Outs:             outs,
		Balls:            balls,
		Strikes:          strikes,
		CurrentPitcherId: currentPitcherId,
		CurrentBatterId:  currentBatterId,
		Venue:            venue,
		LastPlayNotation: lastPlayNotation,
		LastPlay:         lastPlay,
		LastPlayRBIs:     lastPlayRBIs,
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

func getScoreboardLiveGameBatter(info ScoreboardInformation, batterStats statsapi.PlayerStatsResponse, liveGame statsapi.BaseballGameRestObject, gameInfo ScoreboardLiveGame) ScoreboardLiveGameBatter {
	var fullName, lastName, currentBatterPosition, battingAverage, ops, summary string
	var hits, atBats int32

	if liveGame.LiveData != nil && liveGame.LiveData.Plays != nil && liveGame.LiveData.Plays.CurrentPlay != nil && liveGame.LiveData.Plays.CurrentPlay.Matchup != nil && liveGame.LiveData.Plays.CurrentPlay.Matchup.Batter != nil {
		batter := liveGame.LiveData.Plays.CurrentPlay.Matchup.Batter
		batterSource := batter
		if foundBatter := playerLookup(info, batter.Id); foundBatter != nil {
			batterSource = foundBatter
		}
		if batterSource.FullName != nil {
			fullName = *batterSource.FullName
		}
		if batterSource.LastName != nil {
			lastName = *batterSource.LastName
		}

		if liveGame.LiveData.Boxscore != nil && liveGame.LiveData.Boxscore.Teams != nil {
			teams := *liveGame.LiveData.Boxscore.Teams
			var players map[string]statsapi.BaseballRosterEntryRestObject
			if strings.ToLower(gameInfo.HalfInning) == "top" {
				if team, ok := teams["away"]; ok && team.Players != nil {
					players = *team.Players
				}
			} else {
				if team, ok := teams["home"]; ok && team.Players != nil {
					players = *team.Players
				}
			}

			if players != nil {
				idKey := "ID" + fmt.Sprint(gameInfo.CurrentBatterId)
				person, exists := players[idKey]
				if exists {
					if person.Position != nil {
						if person.Position.Abbreviation != nil {
							currentBatterPosition = *person.Position.Abbreviation
						} else if person.Position.Code != nil {
							currentBatterPosition = *person.Position.Code
						}
					}

					if person.Stats != nil && person.Stats.Batting != nil {
						if person.Stats.Batting.Hits != nil {
							hits = *person.Stats.Batting.Hits
						}
						if person.Stats.Batting.AtBats != nil {
							atBats = *person.Stats.Batting.AtBats
						}
						if person.Stats.Batting.Summary != nil {
							summary = *person.Stats.Batting.Summary
						}
					}
				}
			}
		}
	}

	if batterStats.Stats != nil {
		for _, stat := range *batterStats.Stats {
			if stat.Type != nil && stat.Type.DisplayName != nil && strings.ToLower(*stat.Type.DisplayName) == strings.ToLower(string(statsapi.StatTypeSEASON)) {
				if stat.Splits != nil && len(*stat.Splits) > 0 {
					split := (*stat.Splits)[0]
					if split.Stat != nil {
						if split.Stat.Avg != nil {
							battingAverage = *split.Stat.Avg
						}
						if split.Stat.Ops != nil {
							ops = *split.Stat.Ops
						}
					}
				}
			}
		}
	}

	return ScoreboardLiveGameBatter{
		FullName:             fullName,
		LastName:             lastName,
		CurrentPosition:      currentBatterPosition,
		SeasonBattingAverage: battingAverage,
		SeasonOPS:            ops,
		GameHits:             hits,
		GameAtBats:           atBats,
		Summary:              summary,
	}
}

func getScoreboardLiveGamePitcher(info ScoreboardInformation, pitcherStats statsapi.PlayerStatsResponse, liveGame statsapi.BaseballGameRestObject) ScoreboardPitcher {
	var fullName, lastName, throwingHand string

	if liveGame.LiveData != nil && liveGame.LiveData.Plays != nil && liveGame.LiveData.Plays.CurrentPlay != nil && liveGame.LiveData.Plays.CurrentPlay.Matchup != nil {
		if liveGame.LiveData.Plays.CurrentPlay.Matchup.Pitcher != nil {
			pitcher := liveGame.LiveData.Plays.CurrentPlay.Matchup.Pitcher
			pitcherSource := pitcher
			if foundPitcher := playerLookup(info, pitcher.Id); foundPitcher != nil {
				pitcherSource = foundPitcher
			}
			if pitcherSource.FullName != nil {
				fullName = *pitcherSource.FullName
			}
			if pitcherSource.LastName != nil {
				lastName = *pitcherSource.LastName
			}
			throwingHand = getPitchHandFromPerson(pitcherSource)
		}
		if throwingHand == "" && liveGame.LiveData.Plays.CurrentPlay.Matchup.PitchHand != nil && liveGame.LiveData.Plays.CurrentPlay.Matchup.PitchHand.Code != nil {
			throwingHand = *liveGame.LiveData.Plays.CurrentPlay.Matchup.PitchHand.Code
		}
	}

	era, wins, losses, saves := GetScoreboardPitcherStats(pitcherStats)

	return ScoreboardPitcher{
		FullName: fullName,
		LastName: lastName,
		ERA:      era,
		Hand:     throwingHand,
		Wins:     wins,
		Losses:   losses,
		Saves:    saves,
	}
}

func GetScoreboardPitcherStats(pitcherStats statsapi.PlayerStatsResponse) (era string, wins int32, losses int32, saves int32) {
	era = "0.00"
	wins = 0
	losses = 0
	saves = 0

	if pitcherStats.Stats != nil {
		for _, stat := range *pitcherStats.Stats {
			if stat.Type != nil && stat.Type.DisplayName != nil && strings.ToLower(*stat.Type.DisplayName) == strings.ToLower(string(statsapi.StatTypeSEASON)) {
				if stat.Splits != nil && len(*stat.Splits) > 0 {
					split := (*stat.Splits)[0]
					if split.Stat != nil {
						if split.Stat.Era != nil {
							era = *split.Stat.Era
						}
						if split.Stat.Wins != nil {
							wins = *split.Stat.Wins
						}
						if split.Stat.Losses != nil {
							losses = *split.Stat.Losses
						}
						if split.Stat.Saves != nil {
							saves = *split.Stat.Saves
						}
					}
				}
			}
		}
	}

	return era, wins, losses, saves
}
