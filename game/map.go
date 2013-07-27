// Game map functions
package game

import (
	"bufio"
	"encoding/gob"
	"github.com/nsf/termbox-go"
	"image"
	"os"
	"github.com/golang/glog"
)

func init() {
	gob.Register(&Tile{})
	gob.Register(&TerrainChunk{})
}

type Tile struct {
	Name     string
  Passable bool
	Cell     termbox.Cell
}

func NewTile(name string, passable bool, cell termbox.Cell) *Tile {
	return &Tile{name, passable, cell}
}

func (t Tile) String() string {
	return t.Name
}

var tiles = []*Tile{
	&Tile{"invalid", false, termbox.Cell{Ch: 'X', Fg: termbox.ColorRed}},
	&Tile{"empty", true, termbox.Cell{Ch: ' '}},
	&Tile{"wall", false, termbox.Cell{Ch: '#', Fg: termbox.ColorDefault, Bg: termbox.ColorBlack | termbox.AttrUnderline | termbox.AttrBold}},
	&Tile{"grass", true, termbox.Cell{Ch: '.', Fg: termbox.ColorGreen}},
}

func TileByName(n string) *Tile {
	for _, t := range tiles {
		if t.Name == n {
			return t
		}
	}

	return tiles[0]
}

func TileByCh(r rune) *Tile {
	for _, t := range tiles {
		if t.Cell.Ch == r {
			return t
		}
	}

	return tiles[0]
}

const (
	MAP_WIDTH  = 256
	MAP_HEIGHT = 256
)

// Manager for terrain data
type TerrainSystem struct {
  scene *Scene
  terrains map[string] *TerrainChunk
}

func NewTerrainSystem(s *Scene) (*TerrainSystem, error) {
  sys := &TerrainSystem{
    terrains: make(map[string]*TerrainChunk),
  }

  return sys, nil
}

// Load a terrain file into the manager
func (tm *TerrainSystem) LoadFile(mapname, filename string) error {
  if glog.V(2) {
    glog.Infof("loading terrainchunk %s from %s", mapname, filename)
  }

  tc := NewTerrainChunk()

  if err := tc.Load(filename); err != nil {
    if glog.V(1) {
      glog.Infof("loading terrainchunk %s failed: %s", mapname, err)
    }
    return err
  } else {
    tm.terrains[mapname] = tc
  }

  if glog.V(2) {
    glog.Infof("loading terrainchunk %s successful", mapname)
  }

  return nil
}

// Get a terrain by name
func (tm *TerrainSystem) Get(mapname string) (t *TerrainChunk, ok bool) {
  t, ok = tm.terrains[mapname]
  return
}

// MAP_WIDTH*MAP_HEIGHT cells for terrain
type TerrainChunk struct {
  Cells [][]*Tile
}

// Allocate a new TerrainChunk, setting all cells to the invalid cell
func NewTerrainChunk() *TerrainChunk {
  tc := &TerrainChunk{}

  tc.Cells = make([][]*Tile, MAP_WIDTH)
  for col := range tc.Cells {
    tc.Cells[col] = make([]*Tile, MAP_HEIGHT)
  }

  for x := 0; x < MAP_WIDTH; x++ {
    for y := 0; y < MAP_HEIGHT; y++ {
      tc.Cells[x][y] = tiles[0]
    }
  }

  return tc
}

// Load this TerrainChunk's data from a file
func (tc *TerrainChunk) Load(file string) (err error) {
  var f *os.File
  if f, err = os.Open(file); err != nil {
    return
  } else {
    defer f.Close()
  }

  r := bufio.NewReader(f)

  for y := 0; y < MAP_HEIGHT; y++ {
    var str string
    str, err = r.ReadString('\n')
    if err != nil {
      return
    }

    for x := 0; x < MAP_WIDTH; x++ {
      tc.Cells[x][y] = TileByCh(rune(str[x]))
    }
  }

  return
}

// GetAt returns the tile at the given coordinate in the TerrainChunk.
func (tc *TerrainChunk) GetAt(pt image.Point) (t *Tile, ok bool) {
  if pt.X < 0 || pt.X > 256 || pt.Y < 0 || pt.Y > 256 {
    return tiles[0], false
  }

  return tc.Cells[pt.X][pt.Y], true
}

// Blocked returns true if the given coordinate in the TerrainChunk is not passable.
func (tc *TerrainChunk) Blocked(pt image.Point) bool {
  if pt.X < 0 || pt.X > 256 || pt.Y < 0 || pt.Y > 256 {
    return true
  }

  return !tc.Cells[pt.X][pt.Y].Passable
}

