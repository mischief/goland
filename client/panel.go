package main

import (
	termbox "github.com/nsf/termbox-go"
)

// Interface for panels which can handle input events
type InputHandler interface {
	HandleInput(ev termbox.Event)
}
