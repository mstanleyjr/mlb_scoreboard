package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

func main() {
	//ctx, _ := context.WithCancel(context.Background())

	//controller := &scoreboard.DisplayController{}
	//controller.Cond = sync.NewCond(&controller.Mu)
	//controller.Paused = false

	fmt.Println("HERE WE GOOOOO")

	// Create RGB LED matrix config with hardware pulse disabled
	config := &rgbmatrix.DefaultConfig
	config.DisableHardwarePulse = true

	// Create RGB LED matrix
	m, err := rgbmatrix.NewRGBLedMatrix(config)
	if err != nil {
		fmt.Printf("Error creating matrix: %v\n", err)
		return
	}
	fmt.Println("Matrix created successfully")

	// Create canvas
	c := rgbmatrix.NewCanvas(m)
	defer func() {
		err := c.Close()
		if err != nil {
			fmt.Println("Error closing canvas:", err)
		}
	}()

	fmt.Println("Canvas created, rendering colors...")

	// Test 1: Red
	fmt.Println("Drawing RED...")
	draw.Draw(c, c.Bounds(), &image.Uniform{color.RGBA{R: 255, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)
	err = c.Render()
	if err != nil {
		fmt.Println("Error rendering RED:", err)
		return
	}
	time.Sleep(5 * time.Second)

	// Test 2: Green
	fmt.Println("Drawing GREEN...")
	draw.Draw(c, c.Bounds(), &image.Uniform{color.RGBA{R: 0, G: 255, B: 0, A: 255}}, image.ZP, draw.Src)
	err = c.Render()
	if err != nil {
		fmt.Println("Error rendering GREEN:", err)
		return
	}
	time.Sleep(5 * time.Second)

	// Test 3: Blue
	fmt.Println("Drawing BLUE...")
	draw.Draw(c, c.Bounds(), &image.Uniform{color.RGBA{R: 0, G: 0, B: 255, A: 255}}, image.ZP, draw.Src)
	err = c.Render()
	if err != nil {
		fmt.Println("Error rendering BLUE:", err)
		return
	}
	time.Sleep(5 * time.Second)

	// Clear
	fmt.Println("Clearing...")
	draw.Draw(c, c.Bounds(), &image.Uniform{color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)
	err = c.Render()
	if err != nil {
		fmt.Println("Error rendering BLACK:", err)
		return
	}

	fmt.Println("Test complete")

	//
	// make a config object
	// make a client
	// Should I bother with tests here?
	// Future me might like it and they won't be that expensive to mock
	// then if I put this up for future stuff it won't be a tragedy.
}
