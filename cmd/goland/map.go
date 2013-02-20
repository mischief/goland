package main

import "github.com/nsf/termbox-go"

type TerrainType uint32

const (
	MAP_WIDTH  = 256
	MAP_HEIGHT = 256

  TEmpty TerrainType = iota
  TWall                       // can't pass/see through wall
  TFloor                       // passable/visible
)

var (
  MAP_EMPTY = termbox.Cell{Ch: ' '}
	MAP_WALL  = termbox.Cell{Ch: '#'}
	MAP_FLOOR = termbox.Cell{Ch: '.', Fg: termbox.ColorWhite}

  glyphTable = map[rune] *Terrain {
    ' ': &Terrain{MAP_EMPTY, TEmpty, false, false, true},
    '#': &Terrain{MAP_WALL, TWall, true, false, true},
    '.': &Terrain{MAP_FLOOR, TFloor, false, false, true},
  }
)

type Terrain struct {
  Glyph   termbox.Cell
  Type    TerrainType

  Edge, Seen, Lit bool
}

func (t *Terrain) IsEmpty() bool {
  return t.Type == TEmpty
}

func (t *Terrain) IsWall() bool {
  return t.Type == TWall
}

func (t *Terrain) IsFloor() bool {
  return t.Type == TFloor
}

func GlyphToTerrain(g rune) (* Terrain) {
  return glyphTable[g]
}

type MapChunk struct {
	Size        Vector
  Rect        Rectangle
	Locations   [][]*Terrain // land features
  Objects     []*Object // items
  Npcs        []*Unit // active monsters
  Players     []*Player // active players
}

func NewMapChunk() *MapChunk {
	ch := MapChunk{Size: Vector{MAP_WIDTH, MAP_HEIGHT}}
  ch.Rect = Rectangle{0, 0, MAP_WIDTH-1, MAP_HEIGHT-1}

	ch.Locations = make([][]*Terrain, MAP_WIDTH)
	for row := range ch.Locations {
		ch.Locations[row] = make([]*Terrain, MAP_HEIGHT)
	}

	for x := 0; x < MAP_WIDTH; x++ {
		for y := 0; y < MAP_HEIGHT; y++ {
			ch.Locations[x][y] = GlyphToTerrain('.')
		}
	}

	return &ch
}

// return true if the map chunk has a cell with coordinates v.X, v.Y
func (mc *MapChunk) HasVector(v Vector) bool {
  return mc.Rect.Inside(v)
}

// get terrain at v. returns nil, false if it is not present
func (mc *MapChunk) GetTerrain(v Vector) (t *Terrain, ok bool) {
  if ok = mc.HasVector(v); ! ok {
    return
  }
  x, y := v.Round()
  return mc.Locations[x][y], true
}
