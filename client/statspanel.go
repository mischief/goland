package main

import (
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/nsf/termbox-go"
	"image"
	"runtime"
	"time"
)

const (
	FPS_SAMPLES = 64
)

type StatsPanel struct {
	*panel.Buffered

	samples  [64]float64
	current  int
	memstats runtime.MemStats

	fps float64
}

func NewStatsPanel() *StatsPanel {
	sp := &StatsPanel{}

	sp.HandleInput(termbox.Event{Type: termbox.EventResize})

	return sp
}

func (s StatsPanel) String() string {
	return fmt.Sprintf("%5.2f FPS %5.2f MB %d GC %d GR", s.fps, float64(s.memstats.HeapAlloc)/1000000.0, s.memstats.NumGC, runtime.NumGoroutine())
}

func (s *StatsPanel) Update(delta time.Duration) {
	runtime.ReadMemStats(&s.memstats)

	s.samples[s.current%FPS_SAMPLES] = 1.0 / delta.Seconds()
	s.current++

	for i := 0; i < FPS_SAMPLES; i++ {
		s.fps += s.samples[i]
	}

	s.fps /= FPS_SAMPLES
}

func (s *StatsPanel) HandleInput(ev termbox.Event) {
	if ev.Type == termbox.EventResize {
		w, _ := termbox.Size()
		r := image.Rect(1, 1, w-1, 2)
		s.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})
	}
}

func (s *StatsPanel) Draw() {
	w, h := termbox.Size()
	str := fmt.Sprintf("Terminal: %d,%d SZ %s", w, h, s)
	for i, r := range str {
		s.SetCell(i, 0, r, termbox.ColorBlue, termbox.ColorDefault)
	}
	//io.WriteString(s, s.String() + fmt.Sprintf(" %s", s.Bounds()))
	s.Buffered.Draw()
}
