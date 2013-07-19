package graphics

import (
	"github.com/errnoh/termbox/panel"
	"github.com/golang/glog"
	"github.com/mischief/goland/game"
	"github.com/nsf/termbox-go"
	"image"
	"time"
)

const (
	VIEW_START_X = 1
	VIEW_START_Y = 3
	VIEW_PAD_X   = 1
	VIEW_PAD_Y   = 8
)

// ViewPanel holds the main viewport of the game,
// and needs a camera to apply the view transformation.
type ViewPanel struct {
	*panel.Buffered

	rsys *RenderSystem

	cam *game.Actor
}

func NewViewPanel(rsys *RenderSystem, cam *game.Actor) *ViewPanel {
	vp := &ViewPanel{rsys: rsys,
		cam: cam}

	vp.HandleInput(termbox.Event{Type: termbox.EventResize})

	return vp
}

func (vp *ViewPanel) Update(delta time.Duration) {
	// should set camera center
	/*
		p := vp.g.GetPlayer()
		if p != nil {
			vp.cam.SetCenter(image.Pt(p.GetPos()))
		} else {
			vp.cam.SetCenter(image.Pt(256/2, 256/2))
		}
	*/
}

func (vp *ViewPanel) HandleInput(ev termbox.Event) {
	if ev.Type == termbox.EventResize {
		w, h := termbox.Size()
		r := image.Rect(VIEW_START_X, VIEW_START_Y, w-VIEW_PAD_X, h-VIEW_PAD_Y)
		vp.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})

		cam := vp.cam.Get(game.PropCamera).(*Camera)

		cam.Resize(r.Size())

		//vp.cam.SetCenter(image.Pt(vp.g.Player.GetPos()))
	}
}

func (vp *ViewPanel) Draw() {

	vp.Clear()

	cam := vp.cam.Get(game.PropCamera).(*Camera)

	for _, actor := range vp.rsys.scene.Find(game.PropPos, game.PropStaticSprite) {
		pos := actor.Get(game.PropPos).(*game.Pos)
		sp := actor.Get(game.PropStaticSprite).(*StaticSprite)
		ipt := <-pos.Get()
		screenpos := <-cam.Transform(ipt)
		c := sp.GetCell()

		if glog.V(3) {
			glog.Infof("%s (%v) at %s drawn at %s", actor, c, ipt, screenpos)
			glog.Infof("cam %s", cam)
		}

		vp.SetCell(screenpos.X, screenpos.Y, c.Ch, c.Fg, c.Bg)
	}

	// draw terrain
	/*
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
	*/

	vp.Buffered.Draw()
}
