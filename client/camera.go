package main

import (
	goland "github.com/mischief/goland/game"
	"github.com/nsf/tulib"
	"image"
)

const (
	VIEW_START_X = 1
	VIEW_START_Y = 1
	VIEW_PAD_X   = 1
	VIEW_PAD_Y   = 1
)

type Camera struct {
	Render tulib.Buffer    // the memory buffer this camera draws to
	Pos    image.Point     // center of the camera
	Rect   image.Rectangle // camera's bounding box
}

func NewCamera(render tulib.Buffer) Camera {
	c := Camera{
		Render: render,
		Pos:    image.ZP,
		Rect:   image.Rect(0, 0, render.Rect.Width, render.Rect.Height),
	}

	return c
}

// place the camera's center at pt
func (c *Camera) SetCenter(pt image.Point) {
	newpos := pt.Sub(c.Rect.Size().Div(2))
	c.Pos = pt
	c.Rect = image.Rect(newpos.X, newpos.Y, newpos.X+c.Render.Rect.Width, newpos.Y+c.Render.Rect.Height)
}

// 
func (c *Camera) transform(pt image.Point) image.Point {
	return pt.Sub(c.Rect.Min) //.Add(c.Rect.Size().Div(2))
}

// draw r at pt, applying any camera transformations
func (c *Camera) Draw(r goland.Renderable, pt image.Point) {
	r.Draw(&c.Render, c.transform(pt))
}

// adjust camera coordinates by pt
func (c *Camera) Translate(pt image.Point) {
	c.Pos.Add(pt)
	c.SetCenter(c.Pos)
}

// check if world tile pt is inside camera bounds c.Rect
func (c *Camera) ContainsWorldPoint(pt image.Point) bool {
	return pt.In(c.Rect)
}
