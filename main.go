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

func main() {
	fmt.Println("HERE WE GOOOOO")

	// Create RGB LED matrix config with hardware pulse disabled
	config := &rgbmatrix.DefaultConfig
	config.DisableHardwarePulsing = true
	config.Rows = 32
	config.Cols = 64
	config.HardwareMapping = "adafruit-hat"

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

			// TODO: Draw scoreboard content here
			// For now, draw a test pattern
			DrawTestPattern(canvas)

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
	// Draw a red border
	for x := 0; x < 64; x++ {
		c.Set(x, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		c.Set(x, 31, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	for y := 0; y < 64; y++ {
		c.Set(0, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		c.Set(63, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}

	// Draw a green square in the center
	for x := 20; x < 44; x++ {
		for y := 10; y < 22; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	// Draw a blue dot in the very center
	c.Set(32, 16, color.RGBA{R: 0, G: 0, B: 255, A: 255})
}

// DrawText would be a helper to draw text to the canvas
func DrawText(c *rgbmatrix.Canvas, x, y int, text string, col color.RGBA) {
	// TODO: Implement text rendering
	// This would require a font library or bitmap fonts
}
