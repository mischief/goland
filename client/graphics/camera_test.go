package graphics

import (
	"image"
	"testing"
)

func TestCamera(t *testing.T) {
	cam := &Camera{do: make(chan func(*Camera), 10)}

	cam.start()

	cam.Syn()

	size := image.Pt(10, 10)
	center := image.Pt(50, 50)

	cam.Resize(size)
	cam.SetCenter(center)

	tr := cam.GetTransformer()

	worldpos := image.Pt(45, 45)
	newpt := tr(worldpos)

	if newpt.Eq(image.ZP) != true {
		t.Logf("camera translation incorrect - %s not drawn at %s!", worldpos, image.ZP)
	}

	containsf := cam.GetContainsf()

	in := image.Pt(46, 46)
	if !containsf(in) {
		t.Errorf("camera %s doesn't contain point %s", cam, in)
	}

	out := image.Pt(44, 44)
	if containsf(out) {
		t.Errorf("camera %s should not contain point %s!", cam, out)
	}

	t.Logf("%s, %s", cam, newpt)
}
