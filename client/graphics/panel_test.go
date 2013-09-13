package graphics

import (
	"bytes"
	"image"
	"testing"
)

func TestWriteCenteredLine(t *testing.T) {
	bp := NewPanel()

	test := "hello"
	WriteCenteredLine(bp, test, 0, 0, 0)

	startx := (bp.Bounds().Dx() / 2) - (len(test) / 2)
	for i, c := range test {
		at := bp.At(i+startx, 0).Ch
		if at != c {
			t.Fatalf("%c != %c", at, c)
		}
	}

	var got bytes.Buffer
	buf := bp.Buffer()
	for i := startx; i < startx+len(test); i++ {
		got.WriteRune(buf[i].Ch)
	}

	t.Logf("expected %q got %q", test, got.String())
}

func TestActivator(t *testing.T) {
	var act Activator

	act = NewPanel()
	if act.IsActive() {
		t.Fatal("panel already active")
	}

	act.Activate()
	if !act.IsActive() {
		t.Fatal("panel didn't activate")
	}

	act.Deactivate()
	if act.IsActive() {
		t.Fatal("panel didn't deactivate")
	}
}

func TestResize(t *testing.T) {
	p := NewPanel()

	sw, sh := 80, 24

	neww, newh := p.Resize(sw, sh)

	if neww != 78 || newh != 22 {
		t.Fatal("resize with default function failed")
	}

	sw, sh = 100, 100

	p.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(w/2, h/2, w-1, h-1)
	})

	neww, newh = p.Resize(sw, sh)

	if neww != 49 || newh != 49 {
		t.Fatal("resize with custom function failed")
	}
}
