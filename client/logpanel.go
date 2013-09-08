package main

import (
	"github.com/errnoh/termbox/panel"
	"github.com/nsf/termbox-go"
	"image"
	"sync"
)

const (
	// Number of lines to store in the log's circular buffer
	nlines = 4
)

// LogPanel holds a log in a circular buffer,
// i.e. old entries fall off (not off the front, off the back)
type LogPanel struct {
	*panel.Buffered
	sync.Mutex

	lines        int      // number of lines to show
	messages     []string // circular buffer of messages
	start, count int      // tracking for messages
}

// Construct a new LogPanel
func NewLogPanel() *LogPanel {
	lp := &LogPanel{
		lines:    nlines,
		messages: make([]string, nlines),
	}

	lp.HandleInput(termbox.Event{Type: termbox.EventResize})

	return lp
}

// Handle an input event
// Only used for resizing panel
func (lp *LogPanel) HandleInput(ev termbox.Event) {
	if ev.Type == termbox.EventResize {
		w, h := termbox.Size()
		r := image.Rect(1, h-7, w-1, h-3)
		lp.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})
	}
}

// Draw log to internal panel
func (lp *LogPanel) Draw() {
	lp.Lock()
	defer lp.Unlock()
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

// Write a line to the log
func (lp *LogPanel) Write(p []byte) (n int, err error) {
	lp.Lock()
	defer lp.Unlock()
	end := (lp.start + lp.count) % lp.lines

	lp.messages[end] = string(p)

	if lp.count == lp.lines {
		// we're at the end, just start overwriting
		lp.start = (lp.start + 1) % lp.lines
	} else {
		lp.count++
	}

	return len(p), nil
}
