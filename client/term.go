package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"image"
	"log"
)

type KeyHandler func(ev termbox.Event)

type Terminal struct {
	tulib.Buffer
	EventChan chan termbox.Event

	runehandlers map[rune]KeyHandler
	keyhandlers  map[termbox.Key]KeyHandler
}

func (t *Terminal) Size() image.Point {
	return image.Point{t.Rect.Width, t.Rect.Height}
}

func (t *Terminal) Start() error {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	t.Buffer = tulib.TermboxBuffer()

	t.EventChan = make(chan termbox.Event)

	// event generator
	go func(e chan termbox.Event) {
		for {
			e <- termbox.PollEvent()
		}
	}(t.EventChan)

	t.runehandlers = make(map[rune]KeyHandler)
	t.keyhandlers = make(map[termbox.Key]KeyHandler)

	return nil
}

func (t *Terminal) End() {
	termbox.Close()
}

func (t *Terminal) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func (t *Terminal) Flush() {
	termbox.Flush()
}

func (t *Terminal) RunInputHandlers() error {
	select {
	case ev := <-t.EventChan:
		log.Printf("Keypress: %s", tulib.KeyToString(ev.Key, ev.Ch, ev.Mod))

		if ev.Ch != 0 { // this is a character
			if handler, ok := t.runehandlers[ev.Ch]; ok {
				handler(ev)
			}
		} else {
			if handler, ok := t.keyhandlers[ev.Key]; ok {
				handler(ev)
			}
		}

	default:
	}

	return nil
}

func (t *Terminal) HandleRune(r rune, h KeyHandler) {
	t.runehandlers[r] = h
}

func (t *Terminal) HandleKey(k termbox.Key, h KeyHandler) {
	t.keyhandlers[k] = h
}

func (t *Terminal) PrintCell(x, y int, ch termbox.Cell) {
	termbox.SetCell(x, y, ch.Ch, ch.Fg, ch.Bg)
}

func (t *Terminal) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (t *Terminal) Printf(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	t.Print(x, y, fg, bg, s)
}
