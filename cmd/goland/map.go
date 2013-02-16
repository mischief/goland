package main

import "github.com/nsf/termbox-go"

const (
	MAP_WIDTH  = 80
	MAP_HEIGHT = 45
)

var (
	MAP_WALL  = termbox.Cell{Ch: '#'}
	MAP_FLOOR = termbox.Cell{Ch: '.', Fg: termbox.ColorWhite}
)

type Tile struct {
	Ch           termbox.Cell
	Blocked      bool // true if this tile blocks movement
	SightBlocked bool // true if this tile blocks sight
}

func (t *Tile) CanSee() bool {
	return !t.SightBlocked
}

func (t *Tile) IsBlocked() bool {
	return t.Blocked
}

type MapChunk struct {
	Size  Vector
	Tiles [][]Tile
}

func NewMapChunk() *MapChunk {
	ch := MapChunk{Size: Vector{MAP_WIDTH, MAP_HEIGHT}}

	ch.Tiles = make([][]Tile, MAP_WIDTH)
	for row := range ch.Tiles {
		ch.Tiles[row] = make([]Tile, MAP_HEIGHT)
	}

	for x := 0; x < MAP_WIDTH; x++ {
		for y := 0; y < MAP_HEIGHT; y++ {
			ch.Tiles[x][y] = Tile{MAP_FLOOR, false, false}
		}
	}

	return &ch
}
