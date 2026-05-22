package scoreboard

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"
)

const (
	divisionStandingsTopPadding    = 2
	divisionStandingsTitleH        = 8
	divisionStandingsTitleGap      = 6
	divisionStandingsTitleRuleGap  = 3
	divisionStandingsRowBlockH     = 19
	divisionStandingsRowGap        = 1
	divisionStandingsRankX         = 1
	divisionStandingsLineX         = 7
	divisionStandingsTeamX         = 10
	divisionStandingsRecordX       = 10
	divisionStandingsGBRightX      = 63
	divisionStandingsSectionGap    = 4
	divisionStandingsLineGapBottom = 1
	divisionStandingsMaxGBChars    = 4
	divisionStandingsMaxTitleChar  = 10
	divisionStandingsCanvasHeight  = 64
)

var divisionStandingsMonochrome bool
var divisionStandingsGreenBackground bool
var nextMatchupHoldDuration = 2500 * time.Millisecond
var nextMatchupSlideDuration = 500 * time.Millisecond
var lastMatchupHoldDuration = 2500 * time.Millisecond
var lastMatchupSlideDuration = 500 * time.Millisecond
var liveGameHoldDuration = 2200 * time.Millisecond
var liveGameSlideDuration = 400 * time.Millisecond

type divisionStandingsPalette struct {
	background color.RGBA
	title      color.RGBA
	rank       color.RGBA
	team       color.RGBA
	record     color.RGBA
	gamesBack  color.RGBA
	rule       color.RGBA
	line       color.RGBA
}

func SetDivisionStandingsMonochrome(enabled bool) {
	divisionStandingsMonochrome = enabled
}

func SetDivisionStandingsGreenBackground(enabled bool) {
	divisionStandingsGreenBackground = enabled
}

func SetNextMatchupTiming(holdMS, slideMS int) {
	if holdMS <= 0 {
		holdMS = 2500
	}
	if slideMS <= 0 {
		slideMS = 500
	}
	nextMatchupHoldDuration = time.Duration(holdMS) * time.Millisecond
	nextMatchupSlideDuration = time.Duration(slideMS) * time.Millisecond
}

func SetLastMatchupTiming(holdMS, slideMS int) {
	if holdMS <= 0 {
		holdMS = 2500
	}
	if slideMS <= 0 {
		slideMS = 500
	}
	lastMatchupHoldDuration = time.Duration(holdMS) * time.Millisecond
	lastMatchupSlideDuration = time.Duration(slideMS) * time.Millisecond
}

func SetLiveGameTiming(holdMS, slideMS int) {
	if holdMS <= 0 {
		holdMS = 2200
	}
	if slideMS <= 0 {
		slideMS = 400
	}
	liveGameHoldDuration = time.Duration(holdMS) * time.Millisecond
	liveGameSlideDuration = time.Duration(slideMS) * time.Millisecond
}

func divisionStandingsColors() divisionStandingsPalette {
	background := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if divisionStandingsGreenBackground {
		background = color.RGBA{R: 22, G: 67, B: 22, A: 255}
	}

	if divisionStandingsMonochrome {
		lightGray := color.RGBA{R: 220, G: 220, B: 220, A: 255}
		return divisionStandingsPalette{
			background: background,
			title:      lightGray,
			rank:       lightGray,
			team:       lightGray,
			record:     lightGray,
			gamesBack:  lightGray,
			rule:       lightGray,
			line:       lightGray,
		}
	}

	return divisionStandingsPalette{
		background: background,
		title:      color.RGBA{R: 255, G: 255, B: 0, A: 255},
		rank:       color.RGBA{R: 180, G: 180, B: 180, A: 255},
		team:       color.RGBA{},
		record:     color.RGBA{R: 120, G: 180, B: 255, A: 255},
		gamesBack:  color.RGBA{R: 100, G: 200, B: 100, A: 255},
		rule:       color.RGBA{R: 70, G: 70, B: 70, A: 255},
		line:       color.RGBA{R: 90, G: 90, B: 90, A: 255},
	}
}

// PixelCanvas is the minimal surface needed by scoreboard renderers.
// Both rgbmatrix.Canvas and the local terminal sim canvas can satisfy this.
type PixelCanvas interface {
	Set(x, y int, c color.Color)
	Bounds() image.Rectangle
}

// DrawDivisionStandings renders division standings to the LED matrix
func DrawDivisionStandings(c PixelCanvas, division ScoreboardDivision) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	palette := divisionStandingsColors()

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, palette.background)
		}
	}

	contentY := divisionStandingsTopPadding - division.ScrollOffset
	for sectionIndex, section := range division.Sections {
		titleY := contentY
		title := trimToChars(strings.ToUpper(section.Title), divisionStandingsMaxTitleChar)
		if titleY < height && titleY+fontH5x8 >= 0 {
			DrawText5x8Centered(c, titleY, title, palette.title)
		}
		titleRuleY := titleY + divisionStandingsTitleH + divisionStandingsTitleRuleGap
		if titleRuleY >= 0 && titleRuleY < height {
			for x := 1; x < width-1; x++ {
				c.Set(x, titleRuleY, palette.rule)
			}
		}

		contentY += divisionStandingsTitleH + divisionStandingsTitleGap
		for teamIndex, team := range section.Teams {
			nameY := contentY
			statsY := nameY + 10
			separatorY := nameY + divisionStandingsRowBlockH - 1
			if nameY < height && statsY+fontH5x8 >= 0 {
				rank := trimToChars(fmt.Sprintf("%d", team.Rank), 2)
				name := trimText5x8ToWidth(strings.ToUpper(chooseStringValue(team.ShortName, team.Name)), width-divisionStandingsTeamX-1)
				record := trimToChars(fmt.Sprintf("%d-%d", team.Record.Wins, team.Record.Losses), 6)
				gb := trimToChars(formatDivisionGamesBack(team.GamesBack), divisionStandingsMaxGBChars)
				teamColor := GetTeamColor(team.Name)
				if divisionStandingsMonochrome {
					teamColor = palette.team
				}

				DrawText5x8(c, divisionStandingsRankX, nameY, rank, palette.rank)
				DrawText5x8(c, divisionStandingsTeamX, nameY, name, teamColor)
				DrawText5x8(c, divisionStandingsRecordX, statsY, record, palette.record)
				drawText5x8RightAligned(c, divisionStandingsGBRightX, statsY, gb, palette.gamesBack)

				lineTop := nameY - divisionStandingsRowGap
				if teamIndex == 0 {
					lineTop = titleRuleY + 1
				}
				lineBottom := statsY + fontH5x8 - divisionStandingsLineGapBottom
				for y := lineTop; y <= lineBottom; y++ {
					if y >= 0 && y < height {
						c.Set(divisionStandingsLineX, y, palette.line)
					}
				}
			}

			if (teamIndex < len(section.Teams)-1 || sectionIndex < len(division.Sections)-1) && separatorY >= 0 && separatorY < height {
				for x := 1; x < width-1; x++ {
					c.Set(x, separatorY, palette.rule)
				}
			}

			contentY += divisionStandingsRowBlockH
			if teamIndex < len(section.Teams)-1 {
				contentY += divisionStandingsRowGap
			}
		}

		contentY += divisionStandingsSectionGap
	}
}

func formatDivisionGamesBack(gamesBack string) string {
	gb := strings.TrimSpace(gamesBack)
	switch gb {
	case "", "-", "0", "0.0":
		return "-"
	}
	gb = strings.TrimSuffix(gb, ".0")
	return gb
}

func divisionStandingsMaxScrollOffset(displayInfo ScoreboardDivision) int {
	contentHeight := divisionStandingsContentHeight(displayInfo)
	if contentHeight <= divisionStandingsCanvasHeight {
		return 0
	}
	return contentHeight - divisionStandingsCanvasHeight
}

func divisionStandingsContentHeight(displayInfo ScoreboardDivision) int {
	if len(displayInfo.Sections) == 0 {
		return divisionStandingsCanvasHeight
	}

	height := divisionStandingsTopPadding
	for sectionIndex, section := range displayInfo.Sections {
		height += divisionStandingsTitleH + divisionStandingsTitleGap
		height += len(section.Teams) * divisionStandingsRowBlockH
		if len(section.Teams) > 1 {
			height += (len(section.Teams) - 1) * divisionStandingsRowGap
		}
		if sectionIndex < len(displayInfo.Sections)-1 {
			height += divisionStandingsSectionGap
		}
	}
	return height
}

func DrawNextMatchup(c PixelCanvas, m ScoreboardNextMatchup) {
	DrawNextMatchupFrame(c, ScoreboardNextMatchupFrame{
		Matchup:        m,
		PanelIndex:     0,
		NextPanelIndex: -1,
		SlideOffset:    0,
	})
}

func DrawNextMatchupFrame(c PixelCanvas, frame ScoreboardNextMatchupFrame) {
	drawNextMatchup(c, frame)
}

