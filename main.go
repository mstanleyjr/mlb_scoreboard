//go:build pi

package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/internal/appconfig"
	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

var (
	canvas *rgbmatrix.Canvas
)

func main() {
	defaults := appconfig.DefaultConfig()
	configPath := flag.String("config", appconfig.DefaultPath, "path to startup config file")
	divisionStandingsMonochrome := flag.Bool("division-standings-monochrome", defaults.DivisionStandings.Monochrome, "render division standings text and dividers in light gray")
	divisionStandingsGreenBackground := flag.Bool("division-standings-green-background", defaults.DivisionStandings.GreenBackground, "render division standings with a dark scoreboard green background")
	flag.Parse()

	cfg, err := appconfig.Load(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	visited := visitedFlags(flag.CommandLine)
	monochrome := cfg.DivisionStandings.Monochrome
	if visited["division-standings-monochrome"] {
		monochrome = *divisionStandingsMonochrome
	}
	greenBackground := cfg.DivisionStandings.GreenBackground
	if visited["division-standings-green-background"] {
		greenBackground = *divisionStandingsGreenBackground
	}

	scoreboard.SetDivisionStandingsMonochrome(monochrome)
	scoreboard.SetDivisionStandingsGreenBackground(greenBackground)

	fmt.Println("HERE WE GOOOOO")

	// Create RGB LED matrix config
	config := &rgbmatrix.DefaultConfig
	config.DisableHardwarePulsing = true // Required until solder jumper is done
	config.Rows = 64
	config.Cols = 64
	config.HardwareMapping = "adafruit-hat"
	config.Brightness = 100

	fmt.Printf("Config: Rows=%d, Cols=%d, Mapping=%s, DisableHardwarePulsing=%v\n",
		config.Rows, config.Cols, config.HardwareMapping, config.DisableHardwarePulsing)

	// Create RGB LED matrix
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
			draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.Point{}, draw.Src)

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
			case scoreboard.DisplayTypeNextMatchup:
				if next, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardNextMatchup); ok {
					scoreboard.DrawNextMatchup(canvas, next)
				}
			case scoreboard.DisplayTypeLastMatchup:
				if last, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardLastMatchup); ok {
					scoreboard.DrawLastMatchup(canvas, last)
				}
			default:
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

func visitedFlags(fs *flag.FlagSet) map[string]bool {
	visited := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})
	return visited
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

// DrawLoadingScreen draws the loading screen while data is being fetched
func DrawLoadingScreen(c *rgbmatrix.Canvas) {
	bounds := c.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	text := "LOADING"
	textColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	startY := (height - 8) / 2
	if startY < 0 {
		startY = 0
	}

	scoreboard.DrawText5x8Centered(c, startY, text, textColor)
}
