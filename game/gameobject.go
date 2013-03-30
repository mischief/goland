// Core game object interfaces
package game

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	//	uuid "github.com/nu7hatch/gouuid"
	"image"
	//	"log"
	"time"
)

type GID int

var (
	idchan chan GID
)

func init() {
	gob.Register(&GameObject{})
	gob.Register(GID(0))

	idchan = make(chan GID)

	go func() {
		var id GID
		for {
			id++
			idchan <- id
		}
	}()
}

// TODO: remove need for this when drawing terrain with camera.Draw
type Renderable interface {
	// Draw this object on the buffer at pos
	Draw(buf *tulib.Buffer, pos image.Point)
}

type Object interface {
	// Setter/getter for ID
	SetID(id GID)
	GetID() GID

	SetName(name string)
	GetName() string

	// Setter/getter for position
	SetPos(x, y int) bool
	GetPos() (x, y int)

	// Setter/getter for the glyph
	SetGlyph(termbox.Cell)
	GetGlyph() termbox.Cell

	// Setter/getter for tags
	SetTag(tag string, val bool) bool // returns old value
	GetTag(tag string) bool

	GetSubObjects() GameObjectMap
	AddSubObject(obj Object)
	RemoveSubObject(obj Object) Object

	// update this object with delta
	Update(delta time.Duration)

	Renderable
}

type GameObject struct {
	ID         GID             // game object id
	Name       string          // object name
	Pos        image.Point     // object world coordinates
	Glyph      termbox.Cell    // character for this object
	Tags       map[string]bool // object tags
	SubObjects GameObjectMap   // objects associated with this one
}

func NewGameObject(name string) Object {
	gob := &GameObject{
		ID:         <-idchan,
		Name:       name,
		Pos:        image.ZP,
		Glyph:      termbox.Cell{'¡', termbox.ColorRed, termbox.ColorDefault},
		Tags:       make(map[string]bool),
		SubObjects: NewGameObjectMap(),
	}

	return gob
}

func (gob GameObject) String() string {
	var buf bytes.Buffer

	for key, value := range gob.Tags {
		buf.WriteString(fmt.Sprintf(" %s:%t", key, value))
	}

	return fmt.Sprintf("%s (%c) %s %d tags:%s", gob.Name, gob.Glyph.Ch,
		gob.Pos, gob.ID, buf.String())
}

func (gob *GameObject) SetID(id GID) {
	gob.ID = id
}

func (gob *GameObject) GetID() GID {
	return gob.ID
}

func (gob *GameObject) SetName(name string) {
	gob.Name = name
}

func (gob *GameObject) GetName() string {
	return gob.Name
}

func (gob *GameObject) SetPos(x, y int) bool {
	gob.Pos.X = x
	gob.Pos.Y = y
	return true
}

func (gob *GameObject) GetPos() (x, y int) {
	return gob.Pos.X, gob.Pos.Y
}

func (gob *GameObject) SetGlyph(glyph termbox.Cell) {
	gob.Glyph = glyph
}

func (gob *GameObject) GetGlyph() termbox.Cell {
	return gob.Glyph
}

func (gob *GameObject) SetTag(tag string, val bool) (old bool) {
	old = gob.Tags[tag]
	gob.Tags[tag] = val
	return
}

func (gob *GameObject) GetTag(tag string) bool {
	return gob.Tags[tag]
}

func (gob *GameObject) GetSubObjects() GameObjectMap {
	return gob.SubObjects
}

func (gob *GameObject) AddSubObject(obj Object) {
	gob.SubObjects.Add(obj)
}

func (gob *GameObject) RemoveSubObject(obj Object) Object {
	gob.SubObjects.RemoveObject(obj)
	return obj
}

func (gob *GameObject) Update(delta time.Duration) {
}

func (gob *GameObject) Draw(buf *tulib.Buffer, pos image.Point) {
	buf.Set(pos.X, pos.Y, gob.Glyph)
}

// handy interface for a collection of game objects
type GameObjectMap map[GID]Object

func NewGameObjectMap() GameObjectMap {
	g := make(GameObjectMap)
	return g
}

func (gom GameObjectMap) Add(obj Object) {
	// make sure we don't double insert
	if _, ok := gom[obj.GetID()]; !ok {
		gom[obj.GetID()] = obj
	}
}

func (gom GameObjectMap) RemoveObject(obj Object) {
	delete(gom, obj.GetID())
}

func (gom GameObjectMap) FindObjectByID(id GID) Object {
	o, ok := gom[id]

	if !ok {
		return nil
	}

	return o
}

func SamePos(ob1, ob2 Object) bool {
	x1, y1 := ob1.GetPos()
	x2, y2 := ob2.GetPos()
	return x1 == x2 && y1 == y2
}
