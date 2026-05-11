package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
)

const (
	matrixW = 64
	matrixH = 64
)

type Matrix struct {
	w, h int
	pix  []color.RGBA
}

func NewMatrix(w, h int) *Matrix {
	return &Matrix{w: w, h: h, pix: make([]color.RGBA, w*h)}
}

func (m *Matrix) Clear(c color.RGBA) {
	for i := range m.pix {
		m.pix[i] = c
	}
}

func (m *Matrix) Set(x, y int, c color.Color) {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return
	}
	r, g, b, a := c.RGBA()
	m.pix[y*m.w+x] = color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func (m *Matrix) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.w, m.h)
}

func (m *Matrix) At(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return color.RGBA{}
	}
	return m.pix[y*m.w+x]
}

func dominantChar(c color.RGBA, empty rune) rune {
	if c.R == 0 && c.G == 0 && c.B == 0 {
		return empty
	}
	if c.R >= c.G && c.R >= c.B {
		return 'R'
	}
	if c.G >= c.R && c.G >= c.B {
		return 'G'
	}
	return 'B'
}

func (m *Matrix) String(empty rune) string {
	var out strings.Builder
	for y := 0; y < m.h; y++ {
		var row strings.Builder
		row.Grow(m.w)
		for x := 0; x < m.w; x++ {
			row.WriteRune(dominantChar(m.At(x, y), empty))
		}
		out.WriteString(row.String())
		out.WriteByte('\n')
	}
	return out.String()
}

func drawGrid(m *Matrix) {
	gridColor := color.RGBA{R: 0, G: 0, B: 200, A: 255}
	for x := 0; x < matrixW; x += 8 {
		for y := 0; y < matrixH; y++ {
			m.Set(x, y, gridColor)
		}
	}
	for y := 0; y < matrixH; y += 8 {
		for x := 0; x < matrixW; x++ {
			m.Set(x, y, gridColor)
		}
	}
}

func drawRulerY(m *Matrix, y int) {
	rulerColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for x := 0; x < matrixW; x++ {
		m.Set(x, y, rulerColor)
	}
}

func drawLoadingScreen(m *Matrix) {
	m.Clear(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	text := "LOADING"
	textWidth := len(text) * 4 // 3x5 font with 1px gap
	x := (matrixW - textWidth) / 2
	y := (matrixH - 5) / 2
	if x < 0 {
		x = 0
	}
	scoreboard.DrawTextSmall(m, x, y, text, color.RGBA{R: 255, G: 255, B: 0, A: 255})
}

func drawTestPattern(m *Matrix) {
	m.Clear(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	for x := 0; x < matrixW; x++ {
		m.Set(x, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		m.Set(x, matrixH-1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	for y := 0; y < matrixH; y++ {
		m.Set(0, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		m.Set(matrixW-1, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
}

func renderCurrentDisplay(m *Matrix) {
	m.Clear(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	scoreboard.DisplayMutex.RLock()
	defer scoreboard.DisplayMutex.RUnlock()

	switch scoreboard.CurrentDisplayType {
	case scoreboard.DisplayTypeLoading:
		drawLoadingScreen(m)
	case scoreboard.DisplayTypeDivisionStandings:
		if division, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardDivision); ok {
			scoreboard.DrawDivisionStandings(m, division)
		}
	case scoreboard.DisplayTypeLiveGame:
		if game, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardLiveGame); ok {
			scoreboard.DrawLiveGameScore(m, game)
		}
	case scoreboard.DisplayTypeNextMatchup:
		if next, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardNextMatchup); ok {
			scoreboard.DrawNextMatchup(m, next)
		}
	case scoreboard.DisplayTypeLastMatchup:
		if last, ok := scoreboard.CurrentDisplayData.(scoreboard.ScoreboardLastMatchup); ok {
			scoreboard.DrawLastMatchup(m, last)
		}
	default:
		drawTestPattern(m)
	}
}

func main() {
	empty := flag.String("empty", " ", "character for empty pixels (e.g. ' ' or '-')")
	fps := flag.Int("fps", 3, "terminal refresh rate (frames per second)")
	showGrid := flag.Bool("grid", false, "overlay an 8x8 debug grid")
	rulerY := flag.Int("ruler-y", -1, "draw a horizontal ruler at this y (0-63), -1 to disable")
	flag.Parse()

	emptyRune := ' '
	if *empty != "" {
		emptyRune = []rune(*empty)[0]
	}
	if *fps < 1 {
		*fps = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	var wg sync.WaitGroup
	controller := &scoreboard.DisplayController{}
	controller.Cond = sync.NewCond(&controller.Mu)
	controller.Paused = false

	wg.Add(1)
	go scoreboard.StartScoreboard(ctx, &wg, controller)

	m := NewMatrix(matrixW, matrixH)
	ticker := time.NewTicker(time.Second / time.Duration(*fps))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case <-ticker.C:
			renderCurrentDisplay(m)
			if *showGrid {
				drawGrid(m)
			}
			if *rulerY >= 0 && *rulerY < matrixH {
				drawRulerY(m, *rulerY)
			}

			// Clear terminal + move cursor home, then print a full frame.
			fmt.Print("\x1b[H\x1b[2J")
			fmt.Print(m.String(emptyRune))
		}
	}
}
