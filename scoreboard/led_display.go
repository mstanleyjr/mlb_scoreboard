package scoreboard

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"
)

const (
	divisionStandingsTopPadding    = 2
	divisionStandingsTitleH        = 8
	divisionStandingsTitleGap      = 5
	divisionStandingsTitleRuleGap  = 3
	divisionStandingsRowBlockH     = 19
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

				lineTop := nameY
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
		if sectionIndex < len(displayInfo.Sections)-1 {
			height += divisionStandingsSectionGap
		}
	}
	return height
}

func DrawNextMatchup(c PixelCanvas, m ScoreboardNextMatchup) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	green := color.RGBA{R: 100, G: 200, B: 100, A: 255}
	awayCol := GetTeamColor(m.AwayTeam.Name)
	homeCol := GetTeamColor(m.HomeTeam.Name)

	const rowH = fontRowH + 2

	header := "NEXT"
	if m.GameType != "" && m.GameType != "Regular Season" {
		header = trimToChars("NEXT "+m.GameType, 16)
	}
	drawTextCentered(c, 1, header, yellow)
	drawTextCentered(c, 1+rowH, formatNextMatchupDateTime(m.DateTime), green)

	colW := width / 2
	awayX := 0
	homeX := colW

	awayAbbr := teamAbbrev(m.AwayTeam.ShortName, m.AwayTeam.Name)
	homeAbbr := teamAbbrev(m.HomeTeam.ShortName, m.HomeTeam.Name)
	matchup := trimToChars(awayAbbr+" @ "+homeAbbr, 16)
	drawTextCentered(c, 1+rowH*2, matchup, color.RGBA{R: 220, G: 220, B: 220, A: 255})

	awayRecord := trimToChars(fmt.Sprintf("%d-%d", m.AwayTeam.Record.Wins, m.AwayTeam.Record.Losses), 8)
	homeRecord := trimToChars(fmt.Sprintf("%d-%d", m.HomeTeam.Record.Wins, m.HomeTeam.Record.Losses), 8)
	drawTextCenteredInRange(c, awayX, colW, 1+rowH*3, awayRecord, awayCol)
	drawTextCenteredInRange(c, homeX, colW, 1+rowH*3, homeRecord, homeCol)

	awayPitch := trimToChars(pitcherNameOnly(m.AwayTeam.ProbablePitcher), 8)
	homePitch := trimToChars(pitcherNameOnly(m.HomeTeam.ProbablePitcher), 8)
	drawTextCenteredInRange(c, awayX, colW, 1+rowH*4, awayPitch, awayCol)
	drawTextCenteredInRange(c, homeX, colW, 1+rowH*4, homePitch, homeCol)

	awayHand := pitcherHandLabel(m.AwayTeam.ProbablePitcher)
	homeHand := pitcherHandLabel(m.HomeTeam.ProbablePitcher)
	if awayHand != "" {
		drawTextCenteredInRange(c, awayX, colW, 1+rowH*5, awayHand, grey)
	}
	if homeHand != "" {
		drawTextCenteredInRange(c, homeX, colW, 1+rowH*5, homeHand, grey)
	}

	awayERA := m.AwayTeam.ProbablePitcher.ERA
	homeERA := m.HomeTeam.ProbablePitcher.ERA
	if awayERA != "" {
		drawTextCenteredInRange(c, awayX, colW, 1+rowH*5, awayERA, grey)
	}
	if homeERA != "" {
		drawTextCenteredInRange(c, homeX, colW, 1+rowH*5, homeERA, grey)
	}

	venueLines := formatVenueLines(m.Venue, 16, 2)
	if len(venueLines) == 1 {
		drawTextCentered(c, 1+rowH*7, venueLines[0], grey)
	} else if len(venueLines) >= 2 {
		drawTextCentered(c, 1+rowH*6, venueLines[0], grey)
		drawTextCentered(c, 1+rowH*7, venueLines[1], grey)
	}

	_ = height
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

