package main

import (
  "time"
  "github.com/nsf/termbox-go"
)

type Player struct {
  Unit // embed
}

func NewPlayer() (* Player) {
  o := NewUnit()
  o.Ch = termbox.Cell{'@', termbox.ColorGreen, termbox.ColorBlack}
  p := &Player{Unit: o}
  return p
}

func (p *Player) Update(delta time.Duration) {
}

