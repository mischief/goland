package gid

import (
	"encoding/gob"
	"sync/atomic"
)

// ID used by game objects, players, instances
type Gid int32

// ID generator
type GidGen interface {
	Gen() Gid
}

// Serial ID generator.
// Generates IDs like 0, 1, 2, 3 ...
type SerialGen struct {
	now int32
}

func (sg *SerialGen) Gen() Gid {
	id := atomic.AddInt32(&sg.now, 1)

	return Gid(id)
}

// default id generator
var defgen GidGen

// Generate a unique id
func Gen() Gid {
	return defgen.Gen()
}

func init() {
	gob.Register(Gid(0))

	defgen = new(SerialGen)
}
