package gobj

import (
	"github.com/mischief/goland/game/gfx"
	"github.com/mischief/goland/game/gid"
	"time"
)

type ObjectType int8

const (
	VoidObject ObjectType = iota
	PlayerType
	ItemType
)

type Object interface {
	GetType() ObjectType

	// Setter/getter for ID
	SetID(id gid.Gid)
	GetID() gid.Gid

	SetName(name string)
	GetName() string

	// Setter/getter for position
	SetPos(x, y int) bool
	GetPos() (x, y int)

	// Sprite
	SetSprite(gfx.Sprite)
	GetSprite() gfx.Sprite

	// Setter/getter for tags
	SetTag(tag string, val interface{}) interface{} // returns old value
	GetTag(tag string) interface{}

	// update this object with delta
	Update(delta time.Duration)
}
