// Core game object interfaces
package game

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	uuid "github.com/nu7hatch/gouuid"
	"image"
	"log"
	"time"
)

func init() {
	gob.Register(&GameObject{})
	gob.Register(&uuid.UUID{})
}

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

type Object interface {
	Positionable
	Updateable
	Renderable
}

type GameObject struct {
	ID *uuid.UUID

	Pos   image.Point
	Glyph termbox.Cell // character for this object

	Tags map[string]bool
}

func NewGameObject() *GameObject {
	var err error

	gob := &GameObject{
		Pos:   image.ZP,
		Glyph: termbox.Cell{'¡', termbox.ColorRed, termbox.ColorDefault},
	}

	if gob.ID, err = uuid.NewV4(); err != nil {
		log.Printf("NewGameObject: %s", err)
		return nil
	}

	gob.Tags = make(map[string]bool)

	return gob
}

func (gob GameObject) String() string {
	var buf bytes.Buffer

	for key, value := range gob.Tags {
		buf.WriteString(fmt.Sprintf(" %s:%t", key, value))
	}

	return fmt.Sprintf("%s %s (%c) tags:%s", gob.ID, gob.Pos, gob.Glyph.Ch, buf.String())
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

// handy interface for a collection of game objects
type GameObjectMap map[uuid.UUID]*GameObject

func NewGameObjectMap() GameObjectMap {
	g := make(GameObjectMap)

	return g
}

func (gom GameObjectMap) Add(obj *GameObject) {
	// make sure we don't double insert
	if _, ok := gom[*obj.ID]; !ok {
		gom[*obj.ID] = obj
	}
}

func (gom GameObjectMap) RemoveObject(obj *GameObject) {
	delete(gom, *obj.ID)
}

func (gom GameObjectMap) FindObjectByID(id *uuid.UUID) *GameObject {
	o, ok := gom[*id]

	if !ok {
		return nil
	}

	return o
}
