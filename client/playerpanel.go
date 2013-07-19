package main

import (
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/nsf/termbox-go"
	"image"
	"time"
)

// PlayerPanel is a panel which runs the in-game chat and command input.
type PlayerPanel struct {
	*panel.Buffered // Panel

	x, y int
	name string

	g *Game
}

func NewPlayerPanel(g *Game) *PlayerPanel {
	cb := &PlayerPanel{g: g}

	cb.HandleInput(termbox.Event{Type: termbox.EventResize})

	return cb
}

func (c *PlayerPanel) Update(delta time.Duration) {
	/*
		p := c.g.GetPlayer()
		c.x, c.y = p.GetPos()
		c.name = p.GetName()
	*/
}

func (c *PlayerPanel) HandleInput(ev termbox.Event) {
	if ev.Type == termbox.EventResize {
		w, h := termbox.Size()
		r := image.Rect(1, h-2, w/2, h-1)
		c.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})
	}
}

func (c *PlayerPanel) Draw() {
	c.Clear()
	str := fmt.Sprintf("User: %s Pos: %d,%d", c.name, c.x, c.y)
	for i, r := range str {
		c.SetCell(i, 0, r, termbox.ColorBlue, termbox.ColorDefault)
	}
	c.Buffered.Draw()
}
