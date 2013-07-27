package game

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
)

// StaticSprite is an Actor property that contains
// a sprite with a single frame.
type StaticSprite struct {
	do   chan func(*StaticSprite)
	name string
	cell termbox.Cell
}

// Make a new sprite
func NewStaticSprite(name string, c termbox.Cell) *StaticSprite {
	ss := &StaticSprite{
		do:   make(chan func(*StaticSprite)),
		name: name,
		cell: c,
	}

	return ss
}

func (ss StaticSprite) String() string {
	return fmt.Sprintf("sprite %s cell %v", ss.name, ss.cell)
}

func (ss *StaticSprite) Type() PropType {
	return PropStaticSprite
}

func (ss *StaticSprite) GetCell() termbox.Cell {
	return ss.cell
}
