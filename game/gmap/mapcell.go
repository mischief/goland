package gmap

import (
	"github.com/mischief/goland/game/gid"
	"github.com/mischief/goland/game/gobj"
	"github.com/mischief/goland/game/gterrain"
	"sync"
)

//
type MapCell struct {
	Name    string
	objects map[gid.Gid]gobj.Object
	objlock sync.Mutex

	terrain *TerrainChunk
}
