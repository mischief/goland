package gterrain

import (
	"bufio"
	"encoding/gob"
	"github.com/golang/glog"
	"image"
	"os"
)

func init() {
	gob.Register(&TerrainChunk{})
}

const (
	MAP_WIDTH  = 256
	MAP_HEIGHT = 256
)

var (
	maprect = image.Rect(0, 0, MAP_WIDTH, MAP_HEIGHT)

	// translation of rune -> terrain type
	translation = map[rune]Terrain{
		'X': VoidTerrain,
		'.': FloorTerrain,
		'#': WallTerrain,
		'+': DoorTerrain,
	}
)

type TerrainChunk struct {
	Name string
	Locs [][]Terrain
}

func (tc TerrainChunk) String() string {
	return tc.Name
}

// Allocate a new TerrainChunk, setting all cells to the invalid cell
func NewTerrainChunk(n string) *TerrainChunk {
	tc := &TerrainChunk{Name: n}

	tc.Locs = make([][]Terrain, MAP_WIDTH)
	for col := range tc.Locs {
		tc.Locs[col] = make([]Terrain, MAP_HEIGHT)
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
			tc.Locs[x][y] = translation[rune(str[x])]
		}
	}

	return
}

// GetAt returns the tile at the given coordinate in the TerrainChunk.
func (tc *TerrainChunk) At(pt image.Point) Terrain {
	if pt.In(maprect) {
		return tc.Locs[pt.X][pt.Y]
	}

	return VoidTerrain
}

// Blocked returns true if the given coordinate in the TerrainChunk is not passable.
func (tc *TerrainChunk) Blocked(pt image.Point) bool {
	if pt.In(maprect) {
		return tc.Locs[pt.X][pt.Y].Data().BlocksMove()
	}

	return true
}

// Manager of different chunks of terrain
type TerrainSystem struct {
	terrains map[string]*TerrainChunk
}

func NewTerrainSystem() *TerrainSystem {
	sys := &TerrainSystem{
		terrains: make(map[string]*TerrainChunk),
	}

	return sys
}

// Load a terrain file into the manager
func (tm *TerrainSystem) LoadFile(mapname, filename string) error {
	if glog.V(2) {
		glog.Infof("loading terrainchunk %s from %s", mapname, filename)
	}

	tc := NewTerrainChunk(mapname)

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
