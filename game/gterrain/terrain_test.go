package gterrain

import (
	"image"
	"testing"
)

type terraintest struct {
	terr                  Terrain
	blockmove, blocksight bool
}

var tests = []terraintest{
	terraintest{VoidTerrain, true, true},
	terraintest{FloorTerrain, false, false},
	terraintest{WallTerrain, true, true},
	terraintest{DoorTerrain, false, true},
}

func TestTerrain(t *testing.T) {
	for _, test := range tests {
		ter := test.terr.Data()

		if ter.BlocksMove() != test.blockmove {
			t.Errorf("blockmove expected %t got %t", test.blockmove, ter.BlocksMove())
		}

		if ter.BlocksSight() != test.blocksight {
			t.Fatalf("blocksight expected %t got %t", test.blocksight, ter.BlocksSight())
		}
	}
}

func TestTerrainLoad(t *testing.T) {
	ts := NewTerrainSystem()

	tname := "test"
	tfile := "testmap"
	err := ts.LoadFile(tname, tfile)

	if err != nil {
		t.Fatalf("error loading map %s from %s: %s", tname, tfile, err)
	}

	chunk, ok := ts.Get(tname)

	if !ok {
		t.Fatalf("map chunk %s not found in system", tname)
	}

	pts := map[image.Point]Terrain{
		image.Pt(0, 0): VoidTerrain,
		image.Pt(1, 0): FloorTerrain,
		image.Pt(2, 0): WallTerrain,
		image.Pt(3, 0): DoorTerrain,

		image.Pt(0, 255):   WallTerrain,
		image.Pt(255, 0):   WallTerrain,
		image.Pt(255, 255): WallTerrain,
		image.Pt(256, 256): VoidTerrain,
	}

	for p, ter := range pts {
		if chunk.At(p) != ter {
			t.Errorf("chunk %s expected terrain %v at %s got %v", tname, ter, p, chunk.At(p))
		}
	}

}
