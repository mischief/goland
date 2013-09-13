package gid

import (
	"testing"
)

func TestSerialGen(t *testing.T) {
	var gen GidGen

	gen = new(SerialGen)

	id1, id2 := gen.Gen(), gen.Gen()

	if id2 <= id1 {
		t.Errorf("ids not incrementing, got %s after %s", id2, id1)
	}
}

func BenchmarkSerialGen(b *testing.B) {
	var gen GidGen

	gen = new(SerialGen)

	for i := 0; i < b.N; i++ {
		_ = gen.Gen()
	}
}
