package gfx

import (
	"fmt"
	termbox "github.com/nsf/termbox-go"
)

type Sprite interface {
	Cell() termbox.Cell
	Advance()
}

type StaticSprite struct {
	cell termbox.Cell
}

func NewStaticSprite(cell termbox.Cell) *StaticSprite {
	return &StaticSprite{cell}
}

func (s StaticSprite) String() string {
	return fmt.Sprintf("frame 0 of 0 current %c", s.cell.Ch)
}

func (s *StaticSprite) Cell() termbox.Cell {
	return s.cell
}

func (s *StaticSprite) Advance() {
}

type AnimatedSprite struct {
	current int
	frames  []termbox.Cell
}

func (a AnimatedSprite) String() string {
	c := a.frames[a.current]
	return fmt.Sprintf("frame %d of %d current %c of %v", a.current, len(a.frames)-1, c.Ch, a.frames)
}

func (a *AnimatedSprite) Cell() termbox.Cell {
	return a.frames[a.current]
}

func (a *AnimatedSprite) Advance() {
	a.current = (a.current + 1) % len(a.frames)
}

func Get(sp string) Sprite {
	s, ok := gfx[sp]
	if ok {
		return s
	}

	return gfx["void"]
}

var gfx = map[string]Sprite{
	"void":  &StaticSprite{termbox.Cell{Ch: 'X', Fg: termbox.ColorRed}},
	"floor": &StaticSprite{termbox.Cell{Ch: '.', Fg: termbox.ColorGreen}},
	"wall":  &StaticSprite{termbox.Cell{Ch: '#', Fg: termbox.ColorWhite, Bg: termbox.ColorBlack | termbox.AttrUnderline | termbox.AttrBold}},

	/*
		"door": &AnimatedSprite{
			frames: []termbox.Cell{
				termbox.Cell{Ch: '+', Fg: termbox.ColorYellow},
				termbox.Cell{Ch: '/', Fg: termbox.ColorYellow},
			},
		},
	*/

	"flag":  &StaticSprite{termbox.Cell{Ch: '⚑', Fg: termbox.ColorRed}},
	"human": &StaticSprite{termbox.Cell{Ch: '@', Fg: termbox.ColorWhite}},
	/*
		"human": &AnimatedSprite{
			frames: []termbox.Cell{
				termbox.Cell{Ch: '@', Fg: termbox.ColorDefault, Bg: termbox.ColorBlack},
				termbox.Cell{Ch: '@', Fg: termbox.ColorYellow, Bg: termbox.ColorWhite},
			},
		},
	*/
}
