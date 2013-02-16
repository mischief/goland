package main

import (
  "time"
  "flag"
  "os"
  "log"
  "github.com/nsf/termbox-go"
)

const (
  NumFPSSamples = 64
  FPSLimit = 60
)

var (
  fpsSamples [64]float64
  currentSample = 0
  logfile = flag.String("log", "goland.log", "log file")
)

type Game struct {
  P           Player

  Terminal
  logfile     *os.File

  CloseChan   chan bool

  // unexported
  fps         float64
}

func NewGame() *Game {
  g := Game{}

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
  f, err := os.OpenFile(*logfile, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
  if err != nil {
    log.Fatal(err)
  }

  log.SetOutput(f)
  log.Print("Logging started")

  g.logfile = f

  g.Terminal.Start()
}

func (g *Game) End() {
  log.Print("Logging ended")
  g.logfile.Close()
  g.Terminal.End()
}

func (g *Game) Update(delta time.Duration) bool {
  // update fps
  g.fps = g.CalcFPS(delta)

  // poll event channels
  select {
  case ev := <- g.EventChan:
    log.Printf("termbox: %v", ev)
    // handle term event
    if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
      g.CloseChan <- false
      return false
    }
  default:
  }

  return true
}

func(g *Game) CalcFPS(delta time.Duration) float64 {
  fpsSamples[currentSample % NumFPSSamples] = 1.0 / delta.Seconds()
  currentSample++
  fps := 0.0

  for i := 0; i < NumFPSSamples; i++ {
    fps += fpsSamples[i]
  }

  fps /= NumFPSSamples

  return fps
}

func (g *Game) Draw() {
  g.Terminal.Draw()

  //fps := g.CalcFPS(delta)
  g.Printf(0, 0, termbox.ColorRed, termbox.ColorBlack, "FPS: %f", g.fps)

}

