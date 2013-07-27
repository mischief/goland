package main

import (
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/golang/glog"
	"github.com/mischief/goland/client/graphics"
	"github.com/mischief/goland/game/gnet"
	"github.com/nsf/termbox-go"
	"image"
	"time"
)

const (
	// Number of lines to store in the log's circular buffer
	nlines = 4
)

// LogPanel holds a log in a circular buffer,
// i.e. old entries fall off (not off the front, off the back)
type LogPanel struct {
	do chan func(*LogPanel)
	*panel.Buffered

	g *Game

	lines        int      // number of lines to show
	messages     []string // circular buffer of messages
	start, count int      // tracking for messages
}

// Construct a new LogPanel
func NewLogPanel(g *Game) *LogPanel {
	lp := &LogPanel{
		do:       make(chan func(*LogPanel), 1),
		g:        g,
		lines:    nlines,
		messages: make([]string, nlines),
	}

	g.em.On("log", func(i ...interface{}) {
		glog.Info("logpanel logging")
		lp.addline(fmt.Sprintf("error: %s", i[0]))
	})

	g.em.On("packet", func(i ...interface{}) {
		pkt := i[0].(*gnet.Packet)
		if pkt.Tag == "Rchat" {
			line, ok := pkt.Data[0].(string)
			if ok {
				lp.addline(line)
			}
		}
	})

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		lp.do <- func(lp *LogPanel) {
			lp.resize(ev.Width, ev.Height)
		}
	})

	return lp
}

// Draw log to internal panel
func (lp *LogPanel) Draw() {
	lp.Clear()

	fg := termbox.ColorBlue
	bg := termbox.ColorDefault

	y := 0

	if lp.start+lp.count > lp.lines {
		for _, line := range lp.messages[lp.start:] {
			for ic, r := range line {
				lp.SetCell(ic, y, r, fg, bg)
			}
			y++
		}
		for _, line := range lp.messages[:lp.start+lp.count-lp.lines] {
			for ic, r := range line {
				lp.SetCell(ic, y, r, fg, bg)
			}
			y++
		}
	} else {
		for _, line := range lp.messages[lp.start : lp.start+lp.count] {
			for ic, r := range line {
				lp.SetCell(ic, y, r, fg, bg)
			}
			y++
		}
	}

	lp.Buffered.Draw()
}

func (lp *LogPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-lp.do:
			f(lp)
		default:
			return
		}
	}
}

// resize event handler
func (lp *LogPanel) resize(w, h int) {
	r := image.Rect(1, h-7, w-1, h-3)
	lp.Buffered = panel.NewBuffered(r, graphics.BorderStyle)
	lp.SetTitle("log", graphics.TitleStyle)
}

// Write a line to the log
func (lp *LogPanel) addline(p string) {
	lp.do <- func(lp *LogPanel) {

		end := (lp.start + lp.count) % lp.lines

		lp.messages[end] = string(p)

		if lp.count == lp.lines {
			// we're at the end, just start overwriting
			lp.start = (lp.start + 1) % lp.lines
		} else {
			lp.count++
		}
	}

}
