package main

import (
	"image"
)

type Camera struct {
	Pos      image.Point     // center of the camera
	SizeRect image.Rectangle // camera's size
	Rect     image.Rectangle // camera's bounding box
}

func NewCamera(rect image.Rectangle) Camera {
	sz := rect.Size()
	r := image.Rect(0, 0, sz.X, sz.Y)

	c := Camera{
		Pos:      image.ZP,
		SizeRect: r,
		Rect:     r,
	}

	return c
}

// place the camera's center at pt
func (c *Camera) SetCenter(pt image.Point) {
	newpos := pt.Sub(c.SizeRect.Size().Div(2))
	c.Pos = pt
	c.Rect = image.Rect(newpos.X, newpos.Y, newpos.X+c.SizeRect.Dx(), newpos.Y+c.SizeRect.Dy())
}

//
func (c *Camera) Transform(pt image.Point) image.Point {
	return pt.Sub(c.Rect.Min) //.Add(c.Rect.Size().Div(2))
}

// check if world tile pt is inside camera bounds c.Rect
// FIXME
func (c *Camera) ContainsWorldPoint(pt image.Point) bool {
	return pt.In(c.Rect)
}
