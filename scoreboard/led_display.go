package scoreboard

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// DrawDivisionStandings renders division standings to the LED matrix
func DrawDivisionStandings(c *rgbmatrix.Canvas, division ScoreboardDivision) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Clear canvas with black background
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	// Draw title (division name) at top
	DrawTextSmall(c, 2, 2, division.LeagueName, color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Draw team standings
	startY := 12
	lineHeight := 8
	for i, team := range division.Teams {
		y := startY + (i * lineHeight)
		if y > height-8 { // Leave space at bottom
			break // Don't draw off screen
		}

		// Team name (abbreviated)
		teamColor := GetTeamColor(team.Name)
		teamName := team.Name
		if len(teamName) > 8 {
			teamName = teamName[:8] // Truncate long names
		}
		DrawTextSmall(c, 2, y, teamName, teamColor)

		// Record (W-L)
		record := team.Record
		recordStr := ""
		recordStr += formatInt(record.Wins, 2)
		recordStr += "-"
		recordStr += formatInt(record.Losses, 2)
		DrawTextSmall(c, 40, y, recordStr, color.RGBA{R: 100, G: 200, B: 100, A: 255})
	}
}

// DrawLiveGameScore renders a live game score to the LED matrix
func DrawLiveGameScore(c *rgbmatrix.Canvas, game ScoreboardLiveGame) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Clear canvas
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	// Draw away team on left
	awayColor := GetTeamColor(game.AwayTeam.Name)
	DrawTextSmall(c, 2, 5, game.AwayTeam.ShortName, awayColor)
	DrawLargeScore(c, 2, 15, game.AwayTeam.Runs)

	// Draw home team on right
	homeColor := GetTeamColor(game.HomeTeam.Name)
	DrawTextSmall(c, 40, 5, game.HomeTeam.ShortName, homeColor)
	DrawLargeScore(c, 40, 15, game.HomeTeam.Runs)

	// Draw inning/status in middle
	inningStr := formatInt(game.Inning, 1) + ":" + game.HalfInning
	DrawTextSmall(c, 20, 25, inningStr, color.RGBA{R: 200, G: 200, B: 200, A: 255})

	// Draw count (balls/strikes)
	countStr := formatInt(game.Balls, 1) + "-" + formatInt(game.Strikes, 1)
	DrawTextSmall(c, 20, 30, countStr, color.RGBA{R: 255, G: 150, B: 50, A: 255})

	// Draw bases
	DrawBases(c, width/2, height-5, &game.Bases)
}

// DrawLargeScore draws a two-digit score in a larger format
func DrawLargeScore(c *rgbmatrix.Canvas, x, y int, score int) {
	tens := score / 10
	ones := score % 10

	// Draw tens digit
	DrawBigDigit(c, x, y, tens)
	// Draw ones digit
	DrawBigDigit(c, x+12, y, ones)
}

// DrawBigDigit draws a single digit in larger format (roughly 10x10)
func DrawBigDigit(c *rgbmatrix.Canvas, x, y int, digit int) {
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

// DrawBases draws the bases (simplified diamond shape)
func DrawBases(c *rgbmatrix.Canvas, cx, cy int, bases *ScoreboardLiveGameBases) {
	// Draw diamond outline
	emptyCol := color.RGBA{R: 50, G: 50, B: 50, A: 255}
	c.Set(cx, cy-3, emptyCol) // Home
	c.Set(cx+3, cy, emptyCol) // First
	c.Set(cx, cy+3, emptyCol) // Second
	c.Set(cx-3, cy, emptyCol) // Third

	filledCol := color.RGBA{R: 200, G: 100, B: 0, A: 255}

	// Draw occupied bases
	if bases != nil {
		if bases.First {
			c.Set(cx+3, cy, filledCol)
		}
		if bases.Second {
			c.Set(cx, cy+3, filledCol)
		}
		if bases.Third {
			c.Set(cx-3, cy, filledCol)
		}
	}
}

// DrawText draws text at position (x, y) with given color
// Note: This is a placeholder - real text rendering would require a font
func DrawText(c *rgbmatrix.Canvas, x, y int, text string, col color.RGBA) {
	// For now, just draw a colored line for each character
	// In production, you'd want to use a bitmap font library
	for i := 0; i < len(text) && i < 8; i++ {
		// Draw a simple 3x5 character representation
		for j := 0; j < 3; j++ {
			c.Set(x+(i*4)+j, y, col)
			c.Set(x+(i*4)+j, y+4, col)
		}
		for j := 0; j < 5; j++ {
			c.Set(x+(i*4), y+j, col)
			c.Set(x+(i*4)+2, y+j, col)
		}
	}
}

// DrawTextSmall draws small text at position (x, y) with given color
func DrawTextSmall(c *rgbmatrix.Canvas, x, y int, text string, col color.RGBA) {
	// For small text, use a simpler 3x5 dot matrix
	for i := 0; i < len(text) && i < 8; i++ {
		for j := 0; j < 5; j++ {
			if j == 2 {
				// Middle column always on
				c.Set(x+(i*4)+2, y+j, col)
			} else {
				// Randomly turn on some dots for effect
				if (i+j)%2 == 0 {
					c.Set(x+(i*4)+j, y+j, col)
				}
			}
		}
	}
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