// DrawLastMatchup renders the most recently completed game result.
//
// Layout (64x64, 8px rows):
//
//	y= 1  "FINAL" or "FINAL/10" for extra innings
//	y= 9  column headers: "R H E L"
//	y=17  away row: team + R/H/E/LOB
//	y=25  home row: team + R/H/E/LOB
//	y=33  W: winning pitcher last name (green)
//	y=41  L: losing pitcher last name (red)
//	y=49  S: save pitcher last name OR game time
func DrawLastMatchup(c PixelCanvas, m ScoreboardLastMatchup) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	red := color.RGBA{R: 220, G: 60, B: 60, A: 255}
	green := color.RGBA{R: 100, G: 200, B: 100, A: 255}

	const rowH = fontRowH + 2

	header := "FINAL"
	if m.FinalInning > 9 {
		header = fmt.Sprintf("FINAL/%d", m.FinalInning)
	}
	drawTextCentered(c, 1, trimToChars(header, 16), yellow)

	awayAbbr := teamAbbrev(m.AwayTeam.Team.ShortName, m.AwayTeam.Team.Name)
	homeAbbr := teamAbbrev(m.HomeTeam.Team.ShortName, m.HomeTeam.Team.Name)
	gridX := 1
	gridY := 1 + rowH
	gridW := width - 2
	gridH := 17
	winBg := color.RGBA{R: 0, G: 90, B: 0, A: 255}
	if gridW >= 8 && height >= gridY+gridH {
		colWidths := []int{14, 12, 12, 12, 12}
		for dx := 0; dx < gridW; dx++ {
			c.Set(gridX+dx, gridY, grey)
			c.Set(gridX+dx, gridY+gridH-1, grey)
		}
		for dy := 0; dy < gridH; dy++ {
			c.Set(gridX, gridY+dy, grey)
			c.Set(gridX+gridW-1, gridY+dy, grey)
		}
		colX := gridX + 1
		for i := 0; i < len(colWidths)-1; i++ {
			colX += colWidths[i]
			for dy := 1; dy < gridH-1; dy++ {
				c.Set(colX, gridY+dy, grey)
			}
		}
		h1 := gridY + 5
		h2 := gridY + 10
		for dx := 1; dx < gridW-1; dx++ {
			c.Set(gridX+dx, h1, grey)
			c.Set(gridX+dx, h2, grey)
		}

		// Header row: centered R/H/E/L labels.
		statLabels := []string{"", "R", "H", "E", "L"}
		cellX := gridX + 1
		for i, label := range statLabels {
			drawText3x4CenteredInRange(c, cellX, colWidths[i], gridY+1, label, grey)
			cellX += colWidths[i]
		}

		awayVals := []string{awayAbbr, fmt.Sprintf("%d", m.AwayTeam.Team.Runs), fmt.Sprintf("%d", m.AwayTeam.Team.Hits), fmt.Sprintf("%d", m.AwayTeam.Team.Errors), fmt.Sprintf("%d", m.AwayTeam.Team.LOB)}
		homeVals := []string{homeAbbr, fmt.Sprintf("%d", m.HomeTeam.Team.Runs), fmt.Sprintf("%d", m.HomeTeam.Team.Hits), fmt.Sprintf("%d", m.HomeTeam.Team.Errors), fmt.Sprintf("%d", m.HomeTeam.Team.LOB)}

		awayCol := GetTeamColor(m.AwayTeam.Team.Name)
		homeCol := GetTeamColor(m.HomeTeam.Team.Name)
		if m.AwayTeam.Winner {
			awayCol = white
		}
		if m.HomeTeam.Winner {
			homeCol = white
		}

		fillRow := func(rowTop int, rowBottom int, teamCol color.RGBA) {
			for x := gridX + 1; x < gridX+gridW-1; x++ {
				for y := rowTop; y <= rowBottom; y++ {
					c.Set(x, y, teamCol)
				}
			}
		}
		if m.AwayTeam.Winner {
			fillRow(gridY+6, gridY+9, winBg)
		}
		if m.HomeTeam.Winner {
			fillRow(gridY+11, gridY+14, winBg)
		}

		cellY := gridY + 6
		cellX = gridX + 1
		for i, val := range awayVals {
			drawText3x4CenteredInRange(c, cellX, colWidths[i], cellY, val, awayCol)
			cellX += colWidths[i]
		}
		cellY = gridY + 11
		cellX = gridX + 1
		for i, val := range homeVals {
			drawText3x4CenteredInRange(c, cellX, colWidths[i], cellY, val, homeCol)
			cellX += colWidths[i]
		}
	}

	if m.WinningPitcherLastName != "" && m.WinningPitcherLastName != "TBD" {
		DrawText(c, 1, gridY+gridH+1, trimToChars("W:"+m.WinningPitcherLastName, 16), green)
	}
	if m.LosingPitcherLastName != "" && m.LosingPitcherLastName != "TBD" {
		DrawText(c, 1, gridY+gridH+7, trimToChars("L:"+m.LosingPitcherLastName, 16), red)
	}
	if m.SavePitcherLastName != "" && m.SavePitcherLastName != "TBD" {
		DrawText(c, 1, gridY+gridH+13, trimToChars("S:"+m.SavePitcherLastName, 16), white)
	} else {
		drawTextCentered(c, gridY+gridH+13, formatNextMatchupDateTime(m.DateTime), grey)
	}
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
		" #   # ",
		"#     #",
		" #   # ",
		"  ###  ",
	}
	filledPat := []string{
		"  ###  ",
		" ##### ",
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
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Clear canvas
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	awayColor := GetTeamColor(game.AwayTeam.Name)
	homeColor := GetTeamColor(game.HomeTeam.Name)
	white := color.RGBA{R: 220, G: 220, B: 220, A: 255}
	grey := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	dimName := color.RGBA{R: 165, G: 165, B: 165, A: 255}
	orange := color.RGBA{R: 255, G: 150, B: 50, A: 255}

	// ===== COMPACT R/H/E/L GRID =====
	awayAbbr := teamAbbrev(game.AwayTeam.ShortName, game.AwayTeam.Name)
	homeAbbr := teamAbbrev(game.HomeTeam.ShortName, game.HomeTeam.Name)
	gridX := 1
	gridY := 1
	gridW := width - 2
	gridH := 17
	if gridW < 8 || height < gridH+1 {
		return
	}

	// Border and separators.
	for dx := 0; dx < gridW; dx++ {
		c.Set(gridX+dx, gridY, grey)
		c.Set(gridX+dx, gridY+gridH-1, grey)
	}
	for dy := 0; dy < gridH; dy++ {
		c.Set(gridX, gridY+dy, grey)
		c.Set(gridX+gridW-1, gridY+dy, grey)
	}

	colWidths := []int{14, 12, 12, 12, 12}
	colX := gridX + 1
	for i := 0; i < len(colWidths)-1; i++ {
		colX += colWidths[i]
		for dy := 1; dy < gridH-1; dy++ {
			c.Set(colX, gridY+dy, grey)
		}
	}
	h1 := gridY + 5
	h2 := gridY + 10
	for dx := 1; dx < gridW-1; dx++ {
		c.Set(gridX+dx, h1, grey)
		c.Set(gridX+dx, h2, grey)
	}

	// Header row: each stat letter centered in its own cell.
	statLabels := []string{"", "R", "H", "E", "L"}
	cellX := gridX + 1
	for i, label := range statLabels {
		drawText3x4CenteredInRange(c, cellX, colWidths[i], gridY+1, label, grey)
		cellX += colWidths[i]
	}

	awayVals := []string{awayAbbr, fmt.Sprintf("%d", game.AwayTeam.Runs), fmt.Sprintf("%d", game.AwayTeam.Hits), fmt.Sprintf("%d", game.AwayTeam.Errors), fmt.Sprintf("%d", game.AwayTeam.LOB)}
	homeVals := []string{homeAbbr, fmt.Sprintf("%d", game.HomeTeam.Runs), fmt.Sprintf("%d", game.HomeTeam.Hits), fmt.Sprintf("%d", game.HomeTeam.Errors), fmt.Sprintf("%d", game.HomeTeam.LOB)}
	cellY := gridY + 6
	cellX = gridX + 1
	for i, val := range awayVals {
		drawText3x4CenteredInRange(c, cellX, colWidths[i], cellY, val, awayColor)
		cellX += colWidths[i]
	}
	cellY = gridY + 11
	cellX = gridX + 1
	for i, val := range homeVals {
		drawText3x4CenteredInRange(c, cellX, colWidths[i], cellY, val, homeColor)
		cellX += colWidths[i]
	}

	// Inning line below the grid.
	inningLabel := fmt.Sprintf(game.HalfInning+" %d", game.Inning)
	drawInningOutsCentered(c, 1, width-2, gridY+gridH+2, inningLabel, game.Outs, grey, white, grey)

	// Pitcher and batter below the inning line, placed by team side.
	half := strings.ToLower(strings.TrimSpace(game.HalfInning))
	pitcherLeft := half == "bottom" || half == "end"
	leftX := 1
	rightX := width / 2
	halfW := width / 2
	if halfW > 0 {
		halfW--
	}

	pitcherName := "TBD"
	if game.CurrentPitcher.FullName != "" {
		pitcherName = compactLivePlayerName(game.CurrentPitcher.FullName, game.CurrentPitcher.LastName, 10)
	}
	pitcherDetail := strings.TrimSpace(fmt.Sprintf("%s%s", pitcherHandLabel(game.CurrentPitcher), trimToChars(game.CurrentPitcher.ERA, 5)))
	if pitcherDetail == "" {
		pitcherDetail = "TBD"
	}

	batterName := "TBD"
	if game.CurrentBatter.FullName != "" {
		batterName = compactLivePlayerName(game.CurrentBatter.FullName, game.CurrentBatter.LastName, 10)
	}
	batterDetail := strings.TrimSpace(fmt.Sprintf("%s%s", trimToChars(game.CurrentBatter.CurrentPosition, 2), trimToChars(game.CurrentBatter.SeasonBattingAverage, 5)))
	if batterDetail == "" {
		batterDetail = "TBD"
	}

	playerY1 := gridY + gridH + 8
	playerY2 := playerY1 + 6
	if pitcherLeft {
		drawText3x5CenteredInRange(c, leftX, halfW, playerY1, pitcherName, dimName)
		drawText3x5CenteredInRange(c, leftX, halfW, playerY2, pitcherDetail, grey)
		drawText3x5CenteredInRange(c, rightX, halfW, playerY1, batterName, dimName)
		drawText3x5CenteredInRange(c, rightX, halfW, playerY2, batterDetail, grey)
	} else {
		drawText3x5CenteredInRange(c, leftX, halfW, playerY1, batterName, dimName)
		drawText3x5CenteredInRange(c, leftX, halfW, playerY2, batterDetail, grey)
		drawText3x5CenteredInRange(c, rightX, halfW, playerY1, pitcherName, dimName)
		drawText3x5CenteredInRange(c, rightX, halfW, playerY2, pitcherDetail, grey)
	}

	drawBallsStrikesCountCentered(c, playerY2+6, game.Balls, game.Strikes, grey, grey, color.RGBA{R: 100, G: 200, B: 100, A: 255}, grey, color.RGBA{R: 255, G: 150, B: 50, A: 255})

	drawBasesY := playerY2 + 20
	if drawBasesY+7 < height {
		DrawBases(c, width/2, drawBasesY, &game.Bases)
	}

	lastPlayText := strings.TrimSpace(game.LastPlayNotation)
	if lastPlayText != "" {
		lastPlayText = "Prev: " + lastPlayText
		lastPlayY := drawBasesY + 3
		if lastPlayY+4 < height {
			drawText3x5CenteredInRange(c, 1, width-2, lastPlayY, trimToChars(lastPlayText, 14), orange)
		}
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
