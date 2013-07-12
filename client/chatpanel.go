package main

import (
	"bytes"
	"github.com/errnoh/termbox/panel"
	"github.com/mischief/goland/game/gnet"
	"github.com/nsf/termbox-go"
	"image"
	"sync"
	"unicode/utf8"
)

// ChatPanel is a panel which runs the in-game chat and command input.
type ChatPanel struct {
	*panel.Buffered // Panel
	bytes.Buffer    // Buffer for keyboard input
	m               sync.Mutex
	Input           chan termbox.Event
	g               *Game
	term            *Terminal
}

func NewChatPanel(g *Game, t *Terminal) *ChatPanel {
	cb := &ChatPanel{Input: make(chan termbox.Event), g: g, term: t}

	cb.HandleInput(termbox.Event{Type: termbox.EventResize})

	return cb
}

func (c *ChatPanel) HandleInput(ev termbox.Event) {
	c.m.Lock()
	switch ev.Type {
	case termbox.EventKey:
		if ev.Ch != 0 {
			c.WriteRune(ev.Ch)
		} else {
			switch ev.Key {
			case termbox.KeySpace:
				// just add a space
				c.WriteRune(' ')

			case termbox.KeyBackspace:
				fallthrough

			case termbox.KeyBackspace2:
				// on backspace, remove the last rune in the buffer
				if c.Len() > 0 {
					_, size := utf8.DecodeLastRune(c.Bytes())
					c.Truncate(c.Len() - size)
				}

			case termbox.KeyCtrlU:
				// clear the buffer, like a UNIX terminal
				c.Reset()

			case termbox.KeyEnter:
				// input confirmed, send it
				if c.Len() > 0 {
					c.g.SendPacket(gnet.NewPacket("Tchat", c.String()))
					c.Reset()
					c.term.SetAltChan(nil)

				}
			case termbox.KeyEsc:
				// input cancelled
				c.Reset()
				c.term.SetAltChan(nil)
			}
		}
	case termbox.EventResize:
		w, h := termbox.Size()
		r := image.Rect(w-1, h-2, w/2, h-1)
		c.Buffered = panel.NewBuffered(r, termbox.Cell{'s', termbox.ColorGreen, 0})

	}

	c.m.Unlock()
}

func (c *ChatPanel) Draw() {
	c.Clear()
	c.m.Lock()
	str := "Chat: " + c.Buffer.String()
	c.m.Unlock()
	for i, r := range str {
		c.SetCell(i, 0, r, termbox.ColorBlue, termbox.ColorDefault)
	}
	c.Buffered.Draw()
}
