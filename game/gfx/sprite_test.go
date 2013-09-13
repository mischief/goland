package gfx

import (
	"testing"
)

func TestAnimatedSprite(t *testing.T) {
	sp := Get("door")

	t.Logf("using sprite %s", sp)

	frames := []rune{'+', '/', '+'}

	for i, c := range frames {
		t.Logf("frame %d check %c vs %c", i, c, sp.Cell().Ch)
		if sp.Cell().Ch != c {
			t.Errorf("frame %d expected sprite ch %c, got %c", i, c, sp.Cell().Ch)
		}
		sp.Advance()
	}

}
