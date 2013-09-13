package main

import (
	"bytes"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"time"
	"unicode/utf8"
)

// ChatPanel is a panel which runs the in-game chat and command input.
type ChatPanel struct {
	do chan func(*ChatPanel)
	*graphics.BasePanel

	buf bytes.Buffer // Buffer for keyboard input

	g    *Game
	nsys *ClientNetworkSystem
}

func NewChatPanel(g *Game) *ChatPanel {
	cp := &ChatPanel{
		do:        make(chan func(*ChatPanel), 10),
		BasePanel: graphics.NewPanel(),
		g:         g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		cp.do <- func(cp *ChatPanel) {
			cp.Resize(ev.Width, ev.Height)
		}
	})

	return cp
}

func (c *ChatPanel) Draw() {
	if c.Buffered != nil {
		c.Clear()
		str := "> " + c.buf.String()
		for i, r := range str {
			c.SetCell(i, 0, r, graphics.PromptStyle.Fg, graphics.PromptStyle.Bg)
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
						cp.g.SendChat(cp.buf.String())
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
