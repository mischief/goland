package main

import (
	"fmt"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"runtime"
	"time"
)

const (
	FPS_SAMPLES = 64
)

type StatsPanel struct {
	do chan func(*StatsPanel)
	*graphics.BasePanel

	g *Game

	// displayed data
	w, h     int // terminal size
	samples  [64]float64
	current  int
	memstats runtime.MemStats
	numgr    int

	fps float64
}

func NewStatsPanel(g *Game) *StatsPanel {
	sp := &StatsPanel{
		do:        make(chan func(*StatsPanel), 1),
		BasePanel: graphics.NewPanel(),
		g:         g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		sp.do <- func(sp *StatsPanel) {
			sp.w = ev.Width
			sp.h = ev.Height
			sp.Resize(ev.Width, ev.Height)
		}
	})

	return sp
}

func (sp StatsPanel) String() string {
	return fmt.Sprintf("TERM: %dx%d SZ %5.2f FPS %5.2f MB %d GC %d GR", sp.h, sp.w, sp.fps, float64(sp.memstats.HeapAlloc)/1000000.0, sp.memstats.NumGC, runtime.NumGoroutine())
}

func (sp *StatsPanel) Update(delta time.Duration) {
	sp.numgr = runtime.NumGoroutine()
	runtime.ReadMemStats(&sp.memstats)

	sp.samples[sp.current%FPS_SAMPLES] = 1.0 / delta.Seconds()
	sp.current++

	for i := 0; i < FPS_SAMPLES; i++ {
		sp.fps += sp.samples[i]
	}

	sp.fps /= FPS_SAMPLES

	for {
		select {
		case f := <-sp.do:
			f(sp)
		default:
			return
		}
	}
}

func (sp *StatsPanel) Draw() {
	if sp.Buffered != nil {

		str := fmt.Sprintf("%s", sp)
		for i, r := range str {
			sp.SetCell(i, 0, r, graphics.TextStyle.Fg, graphics.TextStyle.Bg)
		}

		sp.Buffered.Draw()
	}
}
