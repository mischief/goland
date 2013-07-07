package main

import (
	"bytes"
	"github.com/mischief/goland/game/gnet"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

type ChatBuffer struct {
	bytes.Buffer

	Input chan termbox.Event

	g    *Game
	term *Terminal
}

func NewChatBuffer(g *Game, t *Terminal) *ChatBuffer {
	cb := &ChatBuffer{Input: make(chan termbox.Event), g: g, term: t}

	go cb.HandleInput()

	return cb
}

func (c *ChatBuffer) HandleInput() {
	for e := range c.Input {
		if e.Ch != 0 {
			c.WriteRune(e.Ch)
		} else {
			switch e.Key {
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
	}
}
