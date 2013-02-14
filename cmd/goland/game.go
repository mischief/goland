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

type State struct {
  X, V float64
}

type Derivative struct {
  DX, DV float64
}

func interpolate(previous, current State, alpha float64) (s State) {
  s.X = current.X * alpha + previous.X * (1 - alpha)
  s.V = current.V * alpha + previous.V * (1 - alpha)
  return
}

func acceleration(state State, t float64) float64 {
  k := 10.0
  b := 1.0
  return - k*state.X - b*state.V
}

func evaluate1(initial State, t float64) (der Derivative) {
  der.DX = initial.V
  der.DV = acceleration(initial, t)
  return
}

func evaluate2(initial State, t, dt float64, d Derivative) (der Derivative) {
  s := State{X: initial.X + d.DX * dt, V: initial.V + d.DV * dt}

  der.DX = s.V
  der.DV = acceleration(s, t+dt)
  return
}

func integrate(state State, t, dt float64) (s State) {
  a := evaluate1(state, t)
  b := evaluate2(state, t, dt * 0.5, a)
  c := evaluate2(state, t, dt * 0.5, b)
  d := evaluate2(state, t, dt, c)

  dxdt := 1.0/6.0 * (a.DX + 2.0 * (b.DX + c.DX) + d.DX)
  dvdt := 1.0/6.0 * (a.DV + 2.0 * (b.DV + c.DV) + d.DV)

  s.X = state.X + dxdt * dt
  s.V = state.V + dvdt * dt
  return
}

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

  current := State{X: 20, V: 10}
  previous := current

  t := 0.0
  dt := 0.1
  accumulator := 0.0

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

      accumulator += delta.Seconds()

      for accumulator >= dt {
        accumulator -= dt
        previous = current
        current = integrate(current, t, dt)
        t += dt
      }

      state := interpolate(previous, current, accumulator/dt)

      g.Update(delta)
      g.Draw()

      g.Print(25+int(state.X), 10, termbox.ColorGreen, termbox.ColorBlack, "@")


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

