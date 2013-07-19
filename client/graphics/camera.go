package graphics

import (
	"github.com/mischief/goland/game"
	"image"
  "fmt"
)

type Camera struct {
	do chan func(*Camera)

	pos      image.Point     // center of the camera
	sizeRect image.Rectangle // camera's size
	rect     image.Rectangle // camera's bounding box
}

func (sys *RenderSystem) Cam(name string) *Camera {
	c := &Camera{do: make(chan func(*Camera))}

	c.start()

	sys.do <- func(sys *RenderSystem) {
		// TODO: attach camera to viewport
	}

	sys.Syn()
	c.Syn()

	return c
}

func (c Camera) String() string {
	return fmt.Sprintf("pos %s size %s", c.pos, c.sizeRect)
}

func (c *Camera) Type() game.PropType {
	return game.PropCamera
}

func (c *Camera) Syn() {
	ack := make(chan bool)
	c.do <- func(c *Camera) {
		ack <- true
	}
	<-ack
	close(ack)
}

func (c *Camera) start() {
  go func() {
    for f := range c.do {
      f(c)
    }
  }()
}

func (c *Camera) SetCenter(p image.Point) {
  c.do <- func(c *Camera) {
    newp := p.Sub(c.sizeRect.Size().Div(2))
    c.pos = newp
    c.rect = image.Rect(newp.X, newp.Y, newp.X + c.sizeRect.Dx(), newp.Y + c.sizeRect.Dy())
  }
}

func (c *Camera) Resize(size image.Point) {
  c.do <- func(c *Camera) {
    c.sizeRect = image.Rect(0, 0, size.X, size.Y)
  }
}

func (c *Camera) Transform(pt image.Point) <-chan image.Point {
  ch := make(chan image.Point)
  c.do <- func(c *Camera) {
    newpt := pt.Sub(c.rect.Min)
    ch <- newpt
    close(ch)
  }

  return ch
}


/*
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
*/
