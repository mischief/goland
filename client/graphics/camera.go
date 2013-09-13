package graphics

import (
	"fmt"
	"github.com/mischief/goland/game"
	"image"
)

// A camera property
type Camera struct {
	do chan func(*Camera)

	name string

	pos      image.Point     // center of the camera
	sizeRect image.Rectangle // camera's size
	rect     image.Rectangle // camera's bounding box
}

// Construct a new camera property with a name
func (sys *RenderSystem) Cam(name string) *Camera {
	c := &Camera{do: make(chan func(*Camera), 10)}

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
		c.rect = image.Rect(newp.X, newp.Y, newp.X+c.sizeRect.Dx(), newp.Y+c.sizeRect.Dy())
	}
}

func (c *Camera) Resize(size image.Point) {
	c.do <- func(c *Camera) {
		c.sizeRect = image.Rect(0, 0, size.X, size.Y)
	}
}

type Transformer func(pt image.Point) image.Point

// Returns a function which applies the viewport transformation of the camera
// for current offsets.
func (c *Camera) GetTransformer() Transformer {
	ch := make(chan Transformer)
	c.do <- func(c *Camera) {
		trans := func(pt image.Point) image.Point {
			return pt.Sub(c.rect.Min)
		}
		ch <- trans
		close(ch)
	}

	return <-ch
}

type Containsf func(pt image.Point) bool

// check if world tile pt is inside camera bounds c.Rect
func (c *Camera) GetContainsf() Containsf {
	ch := make(chan Containsf)
	c.do <- func(c *Camera) {
		containsf := func(pt image.Point) bool {
			return pt.In(c.rect)
		}
		ch <- containsf
		close(ch)
	}

	return <-ch
}

func (c *Camera) GetWorldIntersection(world image.Rectangle) image.Rectangle {
	ch := make(chan image.Rectangle)
	c.do <- func(c *Camera) {
		ch <- c.rect.Intersect(world)
		close(ch)
	}

	return <-ch
}
