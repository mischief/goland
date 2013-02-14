package main

import (
  "fmt"
  "strings"
  "os"
  "github.com/nsf/termbox-go"
)

const (
  superficialSizeLimit = 24
  border = "#"
)

type Terminal struct {
  Width, Height int
  EventChan chan termbox.Event
}

func (t *Terminal) Start() {
  err := termbox.Init()
  if err != nil {
    panic(err)
  }

  t.Width, t.Height = termbox.Size()

  if t.Height < superficialSizeLimit {
    fmt.Println("terminal too small")
    t.End()
    os.Exit(1)
  }

  t.EventChan = make(chan termbox.Event)

  // event generator
  go func(e chan termbox.Event) {
    for {
      e <- termbox.PollEvent()
    }
  }(t.EventChan)

  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

}

func (t *Terminal) End() {
  termbox.Close()
}

func (t *Terminal) Draw() {
  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
  t.drawborder()
}

func (t *Terminal) Flush() {
  termbox.Flush()
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

func (t *Terminal) drawborder() {
  var y int

  t.Print(0, 0, termbox.ColorWhite, termbox.ColorBlack, strings.Repeat(border, t.Width))

  for y = 0; y < t.Height - 1; y++ {
    t.Print(0, y, termbox.ColorWhite, termbox.ColorBlack, border)
    t.Print(t.Width - 1, y, termbox.ColorWhite, termbox.ColorBlack, border)
  }

  t.Print(0, y, termbox.ColorWhite, termbox.ColorBlack, strings.Repeat(border, t.Width))

}