func drawNextMatchup(c PixelCanvas, frame ScoreboardNextMatchupFrame) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	background := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, background)
		}
	}

	drawNextMatchupPanel(c, frame.Matchup, frame.PanelIndex, -frame.SlideOffset)
	if frame.NextPanelIndex >= 0 {
		drawNextMatchupPanel(c, frame.Matchup, frame.NextPanelIndex, width-frame.SlideOffset)
	}
}

func drawNextMatchupPanel(c PixelCanvas, m ScoreboardNextMatchup, panelIndex int, xOffset int) {
	if panelIndex < 0 || panelIndex >= nextMatchupPanelCount() {
		return
	}

	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 160, G: 160, B: 160, A: 255}
	green := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	awayAccent := color.RGBA{R: 120, G: 180, B: 255, A: 255}
	homeAccent := color.RGBA{R: 120, G: 255, B: 140, A: 255}

	awayAbbr := teamAbbrev(m.AwayTeam.ShortName, m.AwayTeam.Name)
	homeAbbr := teamAbbrev(m.HomeTeam.ShortName, m.HomeTeam.Name)
	awayCol := GetTeamColor(m.AwayTeam.Name)
	homeCol := GetTeamColor(m.HomeTeam.Name)

	switch panelIndex {
	case 0:
		header := "NEXT"
		if m.GameType != "" && m.GameType != "Regular Season" {
			header = trimText5x8ToWidth(strings.ToUpper("NEXT "+m.GameType), c.Bounds().Max.X-2)
		}
		drawText5x8CenteredAtOffset(c, xOffset, 1, header, yellow)
		drawText5x8CenteredAtOffset(c, xOffset, 13, formatNextMatchupDateLine(m.DateTime), green)
		drawText5x8CenteredAtOffset(c, xOffset, 23, formatNextMatchupTimeLine(m.DateTime), green)
		drawText5x8CenteredSegmentsAtOffset(c, xOffset, 35, []text5x8Segment{
			{text: awayAbbr, col: awayCol},
			{text: " @ ", col: white},
			{text: homeAbbr, col: homeCol},
		})

		venueLines := formatVenueLines(strings.ToUpper(m.Venue), 12, 2)
		if len(venueLines) == 1 {
			drawText5x8CenteredAtOffset(c, xOffset, 54, venueLines[0], grey)
		} else if len(venueLines) >= 2 {
			drawText5x8CenteredAtOffset(c, xOffset, 47, venueLines[0], grey)
			drawText5x8CenteredAtOffset(c, xOffset, 56, venueLines[1], grey)
		}
	case 1:
		drawText5x8CenteredAtOffset(c, xOffset, 1, "AWAY", awayAccent)
		drawText5x8CenteredSegmentsAtOffset(c, xOffset, 16, []text5x8Segment{
			{text: awayAbbr, col: awayCol},
			{text: " " + strings.ToUpper(nextMatchupTeamRecordValue(m.AwayTeam)), col: white},
		})
		drawText5x8CenteredAtOffset(c, xOffset, 34, strings.ToUpper(nextMatchupPitcherNameLine(m.AwayTeam.ProbablePitcher)), white)
		drawText5x8CenteredAtOffset(c, xOffset, 47, strings.ToUpper(nextMatchupPitcherDetailLine(m.AwayTeam.ProbablePitcher)), grey)
	case 2:
		drawText5x8CenteredAtOffset(c, xOffset, 1, "HOME", homeAccent)
		drawText5x8CenteredSegmentsAtOffset(c, xOffset, 16, []text5x8Segment{
			{text: homeAbbr, col: homeCol},
			{text: " " + strings.ToUpper(nextMatchupTeamRecordValue(m.HomeTeam)), col: white},
		})
		drawText5x8CenteredAtOffset(c, xOffset, 34, strings.ToUpper(nextMatchupPitcherNameLine(m.HomeTeam.ProbablePitcher)), white)
		drawText5x8CenteredAtOffset(c, xOffset, 47, strings.ToUpper(nextMatchupPitcherDetailLine(m.HomeTeam.ProbablePitcher)), grey)
	}
}

