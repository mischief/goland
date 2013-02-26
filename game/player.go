// Player is a Unit that is controllable by a client
// (this should really have no distinction)
package game

type Player struct {
	*Unit
}

func NewPlayer() *Player {
	o := NewUnit()
	o.Glyph = GLYPH_HUMAN
	p := &Player{Unit: o}
	return p
}
