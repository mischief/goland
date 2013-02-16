package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

type Player struct {
	Unit // embed
}

func NewPlayer(g *Game) *Player {
	o := NewUnit(g)
	o.Ch = termbox.Cell{'@', termbox.ColorGreen, termbox.ColorBlack}
	p := &Player{Unit: o}
	return p
}

func (p *Player) Update(delta time.Duration) {
}
