// Core game object interfaces
package gobj

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/mischief/goland/game/gfx"
	"github.com/mischief/goland/game/gid"
	"image"
	"sync"
	"time"
)

func init() {
	gob.Register(&GameObject{})
	gob.Register(&GameObjectMap{})
}

// Data sent over the wire to the client
type GameObjectData struct {
	ID       gid.Gid     // ID
	Backpack gid.Gid     // ID of the backpack object
	Pos      image.Point // Position
	Sprite   gfx.Sprite  // Sprite
	Type     ObjectType  // Type
	Visible  bool        // Visible?
	Name     string      // Name
}

type GameObject struct {
	data GameObjectData
	tags map[string]interface{} // object tags

	// script?

	m sync.Mutex // lock, ew
}

func NewGameObject(id gid.Gid, name string) *GameObject {
	gob := &GameObject{
		data: GameObjectData{
			ID:     id,
			Name:   name,
			Sprite: gfx.Get("void"),
		},
		tags: make(map[string]interface{}),
	}

	return gob
}

func (gob GameObject) String() string {
	gob.m.Lock()
	defer gob.m.Unlock()

	var buf bytes.Buffer

	for key, value := range gob.tags {
		buf.WriteString(fmt.Sprintf(" %s:%v", key, value))
	}

	return fmt.Sprintf("%s (%c) %s %d tags:%s", gob.data.Name, gob.data.Sprite.Cell().Ch,
		gob.data.Pos, gob.data.ID, buf.String())
}

func (gob *GameObject) GetType() ObjectType {
	return VoidObject
}

func (gob *GameObject) SetID(id gid.Gid) {
	gob.m.Lock()
	defer gob.m.Unlock()

	gob.data.ID = id
}

func (gob *GameObject) GetID() gid.Gid {
	gob.m.Lock()
	defer gob.m.Unlock()

	return gob.data.ID
}

func (gob *GameObject) SetName(name string) {
	gob.m.Lock()
	defer gob.m.Unlock()

	gob.data.Name = name
}

func (gob *GameObject) GetName() string {
	gob.m.Lock()
	defer gob.m.Unlock()

	return gob.data.Name
}

func (gob *GameObject) SetPos(x, y int) bool {
	gob.m.Lock()
	defer gob.m.Unlock()

	gob.data.Pos.X = x
	gob.data.Pos.Y = y
	return true
}

func (gob *GameObject) GetPos() (x, y int) {
	gob.m.Lock()
	defer gob.m.Unlock()

	return gob.data.Pos.X, gob.data.Pos.Y
}

func (gob *GameObject) SetSprite(s gfx.Sprite) {
	gob.m.Lock()
	defer gob.m.Unlock()

	gob.data.Sprite = s
}

func (gob *GameObject) GetSprite() gfx.Sprite {
	gob.m.Lock()
	defer gob.m.Unlock()

	return gob.data.Sprite
}

func (gob *GameObject) SetTag(tag string, val interface{}) (old interface{}) {
	gob.m.Lock()
	defer gob.m.Unlock()

	old = gob.tags[tag]
	gob.tags[tag] = val
	return
}

func (gob *GameObject) GetTag(tag string) interface{} {
	gob.m.Lock()
	defer gob.m.Unlock()

	return gob.tags[tag]
}

func (gob *GameObject) Update(delta time.Duration) {
}

// handy interface for a collection of game objects
type GameObjectMap struct {
	Objs map[gid.Gid]Object

	m sync.Mutex
}

func NewGameObjectMap() *GameObjectMap {
	g := GameObjectMap{Objs: make(map[gid.Gid]Object)}
	return &g
}

func (gom *GameObjectMap) Add(obj Object) {
	// make sure we don't double insert
	gom.m.Lock()
	if _, ok := gom.Objs[obj.GetID()]; !ok {
		gom.Objs[obj.GetID()] = obj
	}
	gom.m.Unlock()
}

func (gom *GameObjectMap) RemoveObject(obj Object) {
	gom.m.Lock()
	delete(gom.Objs, obj.GetID())
	gom.m.Unlock()
}

func (gom *GameObjectMap) FindObjectByID(id gid.Gid) Object {
	gom.m.Lock()
	defer gom.m.Unlock()

	o, ok := gom.Objs[id]

	if !ok {
		return nil
	}

	return o
}

func (gom *GameObjectMap) Chan() <-chan Object {
	gom.m.Lock()
	defer gom.m.Unlock()

	ch := make(chan Object, len(gom.Objs))

	for _, o := range gom.Objs {
		ch <- o
	}

	close(ch)

	return ch
}

// return a slice containing the objects
// XXX: crappy hack so lua can iterate the contents
func (gom *GameObjectMap) GetSlice() []Object {
	gom.m.Lock()
	defer gom.m.Unlock()

	r := make([]Object, 1)
	for _, o := range gom.Objs {
		r = append(r, o)
	}
	return r
}

func SamePos(ob1, ob2 Object) bool {
	if ob1 == ob2 {
		return true
	}

	x1, y1 := ob1.GetPos()
	x2, y2 := ob2.GetPos()
	return x1 == x2 && y1 == y2
}
