package main

import (
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"image"
	"time"
)

// PlayerPanel displays player status info
type PlayerPanel struct {
	do              chan func(*PlayerPanel)
	*panel.Buffered // Panel

	x, y int
	name string

	g *Game
}

func NewPlayerPanel(g *Game) *PlayerPanel {
	pp := &PlayerPanel{
		do: make(chan func(*PlayerPanel), 1),
		g:  g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		pp.do <- func(pp *PlayerPanel) {
			pp.resize(ev.Width, ev.Height)
		}
	})

	return pp
}

func (pp *PlayerPanel) resize(w, h int) {
	r := image.Rect(1, h-2, w/2, h-1)
	pp.Buffered = panel.NewBuffered(r, graphics.BorderStyle)
	pp.SetTitle("player", graphics.TitleStyle)
}

func (pp *PlayerPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-pp.do:
			f(pp)
		default:
			return
		}
	}
	/*
		p := c.g.GetPlayer()
		c.x, c.y = p.GetPos()
		c.name = p.GetName()
	*/
}

func (pp *PlayerPanel) Draw() {
	if pp.Buffered != nil {
		pp.Clear()
		str := fmt.Sprintf("user: %s pos: %d,%d", pp.name, pp.x, pp.y)
		for i, r := range str {
			pp.SetCell(i, 0, r, termbox.ColorBlue, termbox.ColorDefault)
		}
		pp.Buffered.Draw()
	}
}
