// Core game object interfaces
package game

import (
	"bytes"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"image"
	"time"
)

type Positionable interface {
	SetPos(image.Point) bool // Set a new position. Returns false if that position is unavailable.
	GetPos() image.Point
}

type Updateable interface {
	Update(delta time.Duration)
}

type Renderable interface {
	Draw(*tulib.Buffer, image.Point)
}

/*type Object interface {
	Positionable
	Updateable
	Renderable
}*/

type GameObject struct {
	Pos   image.Point
	Glyph termbox.Cell // character for this object

	Tags map[string]bool
}

func NewGameObject() *GameObject {
	gob := &GameObject{
		Pos:   image.ZP,
		Glyph: termbox.Cell{'¡', termbox.ColorRed, termbox.ColorDefault},
	}

	gob.Tags = make(map[string]bool)

	return gob
}

func (gob GameObject) String() string {
	var buf bytes.Buffer

	for key, value := range gob.Tags {
		buf.WriteString(fmt.Sprintf(" %s:%t", key, value))
	}

	return fmt.Sprintf("%s (%c) tags:%s", gob.Pos, gob.Glyph.Ch, buf.String())
}

func (gob *GameObject) SetPos(pos image.Point) bool {
	gob.Pos = pos
	return true
}

func (gob *GameObject) GetPos() image.Point {
	return gob.Pos
}

func (gob *GameObject) Update(delta time.Duration) {
}

func (gob *GameObject) Draw(buf *tulib.Buffer, pos image.Point) {
	buf.Set(pos.X, pos.Y, gob.Glyph)
}
