package main

import (
	"github.com/errnoh/termbox/panel"
	"github.com/nsf/termbox-go"
	"image"
	"time"
)

// ViewPanel holds the main viewport of the game,
// and needs a camera to apply the view transformation.
type ViewPanel struct {
	*panel.Buffered

	g *Game

	cam Camera
}

func NewViewPanel(g *Game) *ViewPanel {
	vp := &ViewPanel{g: g}

	vp.HandleInput(termbox.Event{Type: termbox.EventResize})

	return vp
}

func (vp *ViewPanel) Update(delta time.Duration) {
	// should set camera center
	if vp.g.Player != nil {
		vp.cam.SetCenter(image.Pt(vp.g.Player.GetPos()))
	} else {
		vp.cam.SetCenter(image.Pt(256/2, 256/2))
	}
}

func (vp *ViewPanel) HandleInput(ev termbox.Event) {
	if ev.Type == termbox.EventResize {
		w, h := termbox.Size()
		r := image.Rect(VIEW_START_X, VIEW_START_Y, w-VIEW_PAD_X, h-VIEW_PAD_Y)
		vp.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})

		vp.cam = NewCamera(r)

		//vp.cam.SetCenter(image.Pt(vp.g.Player.GetPos()))
	}
}

func (vp *ViewPanel) Draw() {

	vp.Clear()

	// draw terrain
	if vp.g.Map != nil {
		for x, row := range vp.g.Map.Locations {
			for y, terr := range row {
				realpos := vp.cam.Transform(image.Pt(x, y))
				if true || vp.cam.ContainsWorldPoint(realpos) {
					c := terr.Glyph
					vp.SetCell(realpos.X, realpos.Y, c.Ch, c.Fg, c.Bg)
				}
			}
		}
	}

	for o := range vp.g.Objects.Chan() {
		if o.GetTag("visible") {
			realpos := vp.cam.Transform(image.Pt(o.GetPos()))
			g := o.GetGlyph()
			vp.SetCell(realpos.X, realpos.Y, g.Ch, g.Fg, g.Bg)
		}
	}

	vp.Buffered.Draw()
}
