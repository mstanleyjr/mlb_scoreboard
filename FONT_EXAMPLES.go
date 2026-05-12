package main

import (
	"fmt"
	"image/color"

	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
)

// EXAMPLE 1: Using custom bitmap glyphs
func ExampleCustomGlyphs(c scoreboard.PixelCanvas) {
	// Create a glyph builder with 5x7 characters
	gb := scoreboard.NewGlyphBuilder(5, 7)

	// Define custom glyphs using # and space
	gb.AddGlyph('A', []string{
		"  #  ",
		" # # ",
		"#   #",
		"#####",
		"#   #",
		"#   #",
		"#   #",
	})

	gb.AddGlyph('B', []string{
		"#### ",
		"#   #",
		"#### ",
		"#   #",
		"#   #",
		"#   #",
		"#### ",
	})

	// Draw text using custom glyphs
	gb.DrawString(c, 10, 5, "AB", color.RGBA{R: 255, G: 100, B: 100, A: 255})
}

// EXAMPLE 2: Using TrueType fonts
func ExampleTrueTypeFont(c scoreboard.PixelCanvas) {
	// Load a TrueType font from disk
	// On Pi: /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf
	// On Mac: /Library/Fonts/Arial.ttf
	ttf, err := scoreboard.NewTrueTypeFont("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 8)
	if err != nil {
		fmt.Printf("Error loading font: %v\n", err)
		return
	}
	defer ttf.Close()

	// Draw text
	ttf.DrawString(c, 10, 10, "MLB Score", color.RGBA{R: 0, G: 255, B: 0, A: 255})
}

// EXAMPLE 3: Creating a scoreboard-specific font
func ExampleScoreboardFont(c scoreboard.PixelCanvas) {
	// Use the pre-built scoreboard font
	scoreboardFont := scoreboard.NewScoreboardFont()

	// Draw game score
	scoreboardFont.DrawString(c, 5, 20, "3-2", color.RGBA{R: 100, G: 255, B: 100, A: 255})
}
