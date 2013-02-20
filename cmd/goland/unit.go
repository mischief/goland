package main

import (
	"time"
//  "log"
	"github.com/nsf/termbox-go"
)

const (
	PLAYER_SPEED         = 2
  PLAYER_RUN_SPEED     = 4
	DEFAULT_SPEED        = 8
	DEFAULT_ATTACK_SPEED = 8
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
	u := Unit{Ch:  MAP_EMPTY,
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

  t, ok := u.g.Map.GetTerrain(newpos)

  if ok && ! t.IsWall() {
    u.Pos = newpos
  }
}

func (u *Unit) Update(delta time.Duration) {
}

func (u *Unit) Draw(g *Game) {
	x, y := u.Pos.Round()
  g.PrintCell(x, y, u.Ch)
}

