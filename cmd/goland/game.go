package main

import (
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
	"os"
	"time"
)

const (
	NumFPSSamples = 64
	FPSLimit      = 60
)

var (
	fpsSamples    [64]float64
	currentSample = 0
	logfile       = flag.String("log", "goland.log", "log file")
	debug         = flag.Bool("debug", false, "print debugging info")
)

type Game struct {
	P *Player

	Terminal
	logfile *os.File

	CloseChan chan bool

	// unexported
	fps float64

	Objects []Object
}

func NewGame() *Game {
	g := Game{}

	g.P = NewPlayer()

	g.Objects = append(g.Objects, g.P)

	g.CloseChan = make(chan bool, 1)

	return &g
}

func (g *Game) Run() {

	g.Start()

	timer := NewDeltaTimer()
	ticker := time.NewTicker(time.Second / FPSLimit)

	run := true

	for run {
		select {
		case <-ticker.C:
			// frame tick
			delta := timer.DeltaTime()

			if delta.Seconds() > 0.25 {
				delta = time.Duration(250 * time.Millisecond)
			}

			g.Update(delta)
			g.Draw()

			g.Flush()

		case <-g.CloseChan:
			run = false
		}
	}

	g.End()

}

func (g *Game) Start() {
	f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
	log.Print("Logging started")

	g.logfile = f

	g.Terminal.Start()

	g.HandleKey(termbox.KeyEsc, func(ev termbox.Event) { g.CloseChan <- false })

	scale := 2

	g.HandleRune('w', func(_ termbox.Event) { g.P.Move(DIR_UP) })
	g.HandleRune('a', func(_ termbox.Event) { g.P.Move(DIR_LEFT) })
	g.HandleRune('s', func(_ termbox.Event) { g.P.Move(DIR_DOWN) })
	g.HandleRune('d', func(_ termbox.Event) { g.P.Move(DIR_RIGHT) })
	g.HandleRune('W', func(_ termbox.Event) {
		for i := 0; i < scale; i++ {
			g.P.Move(DIR_UP)
		}
	})
	g.HandleRune('A', func(_ termbox.Event) {
		for i := 0; i < scale; i++ {
			g.P.Move(DIR_LEFT)
		}
	})
	g.HandleRune('S', func(_ termbox.Event) {
		for i := 0; i < scale; i++ {
			g.P.Move(DIR_DOWN)
		}
	})
	g.HandleRune('D', func(_ termbox.Event) {
		for i := 0; i < scale; i++ {
			g.P.Move(DIR_RIGHT)
		}
	})

	g.P.Pos = Vector{10, 10}

}

func (g *Game) End() {
	log.Print("Logging ended")
	g.logfile.Close()
	g.Terminal.End()
}

func (g *Game) Update(delta time.Duration) {
	// update fps
	g.fps = g.calcFPS(delta)

	g.RunInputHandlers()

	for _, o := range g.Objects {
		o.Update(delta)
	}

}

func (g *Game) Draw() {
	g.Terminal.Draw()

	labelparams := &tulib.LabelParams{termbox.ColorRed, termbox.ColorBlack, tulib.AlignCenter, '»', false}
	labelrect := tulib.Rect{1, 0, 12, 1}

	g.Terminal.DrawLabel(labelrect, labelparams, []byte(fmt.Sprintf(" FPS: %5.2f ", g.fps)))

	fpsrect := tulib.Rect{14, 0, 9, 1}

	g.Terminal.DrawLabel(fpsrect, labelparams, []byte(fmt.Sprintf(" %dx%d ", g.Terminal.Rect.Width, g.Terminal.Rect.Height)))

	for _, o := range g.Objects {
		o.Draw(g)
	}

	//fps := g.CalcFPS(delta)
	//g.Printf(0, 0, termbox.ColorRed, termbox.ColorBlack, "FPS: %f", g.fps)

}

func (g *Game) calcFPS(delta time.Duration) float64 {
	fpsSamples[currentSample%NumFPSSamples] = 1.0 / delta.Seconds()
	currentSample++
	fps := 0.0

	for i := 0; i < NumFPSSamples; i++ {
		fps += fpsSamples[i]
	}

	fps /= NumFPSSamples

	return fps
}
