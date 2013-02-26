package main

import (
	"github.com/nsf/tulib"
	"image"
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
	CARDINALS = map[rune]Direction{
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

type Moveable interface {
	Move(d Direction)
}

type Updateable interface {
	Update(delta time.Duration)
}

type Renderable interface {
	Draw(*tulib.Buffer, image.Point)
}

type Locateable interface {
	GetPos() image.Point
}

type Object interface {
	Moveable
	Updateable
	Renderable
	Locateable
}
