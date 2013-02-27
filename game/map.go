// Game map functions
package game

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
type Direction int

const (
	MAP_WIDTH  = 256
	MAP_HEIGHT = 256

	T_EMPTY  TerrainType = iota
	T_WALL               // can't pass/see through wall
	T_GROUND             // passable/visible
	T_UNIT

	DIR_UP Direction = iota // player movement instructions
	DIR_DOWN
	DIR_LEFT
	DIR_RIGHT
)

var (
	DirTable = map[Direction]image.Point{
		DIR_UP:    image.Point{0, -1},
		DIR_DOWN:  image.Point{0, 1},
		DIR_LEFT:  image.Point{-1, 0},
		DIR_RIGHT: image.Point{1, 0},
	}

	GLYPH_EMPTY  = termbox.Cell{Ch: ' '}
	GLYPH_WALL   = termbox.Cell{Ch: '#', Fg: termbox.ColorBlack, Bg: termbox.ColorWhite}
	GLYPH_GROUND = termbox.Cell{Ch: '.', Fg: termbox.ColorGreen}
	GLYPH_HUMAN  = termbox.Cell{Ch: '@'}

	// convert a rune to a terrain square
	glyphTable = map[rune]*Terrain{
		' ': &Terrain{GLYPH_EMPTY, T_EMPTY},
		'#': &Terrain{GLYPH_WALL, T_WALL},
		'.': &Terrain{GLYPH_GROUND, T_GROUND},
		'@': &Terrain{GLYPH_HUMAN, T_UNIT},
	}
)

func (tt *TerrainType) String() string {
	switch *tt {
	case T_EMPTY:
		return "empty"
	case T_WALL:
		return "wall"
	case T_GROUND:
		return "ground"
	case T_UNIT:
		return "unit"
	}

	return "unknown"
}

func GlyphToTerrain(g rune) (t *Terrain, ok bool) {
	t, ok = glyphTable[g]
	if !ok {
		t = glyphTable[' ']
	}
	return
}

type Terrain struct {
	//*GameObject
	Glyph termbox.Cell
	Type  TerrainType
}

func (t Terrain) String() string {
	return fmt.Sprintf("(%c %s)", t.Glyph.Ch, t.Type)
}

func (t *Terrain) Draw(b *tulib.Buffer, pt image.Point) {
	b.Set(pt.X, pt.Y, t.Glyph)
}

func (t *Terrain) IsEmpty() bool {
	return t.Type == T_EMPTY
}

func (t *Terrain) IsWall() bool {
	return t.Type == T_WALL
}

func (t *Terrain) IsGround() bool {
	return t.Type == T_GROUND
}

type MapChunk struct {
	Size        image.Point
	Rect        image.Rectangle
	Locations   [][]*Terrain  // land features
	GameObjects []*GameObject // active game objects
	Players     []*Player     // active players
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

func (mc *MapChunk) CheckCollision(gob *GameObject, pos image.Point) bool {
	t, ok := mc.GetTerrain(pos)
	if ok {
		return !t.IsWall()
	}

	return false
}

func MapChunkFromFile(mapfile string) *MapChunk {
	mfh, err := os.Open(mapfile)
	if err != nil {
		log.Printf("Error loading map chunk file '%s': %s", mapfile, err)
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

	return mc
}
