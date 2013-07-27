package main

import (
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"time"
)

var (
  qfg = termbox.ColorBlue
  qbg = termbox.ColorDefault
)

// QuitPanel is responsible for handling game quitting.
type QuitPanel struct {
	do                  chan func(*QuitPanel)
	*graphics.BasePanel       // Panel
	active              int32 // are we enabled or not?
	g                   *Game
}

func NewQuitPanel(g *Game) *QuitPanel {
	qp := &QuitPanel{
		do: make(chan func(*QuitPanel), 10),
    BasePanel: graphics.NewBasePanel(g.rsys),
		g:  g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		qp.do <- func(qp *QuitPanel) {
      qp.Resize(ev.Width, ev.Height)
		}
	})

	return qp
}

func (qp *QuitPanel) Draw() {
	if qp.IsActive() {
		qp.Clear()
		q := "quit now?"
		ent := "enter to confirm"
		esc := "esc to cancel"
		graphics.WriteCenteredLine(qp, q, 0, qfg, qbg)
		graphics.WriteCenteredLine(qp, ent, 2, qfg, qbg)
		graphics.WriteCenteredLine(qp, esc, 3, qfg, qbg)

		qp.BasePanel.Draw()
	}
}

func (qp *QuitPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-qp.do:
			f(qp)
		default:
			return
		}
	}
}

func (qp *QuitPanel) HandleInput(ev termbox.Event) {
	qp.do <- func(qp *QuitPanel) {
		switch ev.Key {

		case termbox.KeyEnter:
			// yes! quit!
			qp.g.Quit()

		case termbox.KeySpace:
			// cancelled
			qp.g.rsys.PopInputHandler()
		}
	}

}
