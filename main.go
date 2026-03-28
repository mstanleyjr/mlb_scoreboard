package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

var (
	canvas *rgbmatrix.Canvas
	matrix *rgbmatrix.RGBLedMatrix
)

// Global display state for LED matrix - now defined in scoreboard package
// var (
// 	currentDisplayType DisplayType
// 	currentDisplayData interface{}
// 	displayMutex       sync.RWMutex
// )

// type DisplayType int

// const (
// 	DisplayTypeLoading DisplayType = iota
// 	DisplayTypeDivisionStandings
// 	DisplayTypeLiveGame
// 	DisplayTypeNextMatchup
// 	DisplayTypeLastMatchup
// )

func main() {
	fmt.Println("HERE WE GOOOOO")

	// Create RGB LED matrix config
	config := &rgbmatrix.DefaultConfig
	config.DisableHardwarePulsing = true // Required until solder jumper is done
	config.Rows = 32
	config.Cols = 64
	config.HardwareMapping = "adafruit-hat"
	config.Brightness = 100
	config.GPIOSlowdown = 3   // Pi 4B
	config.RowAddressType = 0 // 0=default for 32-row, change to 1 when using 64-row

	fmt.Printf("Config: Rows=%d, Cols=%d, Mapping=%s, DisableHardwarePulsing=%v\n",
		config.Rows, config.Cols, config.HardwareMapping, config.DisableHardwarePulsing)

	// Create RGB LED matrix
	var err error
	matrix, err := rgbmatrix.NewRGBLedMatrix(config)
	if err != nil {
		fmt.Printf("Error creating matrix: %v\n", err)
		return
	}
	fmt.Println("Matrix created successfully")

	// Create canvas
	canvas = rgbmatrix.NewCanvas(matrix)
	defer func() {
		err := canvas.Close()
		if err != nil {
			fmt.Println("Error closing canvas:", err)
		}
	}()

	fmt.Println("Canvas created, starting scoreboard...")

	// Set up context and wait group for goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Create display controller
	controller := &scoreboard.DisplayController{}
	controller.Cond = sync.NewCond(&controller.Mu)
	controller.Paused = false

	// Start scoreboard in background
	wg.Add(1)
	go scoreboard.StartScoreboard(ctx, &wg, controller)

	// Render loop - continuously update the display
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Clear canvas (black background)
			draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)

			// Draw current display content
			scoreboard.DisplayMutex.RLock()
			switch scoreboard.CurrentDisplayType {
			case scoreboard.DisplayTypeLoading:
				DrawLoadingScreen(canvas)
			case scoreboard.DisplayTypeDivisionStandings:
				if division, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardDivision); ok {
					scoreboard.DrawDivisionStandings(canvas, division)
				}
			case scoreboard.DisplayTypeLiveGame:
				if game, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardLiveGame); ok {
					scoreboard.DrawLiveGameScore(canvas, game)
				}
			default:
				// Draw test pattern as fallback
				DrawTestPattern(canvas)
			}
			scoreboard.DisplayMutex.RUnlock()

			// Render to LED matrix
			err := canvas.Render()
			if err != nil {
				fmt.Println("Error rendering to matrix:", err)
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Wait for all goroutines to complete or context cancellation
	wg.Wait()

	fmt.Println("Scoreboard shutdown complete")
}

// DrawTestPattern draws a simple test pattern to verify the display is working
func DrawTestPattern(c *rgbmatrix.Canvas) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Draw a red border
	for x := 0; x < width; x++ {
		c.Set(x, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		c.Set(x, height-1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	for y := 0; y < height; y++ {
		c.Set(0, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		c.Set(width-1, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}

	// Draw a green square in the center
	centerX := width / 2
	centerY := height / 2
	squareSize := 8
	for x := centerX - squareSize/2; x < centerX+squareSize/2; x++ {
		for y := centerY - squareSize/2; y < centerY+squareSize/2; y++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				c.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
			}
		}
	}

	// Draw a blue dot in the very center
	c.Set(centerX, centerY, color.RGBA{R: 0, G: 0, B: 255, A: 255})
}

// DrawText would be a helper to draw text to the canvas
func DrawText(c *rgbmatrix.Canvas, x, y int, text string, col color.RGBA) {
	// TODO: Implement text rendering
	// This would require a font library or bitmap fonts
}

// DrawLoadingScreen draws the loading screen while data is being fetched
func DrawLoadingScreen(c *rgbmatrix.Canvas) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Clear canvas
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	// Draw "LOADING" text in the center
	loadingText := "LOADING"
	textColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	// Center the text
	textWidth := len(loadingText) * 4 // Approximate width
	startX := (width - textWidth) / 2
	startY := height / 2

	for i, _ := range loadingText {
		x := startX + (i * 4)
		y := startY

		// Draw a simple character box
		for j := 0; j < 3; j++ {
			c.Set(x+j, y, textColor)
			c.Set(x+j, y+4, textColor)
		}
		for j := 0; j < 5; j++ {
			c.Set(x, y+j, textColor)
			c.Set(x+2, y+j, textColor)
		}
	}

	// Draw animated dots
	dotY := startY + 8
	for i := 0; i < 3; i++ {
		dotX := startX + textWidth + 4 + (i * 6)
		c.Set(dotX, dotY, textColor)
		c.Set(dotX+1, dotY, textColor)
		c.Set(dotX, dotY+1, textColor)
		c.Set(dotX+1, dotY+1, textColor)
	}
}
