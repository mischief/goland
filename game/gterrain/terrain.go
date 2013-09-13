package gterrain

import (
	"encoding/gob"
	"github.com/mischief/goland/game/gfx"
)

func init() {
	gob.Register(VoidTerrain)
	gob.Register(SolidKind)
	gob.Register(&TerrainData{})
}

// Terrain is the ID of a specific terrain
type Terrain int8

const (
	VoidTerrain Terrain = iota
	FloorTerrain
	WallTerrain
	DoorTerrain
)

// Get the data of a particular Terrain
func (t Terrain) Data() *TerrainData {
	return &terrainTable[t]
}

// TerrainKind is what properties the terrain has,
// open, blocked, wall, etc
type TerrainKind int8

const (
	SolidKind TerrainKind = iota
	OpenKind
	WallKind
	DoorKind
)

// TerrainData holds the TerrainKind information about a Terrain
type TerrainData struct {
	Name string
	Kind TerrainKind
	Gfx  gfx.Sprite
}

func (t TerrainData) BlocksSight() bool {
	switch t.Kind {
	case SolidKind, WallKind, DoorKind:
		return true
	}

	return false
}

func (t TerrainData) BlocksMove() bool {
	switch t.Kind {
	case SolidKind, WallKind:
		return true
	}
	return false
}

func TerrainDataByName(n string) TerrainData {
	for _, t := range terrainTable {
		if t.Name == n {
			return t
		}
	}

	return terrainTable[0]
}

var terrainTable = []TerrainData{
	TerrainData{"void", SolidKind, gfx.Get("void")},
	TerrainData{"floor", OpenKind, gfx.Get("floor")},
	TerrainData{"wall", WallKind, gfx.Get("wall")},
	TerrainData{"door", DoorKind, gfx.Get("door")},
}
