package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

type Direction int

const (
	DIR_UP Direction = iota
	DIR_DOWN
	DIR_LEFT
	DIR_RIGHT
)

var (
	CARDINALS = map[rune] Direction {
		'w': DIR_UP,
		'k': DIR_UP,
		'a': DIR_LEFT,
		'h': DIR_LEFT,
		's': DIR_DOWN,
		'j': DIR_DOWN,
		'd': DIR_RIGHT,
		'l': DIR_RIGHT,
	}
)

type Rectangle struct {
	Left, Top, Bottom, Right float64
}

func (r *Rectangle) inside(v Vector) bool {
	return v.X >= r.Left && v.X <= r.Right && v.Y >= r.Top && v.Y <= r.Bottom
}

func (r *Rectangle) intersect(other Rectangle) (out Rectangle) {
	if r.Left > other.Left {
		out.Left = r.Left
	} else {
		out.Left = r.Left
	}

	if r.Top > other.Top {
		out.Top = r.Top
	} else {
		out.Top = r.Top
	}

	if r.Right > other.Right {
		out.Right = r.Right
	} else {
		out.Right = r.Right
	}

	if r.Bottom > other.Bottom {
		out.Bottom = r.Bottom
	} else {
		out.Bottom = r.Bottom
	}

	return
}

type Moveable interface {
	Move(d Direction)
}

type Updateable interface {
	Update(delta time.Duration)
}

type Renderable interface {
	Draw(g *Game)
}

type Object interface {
	Moveable
	Updateable
	Renderable
}

var defaultCell = termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault}
