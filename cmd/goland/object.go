package main

type Direction int

const (
  DIR_UP Direction = iota
  DIR_DOWN
  DIR_LEFT
  DIR_RIGHT
)

type Vector struct {
  X, Y float64
}

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

type Object struct {
  Pos, Vel Vector
}

func (o *Object) Move(d Direction) {
  switch d {
  case DIR_UP:
    o.Pos.Y += 1
  case DIR_DOWN:
    o.Pos.Y -= 1
  case DIR_LEFT:
    o.Pos.X -= 1
  case DIR_RIGHT:
    o.Pos.X += 1
  }
}

