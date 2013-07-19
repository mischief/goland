package graphics

import (
	"github.com/errnoh/termbox/panel"
	"github.com/mischief/goland/game/gutil"
	termbox "github.com/nsf/termbox-go"
)

// Interface for panels which can handle input events
type InputHandler interface {
	HandleInput(ev termbox.Event)
}

type GamePanel interface {
	panel.Panel
	gutil.Updater
	InputHandler
}
