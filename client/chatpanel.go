package main

import (
	"bytes"
	"github.com/errnoh/termbox/panel"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"image"
	"time"
	"unicode/utf8"
)

// ChatPanel is a panel which runs the in-game chat and command input.
type ChatPanel struct {
	do              chan func(*ChatPanel)
	*panel.Buffered // Panel

	buf bytes.Buffer // Buffer for keyboard input

	enabled int32 // are we enabled or not?

	g    *Game
	nsys *ClientNetworkSystem
}

func NewChatPanel(g *Game, nsys *ClientNetworkSystem) *ChatPanel {
	cp := &ChatPanel{
		do: make(chan func(*ChatPanel), 10),
		g:  g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		cp.do <- func(cp *ChatPanel) {
			cp.resize(ev.Width, ev.Height)
		}
	})

	return cp
}

func (c *ChatPanel) Draw() {
	if c.Buffered != nil {
		c.Clear()
		str := "> " + c.buf.String()
		for i, r := range str {
			c.SetCell(i, 0, r, termbox.ColorBlue, termbox.ColorDefault)
		}
		c.Buffered.Draw()
	}
}

func (cp *ChatPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-cp.do:
			f(cp)
		default:
			return
		}
	}
}

func (cp *ChatPanel) HandleInput(ev termbox.Event) {
	cp.do <- func(cp *ChatPanel) {
		switch ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				cp.buf.WriteRune(ev.Ch)
			} else {
				switch ev.Key {
				case termbox.KeySpace:
					// just add a space
					cp.buf.WriteRune(' ')

				case termbox.KeyBackspace:
					fallthrough

				case termbox.KeyBackspace2:
					// on backspace, remove the last rune in the buffer
					if cp.buf.Len() > 0 {
						_, size := utf8.DecodeLastRune(cp.buf.Bytes())
						cp.buf.Truncate(cp.buf.Len() - size)
					}

				case termbox.KeyCtrlU:
					// clear the buffer, like a UNIX terminal
					cp.buf.Reset()

				case termbox.KeyEnter:
					// input confirmed, send it
					if cp.buf.Len() > 0 {
						cp.g.nsys.SendPacket("Tchat", cp.buf.String())
						cp.buf.Reset()
					}

					cp.g.rsys.PopInputHandler()
				case termbox.KeyEsc:
					// input cancelled
					cp.buf.Reset()
					cp.g.rsys.PopInputHandler()
				}
			}

		}
	}

}

func (cp *ChatPanel) resize(w, h int) {
	r := image.Rect(w-1, h-2, w/2, h-1)
	cp.Buffered = panel.NewBuffered(r, graphics.BorderStyle)
	cp.SetTitle("chat", graphics.TitleStyle)
}
