package graphics

import (
	"fmt"
	"github.com/mischief/goland/game"
	termbox "github.com/nsf/termbox-go"
)

type StaticSprite struct {
	do   chan func(*StaticSprite)
	name string
	cell termbox.Cell
}

func (sys *RenderSystem) StaticSprite(name string, c termbox.Cell) *StaticSprite {
	ss := &StaticSprite{
		do:   make(chan func(*StaticSprite)),
		name: name,
		cell: c,
	}

	//ss.start()

	return ss
}

func (ss StaticSprite) String() string {
	return fmt.Sprintf("sprite %s cell %v", ss.name, ss.cell)
}

func (ss *StaticSprite) Type() game.PropType {
	return game.PropStaticSprite
}

func (ss *StaticSprite) GetCell() termbox.Cell {
	return ss.cell
}
