package graphics

import (
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/golang/glog"
	"github.com/mischief/goland/game/gutil"
	termbox "github.com/nsf/termbox-go"
	"image"
	"sync/atomic"
	"time"
)

var (
	TextStyle   = termbox.Cell{}
	TitleStyle  = termbox.Cell{}
	BorderStyle = termbox.Cell{Ch: 's'}
	PromptStyle = termbox.Cell{}

	defaultsize = image.Rect(0, 0, 80, 24)
)

// Interface for panels which can handle input events
type InputHandler interface {
	HandleInput(ev termbox.Event)
}

type GamePanel interface {
	panel.Panel
	gutil.Updater
	//InputHandler

	Activator

	GetTitle() string
	Title(string) GamePanel
	TitleStyle(s termbox.Cell) GamePanel
	SetBord(b termbox.Cell) GamePanel
	SizeFn(MakeSizeFn)
	Resize(w, h int) (int, int)
}

type Activator interface {
	Activate()
	Deactivate()
	IsActive() bool
}

// A function which will give the window coordinates based on current screen size
type MakeSizeFn func(w, h int) image.Rectangle

var (
	defsizefn = func(w, h int) image.Rectangle {
		return image.Rect(1, 1, w-1, h-1)
	}
)

func WriteCenteredLine(p panel.Panel, str string, y int, fg, bg termbox.Attribute) {
	startx := (p.Bounds().Dx() / 2) - (len(str) / 2)

	for i, r := range str {
		p.SetCell(i+startx, y, r, fg, bg)
	}
}

type BasePanel struct {
	// embedded panel
	*panel.Buffered

	// activated for drawing
	active int32

	sizefn MakeSizeFn

	// pos as 0..1
	posx, posy float32

	// width/heigh as 0..1
	width, height float32

	// limits of panel. limits.Min represents the minimum w/h,
	// and limits.Max represents the maxmium w/h
	limits image.Rectangle

	// title
	title      string
	titleStyle termbox.Cell

	// border
	borderStyle termbox.Cell

	Rsys *RenderSystem
}

func NewPanel() *BasePanel {
	bp := &BasePanel{
		Buffered:    panel.NewBuffered(defaultsize, BorderStyle),
		sizefn:      defsizefn,
		titleStyle:  TitleStyle,
		borderStyle: BorderStyle,
	}

	return bp
}

func (bp *BasePanel) String() string {
	return fmt.Sprintf("%s pos %f,%f size %f,%f", bp.title, bp.posx, bp.posy, bp.width, bp.height)
}

func (bp *BasePanel) Update(delta time.Duration) {
}

func (bp *BasePanel) Activate() {
	atomic.StoreInt32(&bp.active, 1)
}

func (bp *BasePanel) Deactivate() {
	atomic.StoreInt32(&bp.active, 0)
}

func (bp *BasePanel) IsActive() bool {
	if atomic.LoadInt32(&bp.active) == 1 {
		return true
	}

	return false
}

func (bp *BasePanel) SizeFn(fn MakeSizeFn) {
	bp.sizefn = fn
}

// Reposition panel 0..1
func (bp *BasePanel) Pos(x, y float32) *BasePanel {
	bp.posx = x
	bp.posy = y

	return bp
}

// Set size of panel 0..1
func (bp *BasePanel) Size(width, height float32) *BasePanel {
	bp.width = width
	bp.height = height

	return bp
}

// Set the absolute size limits of this panel.
// l.Min represents minimum width/height
// l.Max represents maximum width/height
// Sizes are specified as # of lines for height, and # of columns for width
func (bp *BasePanel) SetLimits(l image.Rectangle) *BasePanel {
	bp.limits = l
	return bp
}

// Set title and title style
func (bp *BasePanel) Title(t string) GamePanel {
	bp.title = t

	return bp
}

func (bp *BasePanel) GetTitle() string {
	return bp.title
}

func (bp *BasePanel) TitleStyle(s termbox.Cell) GamePanel {
	bp.titleStyle = s

	return bp
}

// Set border style
func (bp *BasePanel) SetBord(b termbox.Cell) GamePanel {
	bp.borderStyle = b

	return bp
}

// Resize this panel.
// Should be called whenever the screen size changes.
// w, h is the real size of the screen
// returns the computed width and height of the panel
func (bp *BasePanel) Resize(w, h int) (rw, rh int) {
	/*
			realposx := int(float32(w) * bp.posx)
			realposy := int(float32(h) * bp.posy)
			realwidth := int(float32(w) * bp.width)
			realheight := int(float32(h) * bp.height)

			screenlimit := image.Rect(1, 1, w-1, h-1)

			if realwidth < bp.limits.Min.X {
				realwidth = bp.limits.Min.X
			} else if bp.limits.Max.X > 0 && realwidth > bp.limits.Max.X {
				realwidth = bp.limits.Max.X
			}

			if realheight < bp.limits.Min.Y {
				realheight = bp.limits.Min.Y
			} else if bp.limits.Max.Y > 0 && realheight > bp.limits.Max.Y {
				realheight = bp.limits.Max.Y
			}

			realrect := image.Rect(realposx-realwidth/2.0, realposy-realheight/2.0, realposx+realwidth/2.0, realposy+realheight/2.0).Intersect(screenlimit)

		  if realrect.Dy() < 1 {
		    realrect.Max.Y++
		  }

			//if insrect.Dx() > 1 && insrect.Dy() > 1 {
			bp.Buffered = panel.NewBuffered(realrect, bp.borderStyle)
			//}
	*/
	rect := bp.sizefn(w, h)
	bp.Buffered = panel.NewBuffered(rect, bp.borderStyle)

	bp.SetTitle(bp.title, bp.titleStyle)

	if glog.V(3) {
		glog.Infof("resize %s real %s limit %s", bp, rect, bp.limits)
	}

	return rect.Dx(), rect.Dy()
}
