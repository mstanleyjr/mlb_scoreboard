package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"time"

	rgbmatrix "github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
)

func main() {
	//ctx, _ := context.WithCancel(context.Background())

	controller := &scoreboard.DisplayController{}
	controller.Cond = sync.NewCond(&controller.Mu)
	controller.Paused = false

	//var wg sync.WaitGroup
	////wg.Add(1)
	////go scoreboard.StartScoreboard(ctx, &wg, controller)
	////wg.Add(1)
	////go radio.StartRadio(ctx, &wg, controller)
	//// This would be where I do the radio listener to check the button
	//
	//wg.Wait()
	//
	//fmt.Println("HERE WE GO`")
	//err := os.Setenv("MATRIX_EMULATOR", "1")
	//if err != nil {
	//	panic(err)
	//}

	m, _ := rgbmatrix.NewRGBLedMatrix(&rgbmatrix.DefaultConfig)
	c := rgbmatrix.NewCanvas(m)
	defer func(c *rgbmatrix.Canvas) {
		err := c.Close()
		if err != nil {
			fmt.Println("Error closing canvas")
			return
		}
	}(c)

	draw.Draw(c, c.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	err := c.Render()
	if err != nil {
		fmt.Println("Error rendering canvas")
		return
	}

	time.Sleep(15 * time.Second)

	//
	// make a config object
	// make a client
	// Should I bother with tests here?
	// Future me might like it and they won't be that expensive to mock
	// then if I put this up for future stuff it won't be a tragedy.
}
