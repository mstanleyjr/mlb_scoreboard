package scoreboard

import (
	"fmt"
	"image/color"
	"log"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// GlyphBuilder helps you create custom bitmap glyphs
// You define each glyph as a 2D grid of true/false values
type GlyphBuilder struct {
	Width  int
	Height int
	Glyphs map[rune][][]bool
}

// NewGlyphBuilder creates a new glyph builder
func NewGlyphBuilder(width, height int) *GlyphBuilder {
	return &GlyphBuilder{
		Width:  width,
		Height: height,
		Glyphs: make(map[rune][][]bool),
	}
}

// AddGlyph adds a glyph to the builder from a string representation
// Use '#' for pixels that are on, ' ' for pixels that are off
// Example:
//
//	gb.AddGlyph('A', []string{
//	    "  #  ",
//	    " # # ",
//	    "#   #",
//	    "#####",
//	    "#   #",
//	})
func (gb *GlyphBuilder) AddGlyph(ch rune, pattern []string) error {
	if len(pattern) != gb.Height {
		return fmt.Errorf("pattern height %d doesn't match builder height %d", len(pattern), gb.Height)
	}

	glyph := make([][]bool, gb.Height)
	for i, line := range pattern {
		if len(line) != gb.Width {
			return fmt.Errorf("pattern line %d width %d doesn't match builder width %d", i, len(line), gb.Width)
		}
		glyph[i] = make([]bool, gb.Width)
		for j, ch := range line {
			glyph[i][j] = ch == '#'
		}
	}

	gb.Glyphs[ch] = glyph
	return nil
}

// DrawGlyph draws a single glyph on the canvas
func (gb *GlyphBuilder) DrawGlyph(c *rgbmatrix.Canvas, x, y int, ch rune, col color.Color) {
	glyph, ok := gb.Glyphs[ch]
	if !ok {
		return // Character not found, skip
	}

	for row := 0; row < gb.Height; row++ {
		for colIdx := 0; colIdx < gb.Width; colIdx++ {
			if glyph[row][colIdx] {
				px := x + colIdx
				py := y + row
				bounds := c.Bounds()
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					c.Set(px, py, col)
				}
			}
		}
	}
}

// DrawString draws a string using custom glyphs
func (gb *GlyphBuilder) DrawString(c *rgbmatrix.Canvas, x, y int, text string, col color.Color) {
	cx := x
	for _, ch := range text {
		gb.DrawGlyph(c, cx, y, ch, col)
		cx += gb.Width + 1 // width + 1 pixel spacing
	}
}

// --- Pre-built glyph sets ---

// NewScoreboardFont creates a compact 4x7 font optimized for scoreboards
func NewScoreboardFont() *GlyphBuilder {
	gb := NewGlyphBuilder(4, 7)

	// Simple scoreboard digits and common characters
	patterns := map[rune][]string{
		'0': {
			"###",
			"# #",
			"# #",
			"# #",
			"# #",
			"# #",
			"###",
		},
		'1': {
			"  #",
			" ##",
			"  #",
			"  #",
			"  #",
			"  #",
			"###",
		},
		'2': {
			"###",
			"  #",
			"###",
			"#  ",
			"#  ",
			"#  ",
			"###",
		},
		'3': {
			"###",
			"  #",
			"###",
			"  #",
			"  #",
			"  #",
			"###",
		},
		'-': {
			"   ",
			"   ",
			"###",
			"   ",
			"   ",
			"   ",
			"   ",
		},
		':': {
			"   ",
			" # ",
			"   ",
			"   ",
			" # ",
			"   ",
			"   ",
		},
	}

	for ch, pattern := range patterns {
		if err := gb.AddGlyph(ch, pattern); err != nil {
			log.Printf("Failed to add glyph %c: %v", ch, err)
		}
	}

	return gb
}
