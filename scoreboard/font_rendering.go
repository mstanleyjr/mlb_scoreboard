package scoreboard

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// TrueTypeFont manages rendering text with TrueType fonts
type TrueTypeFont struct {
	face font.Face
}

// NewTrueTypeFont loads a TrueType font from a file
// pointSize is the font size in points (e.g., 8, 10, 12)
func NewTrueTypeFont(fontPath string, pointSize float64) (*TrueTypeFont, error) {
	// Read font file
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	// Parse font
	parsed, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	// Create face with DPI=72 (standard screen DPI)
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    pointSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	return &TrueTypeFont{face: face}, nil
}

// DrawString draws text using the TrueType font on the canvas
func (tf *TrueTypeFont) DrawString(c *rgbmatrix.Canvas, x, y int, text string, col color.Color) error {
	bounds := c.Bounds()

	// Create a temporary RGBA image to render text
	img := image.NewRGBA(bounds)

	// Draw text onto temporary image
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: tf.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6((y + 10) * 64)}, // Adjust Y for baseline
	}
	drawer.DrawString(text)

	// Copy rendered text pixels from temp image to canvas
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			pixel := img.At(px, py)
			r, g, b, a := pixel.RGBA()
			// Only copy non-transparent pixels
			if a > 0 {
				c.Set(px, py, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
			}
		}
	}

	return nil
}

// Close releases resources
func (tf *TrueTypeFont) Close() error {
	if tf.face != nil {
		_ = tf.face.Close()
	}
	return nil
}

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
