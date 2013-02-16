package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

const (
	PLAYER_SPEED         = 2
	DEFAULT_SPEED        = 8
	DEFAULT_ATTACK_SPEED = 8
)

var (
  defaultCell = termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault}
)

type Unit struct {
	Pos, Vel Vector
	Ch       termbox.Cell // character for this object
	Hp       int

	Speed, AttackSpeed int
	wait               int

	g *Game // world this unit was created in
}

func NewUnit(g *Game) Unit {
	u := Unit{Ch: defaultCell,
		Speed:       DEFAULT_SPEED,
		AttackSpeed: DEFAULT_ATTACK_SPEED}
	u.g = g

	return u
}

func (u *Unit) Move(d Direction) {
	newpos := u.Pos
	switch d {
	case DIR_UP:
		newpos.Y -= 1
	case DIR_DOWN:
		newpos.Y += 1
	case DIR_LEFT:
		newpos.X -= 1
	case DIR_RIGHT:
		newpos.X += 1
	}

  // check for a valid position
  rect := Rectangle{0, 0, u.g.Map.Size.Y-1, u.g.Map.Size.X-1}
  if ! rect.Inside(newpos) {
    return
  }

	if u.g.Map.Tiles[int(newpos.X)][int(newpos.Y)].IsBlocked() {
		return
	}

	u.Pos = newpos
}

func (u *Unit) Update(delta time.Duration) {
}

func (u *Unit) Draw(g *Game) {
	x, y := u.Pos.Round()
	termbox.SetCell(x, y, u.Ch.Ch, u.Ch.Fg, u.Ch.Bg)
}
