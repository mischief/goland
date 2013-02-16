package main

import (
  "time"
  "github.com/nsf/termbox-go"
)

type Direction int

const (
  DIR_UP Direction = iota
  DIR_DOWN
  DIR_LEFT
  DIR_RIGHT
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

type Unit struct {
  Pos, Vel  Vector
  Ch        termbox.Cell // character for this object
}

var defaultCell = termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault}

func NewUnit() Unit {
  u := Unit{Ch: defaultCell}

  return u
}

func (u *Unit) Move(d Direction) {
  switch d {
  case DIR_UP:
    u.Pos.Y -= 1
  case DIR_DOWN:
    u.Pos.Y += 1
  case DIR_LEFT:
    u.Pos.X -= 1
  case DIR_RIGHT:
    u.Pos.X += 1
  }
}

func (u *Unit) Update(delta time.Duration) {
}

func (u *Unit) Draw(g *Game) {
  x, y := u.Pos.Round()
  termbox.SetCell(x, y, u.Ch.Ch, u.Ch.Fg, u.Ch.Bg)
}

