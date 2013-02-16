package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

type Player struct {
	Unit // embed
}

func NewPlayer() *Player {
	o := NewUnit()
	o.Ch = termbox.Cell{'@', termbox.ColorGreen, termbox.ColorBlack}
	p := &Player{Unit: o}
	return p
}

func (p *Player) Update(delta time.Duration) {
}
