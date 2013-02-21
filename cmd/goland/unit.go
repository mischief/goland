package main

import (
	"fmt"
	"image"
	"time"
	//  "log"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
)

const (
	PLAYER_SPEED         = 2
	PLAYER_RUN_SPEED     = 4
	DEFAULT_SPEED        = 8
	DEFAULT_ATTACK_SPEED = 8
	DEFAULT_HP           = 10
	DEFAULT_AC           = 5
)

type Unit struct {
	Pos, Vel image.Point
	Ch       termbox.Cell // character for this object
	Hp, AC   int

	Speed, AttackSpeed int
	wait               int

	g *Game // world this unit was created in
}

func NewUnit(g *Game) Unit {
	u := Unit{Ch: MAP_EMPTY,
		Hp:          DEFAULT_HP,
		AC:          DEFAULT_AC,
		Speed:       DEFAULT_SPEED,
		AttackSpeed: DEFAULT_ATTACK_SPEED}

	u.g = g

	return u
}

func (u *Unit) String() string {
	return fmt.Sprintf("Pos: %s Hp: %d AC: %d Speed: %d", u.Pos, u.Hp, u.AC, u.Speed)
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

	if ok && !t.IsWall() {
		u.Pos = newpos
	}
}

func (u *Unit) Update(delta time.Duration) {
}

func (u *Unit) Draw(b *tulib.Buffer, pt image.Point) {
	b.Set(pt.X, pt.Y, u.Ch)
}

func (u *Unit) GetPos() image.Point {
	return u.Pos
}
