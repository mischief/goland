package main

import (
	"github.com/errnoh/termbox/panel"
	"github.com/golang/glog"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/client/graphics"
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
	do chan func(*ViewPanel)
	*graphics.BasePanel

	width int

  g *Game
	rsys *graphics.RenderSystem

	cam *game.Actor
}

func NewViewPanel(g *Game, cam *game.Actor) *ViewPanel {
	vp := &ViewPanel{
		do:   make(chan func(*ViewPanel), 10),
    BasePanel: graphics.NewBasePanel(g.rsys),
    g: g,
		rsys: g.rsys,
		cam:  cam,
	}

	vp.g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		vp.do <- func(vp *ViewPanel) {
      rw, rh := vp.Resize(ev.Width, ev.Height)
      vp.width = ev.Width
	    cam := vp.cam.Get(game.PropCamera).(*graphics.Camera)
	    cam.Resize(image.Pt(rw, rh))
		}
	})
	return vp
}

func (vp *ViewPanel) resize(w, h int) {
	r := image.Rect(VIEW_START_X, VIEW_START_Y, w-VIEW_PAD_X, h-VIEW_PAD_Y)
	//vp.width = r.Dx()
	vp.Buffered = panel.NewBuffered(r, graphics.BorderStyle)
	vp.SetTitle("view", graphics.TitleStyle)
}

func (vp *ViewPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-vp.do:
			f(vp)
		default:
			return
		}
	}

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

func (vp *ViewPanel) Draw() {
	if vp.IsActive() {
		vp.Clear()

		cam := vp.cam.Get(game.PropCamera).(*graphics.Camera)

		trans := cam.GetTransformer()

		// draw terrain
		if vp.rsys.Terrain != nil {
			// FIXME
			worldr := image.Rect(0, 0, 256, 256)

			inter := cam.GetWorldIntersection(worldr)

			for x := inter.Min.X; x < inter.Max.X; x++ {
				for y := inter.Min.Y; y < inter.Max.Y; y++ {
					pos := trans(image.Pt(x, y))
					//vp.unsafesetcell(pos.X, pos.Y, vp.rsys.terrain.Cells[x][y].Cell)
					c := vp.rsys.Terrain.Cells[x][y].Cell
					vp.SetCell(pos.X, pos.Y, c.Ch, c.Fg, c.Bg)
				}
			}

		}

		// FIXME: check if objects are in our camera bounding box
		for _, actor := range vp.g.scene.Find(game.PropPos, game.PropStaticSprite) {
			pos := actor.Get(game.PropPos).(*game.Pos)
			sp := actor.Get(game.PropStaticSprite).(*game.StaticSprite)
			ipt := <-pos.Get()
			screenpos := trans(ipt)
			c := sp.GetCell()

			if glog.V(3) {
				glog.Infof("%s (%v) at %s drawn at %s", actor, c, ipt, screenpos)
				glog.Infof("cam %s", cam)
			}

			vp.unsafesetcell(screenpos.X, screenpos.Y, c)
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
}

func (vp *ViewPanel) unsafesetcell(x, y int, cell termbox.Cell) {
	vp.Buffer()[y*vp.width+x] = cell
}
