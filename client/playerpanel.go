package main

import (
	"fmt"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"time"
)

// PlayerPanel displays player status info
type PlayerPanel struct {
	do chan func(*PlayerPanel)
	*graphics.BasePanel

	x, y int
	name string

	g *Game
}

func NewPlayerPanel(g *Game) *PlayerPanel {
	pp := &PlayerPanel{
		do:        make(chan func(*PlayerPanel), 1),
		BasePanel: graphics.NewPanel(),
		g:         g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		pp.do <- func(pp *PlayerPanel) {
			pp.Resize(ev.Width, ev.Height)
		}
	})

	return pp
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
			pp.SetCell(i, 0, r, graphics.TextStyle.Fg, graphics.TextStyle.Bg)
		}
		pp.Buffered.Draw()
	}
}