// formatNextMatchupDateTime formats local time as "Fri 4/4 6:05p".
func formatNextMatchupDateTime(t time.Time) string {
	if t.IsZero() {
		return "TBD"
	}
	lt := t.Local()
	h := lt.Hour()
	m := lt.Minute()
	suf := "a"
	if h >= 12 {
		suf = "p"
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	day := lt.Weekday().String()[:3]
	return fmt.Sprintf("%s %d/%d %d:%02d%s", day, int(lt.Month()), lt.Day(), h, m, suf)
}

func formatNextMatchupDateLine(t time.Time) string {
	if t.IsZero() {
		return "TBD"
	}
	lt := t.Local()
	day := strings.ToUpper(lt.Weekday().String()[:3])
	return fmt.Sprintf("%s %d/%d", day, int(lt.Month()), lt.Day())
}

func formatNextMatchupTimeLine(t time.Time) string {
	if t.IsZero() {
		return "TBD"
	}
	lt := t.Local()
	h := lt.Hour()
	m := lt.Minute()
	suf := "A"
	if h >= 12 {
		suf = "P"
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	return fmt.Sprintf("%d:%02d%s", h, m, suf)
}

func nextMatchupPanelCount() int {
	return 3
}

func nextMatchupMatchupLine(m ScoreboardNextMatchup) string {
	return trimText5x8ToWidth(teamAbbrev(m.AwayTeam.ShortName, m.AwayTeam.Name)+" @ "+teamAbbrev(m.HomeTeam.ShortName, m.HomeTeam.Name), 62)
}

func nextMatchupTeamRecordLine(team ScoreboardNextMatchupTeam) string {
	line := fmt.Sprintf("%s %s", teamAbbrev(team.ShortName, team.Name), nextMatchupTeamRecordValue(team))
	return trimText5x8ToWidth(line, 62)
}

func nextMatchupTeamRecordValue(team ScoreboardNextMatchupTeam) string {
	return fmt.Sprintf("%d-%d", team.Record.Wins, team.Record.Losses)
}

func nextMatchupPitcherNameLine(p ScoreboardPitcher) string {
	return trimText5x8ToWidth(pitcherNameOnly(p), 62)
}

func nextMatchupPitcherDetailLine(p ScoreboardPitcher) string {
	hand := pitcherHandLabel(p)
	era := strings.TrimSpace(p.ERA)
	switch {
	case hand != "" && era != "":
		return trimText5x8ToWidth(hand+" "+era, 62)
	case hand != "":
		return hand
	case era != "":
		return trimText5x8ToWidth(era, 62)
	default:
		return "TBD"
	}
}

func DrawLastMatchup(c PixelCanvas, m ScoreboardLastMatchup) {
	DrawLastMatchupFrame(c, ScoreboardLastMatchupFrame{
		Matchup:        m,
		PanelIndex:     0,
		NextPanelIndex: -1,
		SlideOffset:    0,
	})
}

func DrawLastMatchupFrame(c PixelCanvas, frame ScoreboardLastMatchupFrame) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	drawLastMatchupStaticTop(c, frame.Matchup)
	drawLastMatchupBottomPanel(c, frame.Matchup, frame.PanelIndex, -frame.SlideOffset)
	if frame.NextPanelIndex >= 0 {
		drawLastMatchupBottomPanel(c, frame.Matchup, frame.NextPanelIndex, width-frame.SlideOffset)
	}
}

func lastMatchupPanelCount() int {
	return 3
}

func drawLastMatchupStaticTop(c PixelCanvas, m ScoreboardLastMatchup) {
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	awayAbbr := teamAbbrev(m.AwayTeam.Team.ShortName, m.AwayTeam.Team.Name)
	homeAbbr := teamAbbrev(m.HomeTeam.Team.ShortName, m.HomeTeam.Team.Name)
	awayCol := GetTeamColor(m.AwayTeam.Team.Name)
	homeCol := GetTeamColor(m.HomeTeam.Team.Name)
	drawText5x8CenteredSegmentsAtOffset(c, 0, 3, []text5x8Segment{
		{text: awayAbbr, col: awayCol},
		{text: " @ ", col: white},
		{text: homeAbbr, col: homeCol},
	})
	drawText5x8CenteredSegmentsAtOffset(c, 0, 14, []text5x8Segment{
		{text: fmt.Sprintf("%d", m.AwayTeam.Team.Runs), col: white},
		{text: " - ", col: white},
		{text: fmt.Sprintf("%d", m.HomeTeam.Team.Runs), col: white},
		{text: "  ", col: white},
		{text: lastMatchupStatusTag(m), col: yellow},
	})
}

func lastMatchupStatusTag(m ScoreboardLastMatchup) string {
	if m.FinalInning > 9 {
		return fmt.Sprintf("F/%d", m.FinalInning)
	}
	return "F"
}

func lastMatchupDateLine(m ScoreboardLastMatchup) string {
	if m.DateTime.IsZero() {
		return ""
	}
	lt := m.DateTime.Local()
	return fmt.Sprintf("%d/%d/%02d", int(lt.Month()), lt.Day(), lt.Year()%100)
}

func drawLastMatchupBottomPanel(c PixelCanvas, m ScoreboardLastMatchup, panelIndex int, xOffset int) {
	if panelIndex < 0 || panelIndex >= lastMatchupPanelCount() {
		return
	}

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	red := color.RGBA{R: 220, G: 60, B: 60, A: 255}
	green := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	awayCol := GetTeamColor(m.AwayTeam.Team.Name)
	homeCol := GetTeamColor(m.HomeTeam.Team.Name)

	switch panelIndex {
	case 0:
		drawText5x8CenteredInRangeAtOffset(c, 0, 32, xOffset, 28, fmt.Sprintf("%d-%d", m.AwayTeam.Team.Record.Wins, m.AwayTeam.Team.Record.Losses), white)
		drawText5x8CenteredInRangeAtOffset(c, 32, 32, xOffset, 28, fmt.Sprintf("%d-%d", m.HomeTeam.Team.Record.Wins, m.HomeTeam.Team.Record.Losses), white)
		if line := lastMatchupDateLine(m); line != "" {
			drawText5x8CenteredAtOffset(c, xOffset, 37, line, white)
		}
		venueLines := formatVenueLines(strings.ToUpper(m.Venue), 12, 2)
		if len(venueLines) == 1 {
			drawText5x8CenteredAtOffset(c, xOffset, 50, venueLines[0], grey)
		} else if len(venueLines) >= 2 {
			drawText5x8CenteredAtOffset(c, xOffset, 46, venueLines[0], grey)
			drawText5x8CenteredAtOffset(c, xOffset, 55, venueLines[1], grey)
		}
	case 1:
		drawLastMatchupStatPairRow(c, xOffset, 28, "R", fmt.Sprintf("%d", m.AwayTeam.Team.Runs), fmt.Sprintf("%d", m.HomeTeam.Team.Runs), grey, awayCol, homeCol)
		drawLastMatchupStatPairRow(c, xOffset, 37, "H", fmt.Sprintf("%d", m.AwayTeam.Team.Hits), fmt.Sprintf("%d", m.HomeTeam.Team.Hits), grey, awayCol, homeCol)
		drawLastMatchupStatPairRow(c, xOffset, 46, "E", fmt.Sprintf("%d", m.AwayTeam.Team.Errors), fmt.Sprintf("%d", m.HomeTeam.Team.Errors), grey, awayCol, homeCol)
		drawLastMatchupStatPairRow(c, xOffset, 55, "L", fmt.Sprintf("%d", m.AwayTeam.Team.LOB), fmt.Sprintf("%d", m.HomeTeam.Team.LOB), grey, awayCol, homeCol)
	case 2:
		lines := []struct {
			text string
			col  color.RGBA
		}{
			{text: trimText5x8ToWidth(strings.ToUpper("W:"+m.WinningPitcherLastName), 62), col: green},
			{text: trimText5x8ToWidth(strings.ToUpper("L:"+m.LosingPitcherLastName), 62), col: red},
		}
		if m.SavePitcherLastName != "" && m.SavePitcherLastName != "TBD" {
			lines = append(lines, struct {
				text string
				col  color.RGBA
			}{
				text: trimText5x8ToWidth(strings.ToUpper("S:"+m.SavePitcherLastName), 62),
				col:  white,
			})
		}
		startY := 28
		if len(lines) == 2 {
			startY = 33
		}
		for i, line := range lines {
			drawText5x8CenteredAtOffset(c, xOffset, startY+i*9, line.text, line.col)
		}
	}
}

func drawLastMatchupStatPairRow(c PixelCanvas, xOffset, y int, label, awayVal, homeVal string, labelCol, awayCol, homeCol color.RGBA) {
	drawText5x8CenteredInRangeAtOffset(c, 0, 16, xOffset, y, label, labelCol)
	drawText5x8CenteredInRangeAtOffset(c, 16, 24, xOffset, y, awayVal, awayCol)
	drawText5x8CenteredInRangeAtOffset(c, 40, 24, xOffset, y, homeVal, homeCol)
}

// teamAbbrev chooses an abbreviation suitable for 3-char scoreboard rows.
func teamAbbrev(shortName, fullName string) string {
	if shortName != "" {
		return trimToChars(shortName, 3)
	}
	return trimToChars(fullName, 3)
}

func drawText3x4CenteredInRange(c PixelCanvas, xStart, regionW, y int, text string, col color.RGBA) {
	tw := len(text) * fontAdv3x4
	x := xStart + (regionW-tw)/2
	if x < xStart {
		x = xStart
	}
	DrawText3x4(c, x, y, text, col)
}

func DrawText3x4CenteredAtOffset(c PixelCanvas, xOffset, y int, text string, col color.RGBA) {
	panelWidth := c.Bounds().Max.X
	tw := len(text) * fontAdv3x4
	x := xOffset + (panelWidth-tw)/2
	if x < xOffset {
		x = xOffset
	}
	DrawText3x4(c, x, y, text, col)
}

func drawText3x4CenteredInRangeAtOffset(c PixelCanvas, xStart, regionW, xOffset, y int, text string, col color.RGBA) {
	drawText3x4CenteredInRange(c, xOffset+xStart, regionW, y, text, col)
}

func drawText3x4RightAligned(c PixelCanvas, rightX, y int, text string, col color.RGBA) {
	x := rightX - (len(text) * fontAdv3x4)
	if x < 0 {
		x = 0
	}
	DrawText3x4(c, x, y, text, col)
}

func drawText5x8RightAligned(c PixelCanvas, rightX, y int, text string, col color.RGBA) {
	x := rightX - measureText5x8Width(text)
	if x < 0 {
		x = 0
	}
	DrawText5x8(c, x, y, text, col)
}

func drawText5x8CenteredAtOffset(c PixelCanvas, xOffset, y int, text string, col color.RGBA) {
	panelWidth := c.Bounds().Max.X
	x := xOffset + (panelWidth-measureText5x8Width(text))/2
	if x < xOffset {
		x = xOffset
	}
	DrawText5x8(c, x, y, text, col)
}

func drawText5x8CenteredInRangeAtOffset(c PixelCanvas, xStart, regionW, xOffset, y int, text string, col color.RGBA) {
	x := xOffset + xStart + (regionW-measureText5x8Width(text))/2
	if x < xOffset+xStart {
		x = xOffset + xStart
	}
	DrawText5x8(c, x, y, text, col)
}

func drawText5x8CenteredSegmentsInRangeAtOffset(c PixelCanvas, xStart, regionW, xOffset, y int, segments []text5x8Segment) {
	totalWidth := measureText5x8SegmentsWidth(segments)
	x := xOffset + xStart + (regionW-totalWidth)/2
	if x < xOffset+xStart {
		x = xOffset + xStart
	}

	type segmentRune struct {
		ch  rune
		col color.RGBA
	}

	var runes []segmentRune
	for _, segment := range segments {
		for _, ch := range []rune(segment.text) {
			runes = append(runes, segmentRune{ch: ch, col: segment.col})
		}
	}
	for i, r := range runes {
		DrawText5x8(c, x, y, string(r.ch), r.col)
		x += advance5x8(r.ch, i == len(runes)-1)
	}
}

type text5x8Segment struct {
	text string
	col  color.RGBA
}

func drawText5x8CenteredSegmentsAtOffset(c PixelCanvas, xOffset, y int, segments []text5x8Segment) {
	panelWidth := c.Bounds().Max.X
	totalWidth := measureText5x8SegmentsWidth(segments)
	x := xOffset + (panelWidth-totalWidth)/2
	if x < xOffset {
		x = xOffset
	}

	type segmentRune struct {
		ch  rune
		col color.RGBA
	}

	var runes []segmentRune
	for _, segment := range segments {
		for _, ch := range []rune(segment.text) {
			runes = append(runes, segmentRune{ch: ch, col: segment.col})
		}
	}
	for i, r := range runes {
		DrawText5x8(c, x, y, string(r.ch), r.col)
		x += advance5x8(r.ch, i == len(runes)-1)
	}
}

func measureText5x8SegmentsWidth(segments []text5x8Segment) int {
	var runes []rune
	for _, segment := range segments {
		runes = append(runes, []rune(segment.text)...)
	}
	width := 0
	for i, ch := range runes {
		width += advance5x8(ch, i == len(runes)-1)
	}
	return width
}

func drawTextRightAligned(c PixelCanvas, rightX, y int, text string, col color.RGBA) {
	x := rightX - (len(text) * fontAdv)
	if x < 0 {
		x = 0
	}
	DrawTextSmall(c, x, y, text, col)
}

func chooseStringValue(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func drawOutCircle(c PixelCanvas, x, y int, filled bool, outlineCol, fillCol color.RGBA) {
	outline := []string{
		"  ###  ",
		" ## ## ",
		"#     #",
		"#     #",
		"#     #",
		" ## ## ",
		"  ###  ",
	}
	filledPat := []string{
		"  ###  ",
		" ##### ",
		"#######",
		"#######",
		"#######",
		"#######",
		" ##### ",
		"  ###  ",
	}
	pat := outline
	col := outlineCol
	if filled {
		pat = filledPat
		col = fillCol
	}
	for py, row := range pat {
		for px, ch := range row {
			if ch == '#' {
				c.Set(x+px, y+py, col)
			}
		}
	}
}

func drawOutsCountCentered(c PixelCanvas, y int, outs int, countText string, outlineCol, fillCol, textCol color.RGBA) {
	if outs < 0 {
		outs = 0
	}
	if outs > 2 {
		outs = 2
	}
	circleW := 7
	gap := 3
	textW := len(countText) * fontAdv3x4
	textGap := 8
	totalW := circleW*2 + gap + textGap + textW
	x := (c.Bounds().Max.X - totalW) / 2
	if x < 0 {
		x = 0
	}
	drawOutCircle(c, x, y, outs >= 1, outlineCol, fillCol)
	drawOutCircle(c, x+circleW+gap, y, outs >= 2, outlineCol, fillCol)
	drawOutCircle(c, x-circleW-gap, y, outs >= 3, outlineCol, fillCol)

	DrawText3x4(c, x+circleW*2+gap+textGap, y, countText, textCol)
}

func drawBallsStrikesCountCentered(c PixelCanvas, y int, balls int, strikes int, labelCol, ballOutlineCol, ballFillCol, strikeOutlineCol, strikeFillCol color.RGBA) {
	if balls < 0 {
		balls = 0
	}
	if balls > 3 {
		balls = 3
	}
	if strikes < 0 {
		strikes = 0
	}
	if strikes > 2 {
		strikes = 2
	}

	circleW := 7
	labelGap := 1
	groupGap := 4
	ballGap := 2
	strikeGap := 2
	ballsW := fontAdv3x4 + labelGap + (circleW * 3) + (ballGap * 2)
	strikesW := fontAdv3x4 + labelGap + (circleW * 2) + strikeGap
	totalW := ballsW + groupGap + strikesW
	x := (c.Bounds().Max.X - totalW) / 2
	if x < 0 {
		x = 0
	}

	DrawText3x5(c, x, y, "B", labelCol)
	ballsStart := x + fontAdv3x4 + labelGap
	for i := 0; i < 3; i++ {
		drawOutCircle(c, ballsStart+i*(circleW+ballGap), y, i < balls, ballOutlineCol, ballFillCol)
	}

	strikesStart := ballsStart + (circleW * 3) + (ballGap * 2) + groupGap
	DrawText3x5(c, strikesStart, y, "S", labelCol)
	strikesStart += fontAdv3x4 + labelGap
	for i := 0; i < 2; i++ {
		drawOutCircle(c, strikesStart+i*(circleW+strikeGap), y, i < strikes, strikeOutlineCol, strikeFillCol)
	}
}

func drawInningOutsCentered(c PixelCanvas, xStart, regionW, y int, inningText string, outs int, textCol, outlineCol, fillCol color.RGBA) {
	if outs < 0 {
		outs = 0
	}
	if outs > 3 {
		outs = 3
	}

	circleW := 7
	gap := 2
	textGap := 5
	textW := len(inningText) * fontAdv
	totalW := textW + textGap + (circleW * 3) + (gap * 2)
	x := xStart + (regionW-totalW)/2
	if x < xStart {
		x = xStart
	}

	DrawText(c, x, y, inningText, textCol)
	outsStart := x + textW + textGap
	for i := 0; i < 3; i++ {
		drawOutCircle(c, outsStart+i*(circleW+gap), y, i < outs, outlineCol, fillCol)
	}
}

// trimToChars truncates a string to n bytes (ASCII-safe for this app data).
func trimToChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// drawTextCentered renders text centered horizontally on the canvas.
func drawTextCentered(c PixelCanvas, y int, text string, col color.RGBA) {
	w := c.Bounds().Max.X
	tw := len(text) * fontAdv
	x := (w - tw) / 2
	if x < 0 {
		x = 0
	}
	DrawText(c, x, y, text, col)
}

// pitcherLine returns a compact pitcher summary, max 16 chars.
func pitcherLine(p ScoreboardPitcher) string {
	if p.FullName == "" || p.FullName == "TBD" {
		return "TBD"
	}
	last := p.LastName
	hand := p.Hand
	if hand == "" {
		hand = "?"
	}
	line := fmt.Sprintf("%s %sHP %s", last, hand, p.ERA)
	return trimToChars(line, 16)
}

// formatGameTime formats a local game time as "Mon 7:05p".
func formatGameTime(t time.Time) string {
	if t.IsZero() {
		return "TBD"
	}
	lt := t.Local()
	h := lt.Hour()
	m := lt.Minute()
	suf := "a"
	if h >= 12 {
		suf = "p"
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	day := lt.Weekday().String()[:3]
	return fmt.Sprintf("%s %d:%02d%s", day, h, m, suf)
}

// pitcherNameOnly returns a compact pitcher display name for matchup rows.
// Current behavior: surname only, or TBD when unknown.
func pitcherNameOnly(p ScoreboardPitcher) string {
	if p.FullName == "" || p.FullName == "TBD" {
		return "TBD"
	}
	if strings.TrimSpace(p.LastName) == "" {
		return p.FullName
	}
	return p.LastName
}

func pitcherHandLabel(p ScoreboardPitcher) string {
	switch strings.ToUpper(strings.TrimSpace(p.Hand)) {
	case "L":
		return "LHP"
	case "R":
		return "RHP"
	default:
		return ""
	}
}

func DrawLiveGameScore(c PixelCanvas, game ScoreboardLiveGame) {
	DrawLiveGameFrame(c, liveGameFrameAt(game, time.Now()))
}

func DrawLiveGameFrame(c PixelCanvas, frame ScoreboardLiveGameFrame) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	drawLiveGameStaticTop(c, frame.Game)
	if frame.NextPanelIndex < 0 {
		drawLiveGameBottomPanelFrame(c, frame.Game, frame.PanelIndex, 0)
	} else {
		drawLiveGameBottomPanelFrame(c, frame.Game, frame.PanelIndex, -frame.SlideOffset)
		drawLiveGameBottomPanelFrame(c, frame.Game, frame.NextPanelIndex, width-frame.SlideOffset)
	}
}

func drawLiveGameBottomPanelFrame(c PixelCanvas, game ScoreboardLiveGame, panelIndex int, xOffset int) {
	if panelIndex == 3 {
		drawLiveGameLastPlayPanelBuffered(c, game, xOffset)
		return
	}
	drawLiveGameBottomPanel(c, game, panelIndex, xOffset)
}

func liveGameFrameAt(game ScoreboardLiveGame, now time.Time) ScoreboardLiveGameFrame {
	panelCount := liveGamePanelCount()
	if panelCount == 0 {
		return ScoreboardLiveGameFrame{Game: game, PanelIndex: 0, NextPanelIndex: -1}
	}

	phaseDuration := liveGameHoldDuration + liveGameSlideDuration
	if phaseDuration <= 0 {
		return ScoreboardLiveGameFrame{Game: game, PanelIndex: 0, NextPanelIndex: -1}
	}

	elapsed := time.Duration(now.UnixNano())
	if elapsed < 0 {
		elapsed = -elapsed
	}
	cycleDuration := time.Duration(panelCount) * phaseDuration
	if cycleDuration <= 0 {
		return ScoreboardLiveGameFrame{Game: game, PanelIndex: 0, NextPanelIndex: -1}
	}
	elapsed %= cycleDuration

	panelIndex := int(elapsed / phaseDuration)
	phaseOffset := elapsed % phaseDuration
	if phaseOffset < liveGameHoldDuration {
		return ScoreboardLiveGameFrame{Game: game, PanelIndex: panelIndex, NextPanelIndex: -1}
	}

	nextPanelIndex := (panelIndex + 1) % panelCount
	return ScoreboardLiveGameFrame{
		Game:           game,
		PanelIndex:     panelIndex,
		NextPanelIndex: nextPanelIndex,
		SlideOffset:    liveGameSlideOffset(phaseOffset-liveGameHoldDuration, liveGameSlideDuration, 64),
	}
}

func liveGameSlideOffset(progress, duration time.Duration, distance int) int {
	if distance <= 0 {
		return 0
	}
	if duration <= 0 || progress >= duration {
		return distance
	}
	if progress <= 0 {
		return 0
	}
	ratio := float64(progress) / float64(duration)
	eased := 0.5 - 0.5*math.Cos(ratio*math.Pi)
	offset := int(math.Round(eased * float64(distance)))
	if offset < 0 {
		return 0
	}
	if offset > distance {
		return distance
	}
	return offset
}

func liveGamePanelCount() int {
	return 4
}

func drawLiveGameStaticTop(c PixelCanvas, game ScoreboardLiveGame) {
	awayColor := GetTeamColor(game.AwayTeam.Name)
	homeColor := GetTeamColor(game.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	awayAbbr := teamAbbrev(game.AwayTeam.ShortName, game.AwayTeam.Name)
	homeAbbr := teamAbbrev(game.HomeTeam.ShortName, game.HomeTeam.Name)
	drawText5x8CenteredSegmentsAtOffset(c, 0, 1, []text5x8Segment{
		{text: awayAbbr, col: awayColor},
		{text: " @ ", col: white},
		{text: homeAbbr, col: homeColor},
	})
	drawText5x8CenteredSegmentsInRangeAtOffset(c, -3, 48, 0, 10, []text5x8Segment{
		{text: fmt.Sprintf("%d", game.AwayTeam.Runs), col: awayColor},
		{text: " - ", col: white},
		{text: fmt.Sprintf("%d", game.HomeTeam.Runs), col: homeColor},
	})
	DrawBases(c, 50, 16, &game.Bases)
	drawLiveGameStatusRow(c, 19, game, grey, white)
}

func drawLiveGameBottomPanel(c PixelCanvas, game ScoreboardLiveGame, panelIndex int, xOffset int) {
	if panelIndex < 0 || panelIndex >= liveGamePanelCount() {
		return
	}

	awayColor := GetTeamColor(game.AwayTeam.Name)
	homeColor := GetTeamColor(game.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}

	switch panelIndex {
	case 0:
		drawLiveGameStatRow(c, xOffset, 29, "R", fmt.Sprintf("%d", game.AwayTeam.Runs), fmt.Sprintf("%d", game.HomeTeam.Runs), grey, awayColor, homeColor)
		drawLiveGameStatRow(c, xOffset, 38, "H", fmt.Sprintf("%d", game.AwayTeam.Hits), fmt.Sprintf("%d", game.HomeTeam.Hits), grey, awayColor, homeColor)
		drawLiveGameStatRow(c, xOffset, 47, "E", fmt.Sprintf("%d", game.AwayTeam.Errors), fmt.Sprintf("%d", game.HomeTeam.Errors), grey, awayColor, homeColor)
		drawLiveGameStatRow(c, xOffset, 56, "L", fmt.Sprintf("%d", game.AwayTeam.LOB), fmt.Sprintf("%d", game.HomeTeam.LOB), grey, awayColor, homeColor)
	case 1:
		drawText5x8CenteredAtOffset(c, xOffset, 29, liveGameBatterName(game), liveGameBatterColor(game))
		drawText5x8CenteredAtOffset(c, xOffset, 38, liveGameBatterPrimaryLine(game), grey)
		drawText5x8CenteredAtOffset(c, xOffset, 47, liveGameBatterSecondaryLine(game), grey)
		drawText5x8CenteredAtOffset(c, xOffset, 56, liveGameBatterSummaryLine(game), white)
	case 2:
		drawText5x8CenteredAtOffset(c, xOffset, 29, liveGamePitcherName(game), liveGamePitcherColor(game))
		drawText5x8CenteredAtOffset(c, xOffset, 38, liveGamePitcherPrimaryLine(game), grey)
		drawText5x8CenteredAtOffset(c, xOffset, 47, liveGamePitcherSecondaryLine(game), grey)
		drawText5x8CenteredAtOffset(c, xOffset, 56, liveGamePitcherTertiaryLine(game), white)
	case 3:
		drawLiveGameLastPlayPanel(c, game, xOffset, orange)
	}
}

func drawLiveGameLastPlayPanel(c PixelCanvas, game ScoreboardLiveGame, xOffset int, accentCol color.RGBA) {
	notation := lastPlayDisplayNotation(game)
	if notation == "" {
		lines := liveGameLastPlayLines(game)
		for i, line := range lines {
			drawText5x8CenteredAtOffset(c, xOffset, 31+i*9, line, accentCol)
		}
		return
	}

	panelH := 64
	startY := 29
	endY := panelH - 2
	diamondSize := endY - startY
	if diamondSize > panelH {
		diamondSize = panelH
	}
	if diamondSize <= 0 {
		return
	}

	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	drawScorebookDiamondCentered(c, xOffset, startY, diamondSize, accentCol)
	drawLastPlayAdvancementPath(c, xOffset, startY, diamondSize, notation, white)
	drawLastPlayRBIDots(c, xOffset, game.LastPlayRBIs, accentCol)

	textY := startY + diamondSize/2 - 4
	drawLastPlayNotationCenteredAtOffset(c, xOffset, textY, game, accentCol)
}

func drawLiveGameLastPlayPanelBuffered(c PixelCanvas, game ScoreboardLiveGame, xOffset int) {
	panel := NewMockCanvas(64, 64)
	background := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}

	for x := 0; x < panel.w; x++ {
		for y := 0; y < panel.h; y++ {
			panel.Set(x, y, background)
		}
	}

	drawLiveGameLastPlayPanel(panel, game, 0, orange)

	bounds := c.Bounds()
	xStart := 0
	if xOffset < bounds.Min.X {
		xStart = bounds.Min.X - xOffset
	}
	xEnd := panel.w
	if xOffset+xEnd > bounds.Max.X {
		xEnd = bounds.Max.X - xOffset
	}
	if xStart >= xEnd {
		return
	}

	yStart := 29
	if yStart < bounds.Min.Y {
		yStart = bounds.Min.Y
	}
	yEnd := panel.h
	if yEnd > bounds.Max.Y {
		yEnd = bounds.Max.Y
	}
	if yStart >= yEnd {
		return
	}

	for y := yStart; y < yEnd; y++ {
		rowOffset := y * panel.w
		for x := xStart; x < xEnd; x++ {
			c.Set(xOffset+x, y, panel.pix[rowOffset+x])
		}
	}
}

func drawLastPlayNotationCenteredAtOffset(c PixelCanvas, xOffset, y int, game ScoreboardLiveGame, col color.RGBA) {
	notation := lastPlayDisplayNotation(game)
	if isLookingStrikeoutNotation(notation, game.LastPlay) {
		drawMirroredK5x8CenteredAtOffset(c, xOffset, y, col)
		return
	}
	drawText5x8CenteredAtOffset(c, xOffset, y, notation, col)
}

func lastPlayDisplayNotation(game ScoreboardLiveGame) string {
	notation := strings.TrimSpace(game.LastPlayNotation)
	if notation == "" {
		return ""
	}
	if game.LastPlayRBIs <= 0 {
		return notation
	}
	suffix := fmt.Sprintf(", %d RBI", game.LastPlayRBIs)
	return strings.TrimSuffix(notation, suffix)
}

func isLookingStrikeoutNotation(notation, lastPlay string) bool {
	return strings.EqualFold(strings.TrimSpace(notation), "K") &&
		strings.Contains(strings.ToLower(lastPlay), "looking")
}

func drawMirroredK5x8CenteredAtOffset(c PixelCanvas, xOffset, y int, col color.RGBA) {
	glyph, ok := font5x8['K']
	if !ok {
		drawText5x8CenteredAtOffset(c, xOffset, y, "K", col)
		return
	}

	panelWidth := c.Bounds().Max.X
	x := xOffset + (panelWidth-fontW5x8)/2
	if x < xOffset {
		x = xOffset
	}

	for row := 0; row < fontH5x8; row++ {
		mirrored := mirror5BitRow(glyph[row])
		for colIdx := 0; colIdx < fontW5x8; colIdx++ {
			if mirrored&(1<<uint(fontW5x8-1-colIdx)) != 0 {
				c.Set(x+colIdx, y+row, col)
			}
		}
	}
}

func mirror5BitRow(row byte) byte {
	var mirrored byte
	for i := 0; i < fontW5x8; i++ {
		if row&(1<<uint(i)) != 0 {
			mirrored |= 1 << uint(fontW5x8-1-i)
		}
	}
	return mirrored
}

func lastPlayAchievedBases(notation string) (first, second, third, home bool) {
	switch strings.ToUpper(strings.TrimSpace(notation)) {
	case "1B", "BB", "HBP":
		return true, false, false, false
	case "2B":
		return true, true, false, false
	case "3B":
		return true, true, true, false
	case "HR":
		return true, true, true, true
	default:
		return false, false, false, false
	}
}

func drawLastPlayAdvancementPath(c PixelCanvas, xOffset, topY, size int, notation string, col color.RGBA) {
	first, second, third, home := lastPlayAchievedBases(notation)
	if !first && !second && !third && !home {
		return
	}

	panelW := 64
	x := xOffset + (panelW-size)/2
	if x < xOffset {
		x = xOffset
	}
	cx, cy := x+size/2, topY+size/2
	r := size / 2

	if first {
		drawOffsetDiamondEdge(c, cx, cy, r, "home-first", col)
	}
	if second {
		drawOffsetDiamondEdge(c, cx, cy, r, "first-second", col)
	}
	if third {
		drawOffsetDiamondEdge(c, cx, cy, r, "second-third", col)
	}
	if home {
		drawOffsetDiamondEdge(c, cx, cy, r, "third-home", col)
	}
}

func drawOffsetDiamondEdge(c PixelCanvas, cx, cy, r int, edge string, col color.RGBA) {
	for i := 0; i <= r; i++ {
		x := cx
		y := cy
		switch edge {
		case "home-first":
			x = cx + i + 1
			y = cy + (r - i) + 1
		case "first-second":
			x = cx + i + 1
			y = cy - (r - i) - 1
		case "second-third":
			x = cx - i - 1
			y = cy - (r - i) - 1
		case "third-home":
			x = cx - i - 1
			y = cy + (r - i) + 1
		default:
			return
		}
		c.Set(x, y, col)
	}
}

func drawSmallDiamondOutline(c PixelCanvas, cx, cy, r int, col color.RGBA) {
	for dy := -r; dy <= r; dy++ {
		span := r - absInt(dy)
		if span == 0 {
			c.Set(cx, cy+dy, col)
			continue
		}
		c.Set(cx-span, cy+dy, col)
		c.Set(cx+span, cy+dy, col)
	}
}

func drawLastPlayRBIDots(c PixelCanvas, xOffset, rbis int, col color.RGBA) {
	if rbis <= 0 {
		return
	}
	if rbis > 4 {
		rbis = 4
	}

	x := xOffset + 56
	startY := 34
	gap := 7
	for i := 0; i < rbis; i++ {
		drawSmallDiamondOutline(c, x, startY+i*gap, 1, col)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func drawLiveGameStatRow(c PixelCanvas, xOffset, y int, label, awayVal, homeVal string, labelCol, awayCol, homeCol color.RGBA) {
	drawText5x8CenteredInRangeAtOffset(c, 0, 16, xOffset, y, label, labelCol)
	drawText5x8CenteredInRangeAtOffset(c, 16, 20, xOffset, y, awayVal, awayCol)
	drawText5x8CenteredInRangeAtOffset(c, 44, 20, xOffset, y, homeVal, homeCol)
}

func liveGameStatusHalf(game ScoreboardLiveGame) string {
	half := strings.ToUpper(strings.TrimSpace(game.HalfInning))
	switch half {
	case "TOP":
		return "T"
	case "BOTTOM":
		return "B"
	case "MID":
		return "M"
	case "END":
		return "E"
	default:
		return trimToChars(half, 1)
	}
}

func liveGameStatusInningText(game ScoreboardLiveGame) string {
	return fmt.Sprintf("%s%d", liveGameStatusHalf(game), game.Inning)
}

func measureLiveGameStatusTextWidth(game ScoreboardLiveGame) int {
	inningText := liveGameStatusInningText(game)
	ballsText := fmt.Sprintf("%d", game.Balls)
	strikesText := fmt.Sprintf("%d", game.Strikes)
	inningGap := 3
	countDashGap := 0
	dashW := 3
	return measureText5x8Width(inningText) + inningGap + measureText5x8Width(ballsText) + countDashGap + dashW + countDashGap + measureText5x8Width(strikesText)
}

func drawLiveGameStatusText(c PixelCanvas, x, y int, game ScoreboardLiveGame, textCol color.RGBA) int {
	inningText := liveGameStatusInningText(game)
	ballsText := fmt.Sprintf("%d", game.Balls)
	strikesText := fmt.Sprintf("%d", game.Strikes)
	inningGap := 3
	countDashGap := 0

	DrawText5x8(c, x, y, inningText, textCol)
	x += measureText5x8Width(inningText) + inningGap
	DrawText5x8(c, x, y, ballsText, textCol)
	x += measureText5x8Width(ballsText) + countDashGap
	drawCompactStatusDash(c, x, y, textCol)
	x += 3 + countDashGap
	DrawText5x8(c, x, y, strikesText, textCol)
	return x + measureText5x8Width(strikesText)
}

func drawCompactStatusDash(c PixelCanvas, x, y int, col color.RGBA) {
	for dx := 0; dx < 3; dx++ {
		c.Set(x+dx, y+3, col)
	}
}

func drawLiveGameStatusRow(c PixelCanvas, y int, game ScoreboardLiveGame, textCol, outFillCol color.RGBA) {
	textW := measureLiveGameStatusTextWidth(game)
	circleW := 7
	circleGap := 2
	textGap := 3
	totalW := textW + textGap + circleW*3 + circleGap*2
	x := (c.Bounds().Max.X - totalW) / 2
	if x < 0 {
		x = 0
	}
	circleX := drawLiveGameStatusText(c, x, y, game, textCol) + textGap
	for i := 0; i < 3; i++ {
		drawOutCircle(c, circleX+i*(circleW+circleGap), y+1, i < game.Outs, textCol, outFillCol)
	}
}

func liveGameBatterName(game ScoreboardLiveGame) string {
	return trimText5x8ToWidth(strings.ToUpper(compactLivePlayerName(game.CurrentBatter.FullName, game.CurrentBatter.LastName, 12)), 62)
}

func liveGameBatterColor(game ScoreboardLiveGame) color.RGBA {
	if strings.EqualFold(strings.TrimSpace(game.HalfInning), "bottom") {
		return GetTeamColor(game.HomeTeam.Name)
	}
	return GetTeamColor(game.AwayTeam.Name)
}

func liveGameBatterPrimaryLine(game ScoreboardLiveGame) string {
	pos := strings.ToUpper(strings.TrimSpace(game.CurrentBatter.CurrentPosition))
	avg := strings.TrimSpace(game.CurrentBatter.SeasonBattingAverage)
	if pos == "" && avg == "" {
		return "AT BAT"
	}
	line := strings.TrimSpace(strings.Join([]string{pos, "AVG", avg}, " "))
	return trimText5x8ToWidth(strings.ToUpper(line), 62)
}

func liveGameBatterSecondaryLine(game ScoreboardLiveGame) string {
	ops := strings.TrimSpace(game.CurrentBatter.SeasonOPS)
	if ops == "" {
		return "OPS TBD"
	}
	return trimText5x8ToWidth(strings.ToUpper("OPS "+ops), 62)
}

func liveGameBatterSummaryLine(game ScoreboardLiveGame) string {
	summary := strings.TrimSpace(game.CurrentBatter.Summary)
	if summary == "" && game.CurrentBatter.GameAtBats > 0 {
		summary = fmt.Sprintf("%d-%d", game.CurrentBatter.GameHits, game.CurrentBatter.GameAtBats)
	}
	if summary == "" {
		summary = "TBD"
	}
	return trimText5x8ToWidth(strings.ToUpper(summary), 62)
}

func liveGamePitcherName(game ScoreboardLiveGame) string {
	return trimText5x8ToWidth(strings.ToUpper(compactLivePlayerName(game.CurrentPitcher.FullName, game.CurrentPitcher.LastName, 12)), 62)
}

func liveGamePitcherColor(game ScoreboardLiveGame) color.RGBA {
	if strings.EqualFold(strings.TrimSpace(game.HalfInning), "bottom") {
		return GetTeamColor(game.AwayTeam.Name)
	}
	return GetTeamColor(game.HomeTeam.Name)
}

func liveGamePitcherPrimaryLine(game ScoreboardLiveGame) string {
	hand := pitcherHandLabel(game.CurrentPitcher)
	if hand == "" {
		return "PITCHER"
	}
	return trimText5x8ToWidth(strings.ToUpper(hand), 62)
}

func liveGamePitcherSecondaryLine(game ScoreboardLiveGame) string {
	era := strings.TrimSpace(game.CurrentPitcher.ERA)
	if era == "" {
		return "ERA TBD"
	}
	return trimText5x8ToWidth(strings.ToUpper("ERA "+era), 62)
}

func liveGamePitcherTertiaryLine(game ScoreboardLiveGame) string {
	return trimText5x8ToWidth(strings.ToUpper(fmt.Sprintf("W-L %d-%d", game.CurrentPitcher.Wins, game.CurrentPitcher.Losses)), 62)
}

func liveGameLastPlayLines(game ScoreboardLiveGame) []string {
	text := strings.TrimSpace(game.LastPlayNotation)
	if text == "" {
		text = strings.TrimSpace(game.LastPlay)
	}
	if text == "" {
		text = "NO LAST PLAY"
	}
	lines := wrapTextLines(strings.ToUpper(text), 10, 3)
	for i := range lines {
		lines[i] = trimText5x8ToWidth(lines[i], 62)
	}
	for len(lines) < 3 {
		lines = append(lines, "")
	}
	return lines[:3]
}

// drawScorebookDiamondCentered draws a simple diamond (base) graphic centered
// in the bottom panel area. xOffset is the panel's horizontal offset; topY is the
// starting Y coordinate to draw the diamond pattern.
// drawScorebookDiamondCentered draws a scalable, outline-only diamond (square) centered in the panel.
func drawScorebookDiamondCentered(c PixelCanvas, xOffset, topY, size int, col color.RGBA) {
	panelW := 64 // Always center diamond in a single panel, not the whole canvas
	x := xOffset + (panelW-size)/2
	if x < xOffset {
		x = xOffset
	}
	// Draw outline diamond using Bresenham's line algorithm for each edge
	cx, cy := x+size/2, topY+size/2
	r := size / 2
	for i := 0; i <= r; i++ {
		// Four symmetric points per row
		c.Set(cx-i, cy-(r-i), col) // upper left
		c.Set(cx+i, cy-(r-i), col) // upper right
		c.Set(cx-i, cy+(r-i), col) // lower left
		c.Set(cx+i, cy+(r-i), col) // lower right
	}
}

// DrawLargeScore draws a two-digit score in a larger format
func DrawLargeScore(c PixelCanvas, x, y int, score int) {
	tens := score / 10
	ones := score % 10

	// Draw tens digit
	DrawBigDigit(c, x, y, tens)
	// Draw ones digit
	DrawBigDigit(c, x+12, y, ones)
}

// DrawBigDigit draws a single digit in larger format (roughly 10x10)
func DrawBigDigit(c PixelCanvas, x, y int, digit int) {
	col := color.RGBA{R: 100, G: 255, B: 100, A: 255}

	// Simple 7-segment style digit patterns
	switch digit {
	case 0:
		// Draw box outline
		for i := 0; i < 8; i++ {
			c.Set(x+i, y, col)
			c.Set(x+i, y+8, col)
			c.Set(x, y+i, col)
			c.Set(x+7, y+i, col)
		}
	case 1:
		// Two vertical lines on right
		for i := 0; i < 8; i++ {
			c.Set(x+6, y+i, col)
			c.Set(x+7, y+i, col)
		}
	case 2:
		// Top line
		for i := 0; i < 8; i++ {
			c.Set(x+i, y, col)
		}
		// Top-right vertical
		for i := 0; i < 4; i++ {
			c.Set(x+6, y+i, col)
		}
		// Middle line
		for i := 0; i < 8; i++ {
			c.Set(x+i, y+4, col)
		}
		// Bottom-left vertical
		for i := 4; i < 8; i++ {
			c.Set(x, y+i, col)
		}
		// Bottom line
		for i := 0; i < 8; i++ {
			c.Set(x+i, y+8, col)
		}
	default:
		// Draw a simple line for unknown digits
		for i := 0; i < 8; i++ {
			c.Set(x+i, y+4, col)
		}
	}
}

// DrawBases draws three slightly larger diamonds for first/top/third bases.
func DrawBases(c PixelCanvas, cx, cy int, bases *ScoreboardLiveGameBases) {
	emptyCol := color.RGBA{R: 50, G: 50, B: 50, A: 255}
	filledCol := color.RGBA{R: 200, G: 100, B: 0, A: 255}
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	r := 3

	drawDiamond := func(x, y int, filled bool) {
		col := emptyCol
		if filled {
			col = filledCol
		}
		for dy := -r; dy <= r; dy++ {
			span := r - abs(dy)
			if filled {
				for dx := -span; dx <= span; dx++ {
					c.Set(x+dx, y+dy, col)
				}
				continue
			}
			if span == 0 {
				c.Set(x, y+dy, col)
				continue
			}
			c.Set(x-span, y+dy, col)
			c.Set(x+span, y+dy, col)
		}
	}

	firstFilled, secondFilled, thirdFilled := false, false, false
	if bases != nil {
		firstFilled = bases.First
		secondFilled = bases.Second
		thirdFilled = bases.Third
	}

	// First base is on the right, second is on top, third is on the left.
	drawDiamond(cx+7, cy-2, firstFilled)
	drawDiamond(cx, cy-5, secondFilled)
	drawDiamond(cx-7, cy-2, thirdFilled)
}

// DrawText draws text using the compact 3x5 bitmap font at position (x, y).
// Each character is fontW px wide, fontH px tall, advancing fontAdv px per char.
func DrawText(c PixelCanvas, x, y int, text string, col color.RGBA) {
	bounds := c.Bounds()
	cx := x
	for _, ch := range text {
		if cx >= bounds.Max.X {
			break
		}

		glyph, ok := font3x5[ch]
		if !ok {
			glyph = font3x5[' ']
		}

		for row := 0; row < fontH; row++ {
			for col_idx := 0; col_idx < fontW; col_idx++ {
				if glyph[row]&(1<<uint(fontW-1-col_idx)) != 0 {
					px := cx + col_idx
					py := y + row
					if px >= 0 && px < bounds.Max.X && py >= 0 && py < bounds.Max.Y {
						c.Set(px, py, col)
					}
				}
			}
		}
		cx += fontAdv
	}
}

// DrawTextSmall is an alias for DrawText - same font, used throughout
func DrawTextSmall(c PixelCanvas, x, y int, text string, col color.RGBA) {
	DrawText(c, x, y, text, col)
}

// DrawText3x3 draws text using the 3x3 bitmap font at position (x, y).
// Compact height (3px), normal width (3px), same advance as 3x5.
func DrawText3x3(c PixelCanvas, x, y int, text string, col color.RGBA) {
	bounds := c.Bounds()
	cx := x
	for _, ch := range text {
		if cx >= bounds.Max.X {
			break
		}

		glyph, ok := font3x3[ch]
		if !ok {
			glyph = font3x3[' ']
		}

		for row := 0; row < fontH3x3; row++ {
			for col_idx := 0; col_idx < fontW3x3; col_idx++ {
				if glyph[row]&(1<<uint(fontW3x3-1-col_idx)) != 0 {
					px := cx + col_idx
					py := y + row
					if px >= 0 && px < bounds.Max.X && py >= 0 && py < bounds.Max.Y {
						c.Set(px, py, col)
					}
				}
			}
		}
		cx += fontAdv3x3
	}
}

// DrawText3x3Centered renders 3x3 text centered horizontally on the canvas.
func DrawText3x3Centered(c PixelCanvas, y int, text string, col color.RGBA) {
	w := c.Bounds().Max.X
	tw := len(text) * fontAdv3x3
	x := (w - tw) / 2
	if x < 0 {
		x = 0
	}
	DrawText3x3(c, x, y, text, col)
}

// DrawText3x4 draws text using the 3x4 bitmap font at position (x, y).
// Balanced height (4px), normal width (3px), same advance as 3x5.
func DrawText3x4(c PixelCanvas, x, y int, text string, col color.RGBA) {
	bounds := c.Bounds()
	cx := x
	for _, ch := range text {
		if cx >= bounds.Max.X {
			break
		}

		glyph, ok := font3x4[ch]
		if !ok {
			glyph = font3x4[' ']
		}

		for row := 0; row < fontH3x4; row++ {
			for col_idx := 0; col_idx < fontW3x4; col_idx++ {
				if glyph[row]&(1<<uint(fontW3x4-1-col_idx)) != 0 {
					px := cx + col_idx
					py := y + row
					if px >= 0 && px < bounds.Max.X && py >= 0 && py < bounds.Max.Y {
						c.Set(px, py, col)
					}
				}
			}
		}
		cx += fontAdv3x4
	}
}

// DrawText3x4Centered renders 3x4 text centered horizontally on the canvas.
func DrawText3x4Centered(c PixelCanvas, y int, text string, col color.RGBA) {
	w := c.Bounds().Max.X
	tw := len(text) * fontAdv3x4
	x := (w - tw) / 2
	if x < 0 {
		x = 0
	}
	DrawText3x4(c, x, y, text, col)
}

// DrawText3x5 draws text using the 3x5 bitmap font at position (x, y).
func DrawText3x5(c PixelCanvas, x, y int, text string, col color.RGBA) {
	bounds := c.Bounds()
	cx := x
	for _, ch := range text {
		if cx >= bounds.Max.X {
			break
		}

		glyph, ok := font3x5[ch]
		if !ok {
			glyph = font3x5[' ']
		}

		for row := 0; row < fontH; row++ {
			for colIdx := 0; colIdx < fontW; colIdx++ {
				if glyph[row]&(1<<uint(fontW-1-colIdx)) != 0 {
					px := cx + colIdx
					py := y + row
					if px >= 0 && px < bounds.Max.X && py >= 0 && py < bounds.Max.Y {
						c.Set(px, py, col)
					}
				}
			}
		}
		cx += fontAdv
	}
}

// DrawText3x5Centered renders 3x5 text centered horizontally on the canvas.
func DrawText3x5Centered(c PixelCanvas, y int, text string, col color.RGBA) {
	w := c.Bounds().Max.X
	tw := len(text) * fontAdv
	x := (w - tw) / 2
	if x < 0 {
		x = 0
	}
	DrawText3x5(c, x, y, text, col)
}

func DrawText5x8(c PixelCanvas, x, y int, text string, col color.RGBA) {
	bounds := c.Bounds()
	cx := x
	runes := []rune(text)
	for i, ch := range runes {
		if cx >= bounds.Max.X {
			break
		}

		glyph, ok := font5x8[ch]
		if !ok {
			glyph = font5x8[' ']
		}

		for row := 0; row < fontH5x8; row++ {
			for colIdx := 0; colIdx < fontW5x8; colIdx++ {
				if glyph[row]&(1<<uint(fontW5x8-1-colIdx)) != 0 {
					px := cx + colIdx
					py := y + row
					if px >= 0 && px < bounds.Max.X && py >= 0 && py < bounds.Max.Y {
						c.Set(px, py, col)
					}
				}
			}
		}
		cx += advance5x8(ch, i == len(runes)-1)
	}
}

func DrawText5x8Centered(c PixelCanvas, y int, text string, col color.RGBA) {
	w := c.Bounds().Max.X
	tw := measureText5x8Width(text)
	x := (w - tw) / 2
	if x < 0 {
		x = 0
	}
	DrawText5x8(c, x, y, text, col)
}

func measureText5x8Width(text string) int {
	width := 0
	runes := []rune(text)
	for i, ch := range runes {
		width += advance5x8(ch, i == len(runes)-1)
	}
	return width
}

func trimText5x8ToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(text)
	if measureText5x8Width(text) <= maxWidth {
		return text
	}
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		trimmed := string(runes)
		if measureText5x8Width(trimmed) <= maxWidth {
			return trimmed
		}
	}
	return ""
}

func advance5x8(ch rune, isLast bool) int {
	if ch == ' ' {
		if isLast {
			return 0
		}
		return 4
	}
	if isLast {
		return fontW5x8
	}
	return fontAdv5x8
}

func drawText3x5CenteredInRange(c PixelCanvas, xStart, regionW, y int, text string, col color.RGBA) {
	tw := len(text) * fontAdv
	x := xStart + (regionW-tw)/2
	if x < xStart {
		x = xStart
	}
	DrawText3x5(c, x, y, text, col)
}

func compactLivePlayerName(fullName, lastName string, maxChars int) string {
	fullName = strings.TrimSpace(fullName)
	lastName = strings.TrimSpace(lastName)
	if lastName != "" && lastName != fullName {
		return trimToChars(lastName, maxChars)
	}
	if fullName != "" {
		return trimToChars(fullName, maxChars)
	}
	return "TBD"
}

// GetTeamColor returns a team's brand color
func GetTeamColor(teamName string) color.RGBA {
	// Simplified team colors
	switch teamName {
	case "Washington Nationals":
		return color.RGBA{R: 171, G: 0, B: 40, A: 255} // Red
	case "New York Mets":
		return color.RGBA{R: 0, G: 33, B: 71, A: 255} // Blue
	case "Philadelphia Phillies":
		return color.RGBA{R: 155, G: 25, B: 25, A: 255} // Maroon
	case "Atlanta Braves":
		return color.RGBA{R: 206, G: 17, B: 38, A: 255} // Red
	case "Miami Marlins":
		return color.RGBA{R: 0, G: 41, B: 82, A: 255} // Dark Blue
	case "New York Yankees":
		return color.RGBA{R: 12, G: 35, B: 64, A: 255} // Navy
	case "Boston Red Sox":
		return color.RGBA{R: 189, G: 16, B: 32, A: 255} // Red
	case "Tampa Bay Rays":
		return color.RGBA{R: 0, G: 48, B: 135, A: 255} // Blue
	case "Toronto Blue Jays":
		return color.RGBA{R: 0, G: 48, B: 133, A: 255} // Blue
	case "Chicago White Sox":
		return color.RGBA{R: 39, G: 34, B: 44, A: 255} // Black
	case "Cleveland Guardians":
		return color.RGBA{R: 0, G: 43, B: 57, A: 255} // Navy
	case "Detroit Tigers":
		return color.RGBA{R: 12, G: 35, B: 64, A: 255} // Navy
	case "Kansas City Royals":
		return color.RGBA{R: 16, G: 38, B: 103, A: 255} // Blue
	case "Minnesota Twins":
		return color.RGBA{R: 2, G: 33, B: 47, A: 255} // Navy
	case "Houston Astros":
		return color.RGBA{R: 235, G: 108, B: 35, A: 255} // Orange
	case "Los Angeles Angels":
		return color.RGBA{R: 186, G: 0, B: 33, A: 255} // Red
	case "Oakland Athletics":
		return color.RGBA{R: 3, G: 46, B: 66, A: 255} // Dark Green
	case "Seattle Mariners":
		return color.RGBA{R: 12, G: 60, B: 96, A: 255} // Navy
	case "Texas Rangers":
		return color.RGBA{R: 0, G: 34, B: 85, A: 255} // Blue
	case "Arizona Diamondbacks":
		return color.RGBA{R: 167, G: 25, B: 48, A: 255} // Red
	case "Colorado Rockies":
		return color.RGBA{R: 51, G: 38, B: 102, A: 255} // Purple
	case "Los Angeles Dodgers":
		return color.RGBA{R: 0, G: 43, B: 94, A: 255} // Blue
	case "San Diego Padres":
		return color.RGBA{R: 44, G: 31, B: 71, A: 255} // Brown
	case "San Francisco Giants":
		return color.RGBA{R: 253, G: 103, B: 8, A: 255} // Orange
	case "Chicago Cubs":
		return color.RGBA{R: 14, G: 43, B: 74, A: 255} // Blue
	case "Cincinnati Reds":
		return color.RGBA{R: 198, G: 12, B: 12, A: 255} // Red
	case "Milwaukee Brewers":
		return color.RGBA{R: 19, G: 51, B: 96, A: 255} // Navy
	case "Pittsburgh Pirates":
		return color.RGBA{R: 39, G: 33, B: 39, A: 255} // Black
	case "St. Louis Cardinals":
		return color.RGBA{R: 198, G: 12, B: 12, A: 255} // Red
	default:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255} // Gray
	}
}

// formatInt converts an int to a padded string
func formatInt(val, digits int) string {
	str := ""
	for i := 0; i < digits-1; i++ {
		if val < pow(10, digits-i-1) {
			str += "0"
		}
	}
	// Convert number to string manually (no strconv)
	numStr := ""
	if val == 0 {
		numStr = "0"
	} else {
		temp := val
		for temp > 0 {
			numStr = string(rune('0'+(temp%10))) + numStr
			temp /= 10
		}
	}
	return str + numStr
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// wrapTextLines wraps text into up to maxLines lines of maxChars each.
// It prefers word boundaries and falls back to hard truncation for long words.
func wrapTextLines(text string, maxChars, maxLines int) []string {
	if maxChars <= 0 || maxLines <= 0 {
		return []string{}
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	words := strings.Fields(text)
	lines := make([]string, 0, maxLines)
	curr := ""

	flush := func() {
		if curr != "" && len(lines) < maxLines {
			lines = append(lines, curr)
			curr = ""
		}
	}

	for _, w := range words {
		if len(w) > maxChars {
			flush()
			for len(w) > 0 && len(lines) < maxLines {
				chunk := w
				if len(chunk) > maxChars {
					chunk = w[:maxChars]
				}
				lines = append(lines, chunk)
				w = w[len(chunk):]
			}
			if len(lines) >= maxLines {
				break
			}
			continue
		}

		if curr == "" {
			curr = w
			continue
		}

		if len(curr)+1+len(w) <= maxChars {
			curr += " " + w
		} else {
			flush()
			curr = w
		}

		if len(lines) >= maxLines {
			break
		}
	}
	flush()

	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	if len(lines) == maxLines {
		last := lines[maxLines-1]
		if len(last) > maxChars {
			lines[maxLines-1] = last[:maxChars]
		}
	}
	return lines
}

// formatVenueLines wraps venue text for small LED panels.
func formatVenueLines(venue string, maxChars, maxLines int) []string {
	venue = strings.Join(strings.Fields(strings.TrimSpace(venue)), " ")
	if venue == "" {
		return []string{}
	}

	lines := wrapTextLines(venue, maxChars, maxLines)
	if len(lines) == 0 {
		return []string{}
	}

	full := wrapTextLines(venue, maxChars, 99)
	if len(full) > maxLines {
		lines[len(lines)-1] = ellipsizeASCII(lines[len(lines)-1], maxChars)
	}

	return lines
}

func ellipsizeASCII(s string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	if len(s) <= maxChars {
		if len(s) >= 3 {
			return s[:len(s)-3] + "..."
		}
		return trimToChars(s, maxChars)
	}
	if maxChars <= 3 {
		return trimToChars("...", maxChars)
	}
	return s[:maxChars-3] + "..."
}

// Removed RenderDivisionStandings/RenderLiveGameScore wrappers.
// Rendering is initiated directly by display functions via DrawDivisionStandings/DrawLiveGameScore.

// drawTextCenteredInRange centers text in a horizontal sub-region of the canvas.
func drawTextCenteredInRange(c PixelCanvas, xStart, regionW, y int, text string, col color.RGBA) {
	tw := len(text) * fontAdv
	x := xStart + (regionW-tw)/2
	if x < xStart {
		x = xStart
	}
	DrawText(c, x, y, text, col)
}
