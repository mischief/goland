package main

import (
	"bufio"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"image"
	"log"
	"os"
)

type TerrainType uint32

const (
	MAP_WIDTH  = 256
	MAP_HEIGHT = 256

	TEmpty TerrainType = iota
	TWall              // can't pass/see through wall
	TFloor             // passable/visible
)

var (
	MAP_EMPTY = termbox.Cell{Ch: ' '}
	MAP_WALL  = termbox.Cell{Ch: '#'}
	MAP_FLOOR = termbox.Cell{Ch: '.', Fg: termbox.ColorWhite}

	glyphTable = map[rune]*Terrain{
		' ': &Terrain{MAP_EMPTY, TEmpty, false, false, true},
		'#': &Terrain{MAP_WALL, TWall, true, false, true},
		'.': &Terrain{MAP_FLOOR, TFloor, false, false, true},
	}
)

func GlyphToTerrain(g rune) (t *Terrain, ok bool) {
	t, ok = glyphTable[g]
	if !ok {
		t = glyphTable[' ']
	}
	return
}

type Terrain struct {
	Glyph termbox.Cell
	Type  TerrainType

	Edge, Seen, Lit bool
}

func (t *Terrain) String() string {
	return fmt.Sprintf("(%c %d %t %t %t)", t.Glyph.Ch, t.Type, t.Edge, t.Seen, t.Lit)
}

func (t *Terrain) Draw(b *tulib.Buffer, pt image.Point) {
	b.Set(pt.X, pt.Y, t.Glyph)
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

type MapChunk struct {
	Size      image.Point
	Rect      image.Rectangle
	Locations [][]*Terrain // land features
	Objects   []*Object    // items
	Npcs      []*Unit      // active monsters
	Players   []*Player    // active players
}

func NewMapChunk() *MapChunk {
	ch := MapChunk{Size: image.Pt(MAP_WIDTH, MAP_HEIGHT)}
	ch.Rect = image.Rect(0, 0, MAP_WIDTH, MAP_HEIGHT)

	ch.Locations = make([][]*Terrain, MAP_WIDTH)
	for row := range ch.Locations {
		ch.Locations[row] = make([]*Terrain, MAP_HEIGHT)
	}

	for x := 0; x < MAP_WIDTH; x++ {
		for y := 0; y < MAP_HEIGHT; y++ {
			g, _ := GlyphToTerrain('.')
			ch.Locations[x][y] = g
		}
	}

	return &ch
}

// return true if the map chunk has a cell with coordinates v.X, v.Y
func (mc *MapChunk) HasCell(pt image.Point) bool {
	return pt.In(mc.Rect)
}

// get terrain at v. returns nil, false if it is not present
func (mc *MapChunk) GetTerrain(pt image.Point) (t *Terrain, ok bool) {
	if ok = mc.HasCell(pt); !ok {
		return
	}
	return mc.Locations[pt.X][pt.Y], true
}

func MapChunkFromFile(mapfile string) *MapChunk {
	mfh, err := os.Open(mapfile)
	if err != nil {
		log.Printf("can't open map file %s: %s", mapfile, err)
		return nil
	}

	defer mfh.Close()

	r := bufio.NewReader(mfh)

	mc := NewMapChunk()

	for y := 0; y < MAP_HEIGHT; y++ {
		str, err := r.ReadString('\n')
		if err != nil {
			log.Printf("map read error: %s", err)
			return nil
		}

		for x := 0; x < MAP_WIDTH; x++ {
			g, ok := GlyphToTerrain(rune(str[x]))
			if !ok {
				log.Printf("invalid map tile '%c' at %s:%d:%d", str[x], mapfile, y, x)
				return nil
			}

			mc.Locations[x][y] = g
		}
	}

	log.Printf("loaded map file %s ", mapfile)

	return mc
}
