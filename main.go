package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"
	"time"

	rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
	"golang.org/x/image/draw"

	//rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx // Will be used when goroutines are enabled

	controller := &scoreboard.DisplayController{}
	controller.Cond = sync.NewCond(&controller.Mu)
	controller.Paused = false

	//var wg sync.WaitGroup
	//wg.Add(1)
	//go scoreboard.StartScoreboard(ctx, &wg, controller)
	//wg.Add(1)
	//go radio.StartRadio(ctx, &wg, controller)

	//wg.Wait()

	fmt.Println("HERE WE GO")

	// Check if we should use the matrix emulator
	// Default is disabled for headless servers
	useEmulator := os.Getenv("USE_MATRIX_EMULATOR") == "1"

	if useEmulator && os.Getenv("DISPLAY") != "" {
		// Only initialize emulator if both conditions are met:
		// 1. USE_MATRIX_EMULATOR=1 is explicitly set
		// 2. DISPLAY environment variable is available (X11 is running)
		err := os.Setenv("MATRIX_EMULATOR", "1")
		if err != nil {
			panic(err)
		}

		m, _ := rgbmatrix.NewRGBLedMatrix(&rgbmatrix.DefaultConfig)
		c := rgbmatrix.NewCanvas(m)
		defer func() {
			_ = c.Close()
		}()

		draw.Draw(c, c.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		if err := c.Render(); err != nil {
			fmt.Printf("Error rendering canvas: %v\n", err)
		}

		time.Sleep(30 * time.Second)
	}

	//
	// make a config object
	// make a client
	// Should I bother with tests here?
	// Future me might like it and they won't be that expensive to mock
	// then if I put this up for future stuff it won't be a tragedy.

	// So I'll need to gothread the api stuff AND the radio stuff and see if I can pass both the controller for the matrix so I can override when I change the
	//	and the tuner

	// I think the next step is to get the basic api calls working and Identified what I need /want
	// I think let's focus on the stuff for the scoreboard in between games since we can do that for now
	// Then we can add a good test game or two and work on various displays for different plays

	// Once that is going on the actual board, we can work on the radio side of things

	// big question is do I want to do the board only stuff and THEN solder on the radio stuff?

	// Maybe not

}
