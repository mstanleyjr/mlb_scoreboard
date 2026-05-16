package termsim

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

type Options struct {
	Empty                       string
	FPS                         int
	ShowGrid                    bool
	RulerY                      int
	DivisionStandingsMonochrome bool
	DivisionStandingsGreenBG    bool
}

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

func RunCLI(args []string) error {
	fs := flag.NewFlagSet("mlb_scoreboard", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	empty := fs.String("empty", " ", "character for empty pixels (e.g. ' ' or '-')")
	fps := fs.Int("fps", 3, "terminal refresh rate (frames per second)")
	showGrid := fs.Bool("grid", false, "overlay an 8x8 debug grid")
	rulerY := fs.Int("ruler-y", -1, "draw a horizontal ruler at this y (0-63), -1 to disable")
	divisionStandingsMonochrome := fs.Bool("division-standings-monochrome", false, "render division standings text and dividers in light gray")
	divisionStandingsGreenBG := fs.Bool("division-standings-green-background", false, "render division standings with a dark scoreboard green background")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return Run(context.Background(), Options{
		Empty:                       *empty,
		FPS:                         *fps,
		ShowGrid:                    *showGrid,
		RulerY:                      *rulerY,
		DivisionStandingsMonochrome: *divisionStandingsMonochrome,
		DivisionStandingsGreenBG:    *divisionStandingsGreenBG,
	})
}

func Run(parent context.Context, opts Options) error {
	ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	emptyRune := ' '
	if opts.Empty != "" {
		emptyRune = []rune(opts.Empty)[0]
	}
	if opts.FPS < 1 {
		opts.FPS = 1
	}
	scoreboard.SetDivisionStandingsMonochrome(opts.DivisionStandingsMonochrome)
	scoreboard.SetDivisionStandingsGreenBackground(opts.DivisionStandingsGreenBG)

	var wg sync.WaitGroup
	controller := &scoreboard.DisplayController{}
	controller.Cond = sync.NewCond(&controller.Mu)
	controller.Paused = false

	wg.Add(1)
	go scoreboard.StartScoreboard(ctx, &wg, controller)

	m := NewMatrix(matrixW, matrixH)
	ticker := time.NewTicker(time.Second / time.Duration(opts.FPS))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case <-ticker.C:
			renderCurrentDisplay(m)
			if opts.ShowGrid {
				drawGrid(m)
			}
			if opts.RulerY >= 0 && opts.RulerY < matrixH {
				drawRulerY(m, opts.RulerY)
			}

			fmt.Print("\x1b[H\x1b[2J")
			fmt.Print(m.String(emptyRune))
		}
	}
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
	y := (matrixH - 8) / 2
	if y < 0 {
		y = 0
	}
	scoreboard.DrawText5x8Centered(m, y, text, color.RGBA{R: 255, G: 255, B: 0, A: 255})
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
