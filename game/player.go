// Player is a Unit that is controllable by a client
// (this should really have no distinction)
package game

import "encoding/gob"

func init() {
	gob.Register(&Player{})
}

type Player struct {
	*Unit
}

func NewPlayer(name string) *Player {
	o := NewUnit(name)
	o.SetGlyph(GLYPH_HUMAN)
	p := &Player{Unit: o}

	p.SetTag("player", true)
	return p
}
