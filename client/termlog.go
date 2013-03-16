package main

import (
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"image"
)

type TermLog struct {
	size  image.Point // log WxH
	lines int         // number of lines to show

	messages     []string
	start, count int // tracking for messages
}

func NewTermLog(size image.Point) *TermLog {
	tl := &TermLog{size: size, lines: size.Y}

	tl.messages = make([]string, tl.lines)

	tl.start = 0
	tl.count = 0

	return tl
}

// Draw our log on buf starting at x,y of pos.
func (tl *TermLog) Draw(buf *tulib.Buffer, pos image.Point) {

	// clear
	for y := 0; y < tl.size.Y; y++ {
		for x := 0; x < tl.size.X; x++ {
			buf.Set(pos.X+x, pos.X+y, termbox.Cell{Ch: ' '})
		}
	}

	y := pos.Y

	lineparams := &tulib.LabelParams{termbox.ColorGreen, termbox.ColorBlack, tulib.AlignLeft, '.', false}

	// please fix this mess
	if tl.start+tl.count > tl.lines {
		for _, line := range tl.messages[tl.start:] {
			linerect := tulib.Rect{pos.X, y, tl.size.X, 1}
			buf.DrawLabel(linerect, lineparams, []byte(line))
			y++
		}
		for _, line := range tl.messages[:tl.start+tl.count-tl.lines] {
			linerect := tulib.Rect{pos.X, y, tl.size.X, 1}
			buf.DrawLabel(linerect, lineparams, []byte(line))
			y++
		}
	} else {
		for _, line := range tl.messages[tl.start : tl.start+tl.count] {
			linerect := tulib.Rect{pos.X, y, tl.size.X, 1}
			buf.DrawLabel(linerect, lineparams, []byte(line))
			y++
		}
	}
}

func (tl *TermLog) Write(p []byte) (n int, err error) {
	end := (tl.start + tl.count) % tl.lines

	tl.messages[end] = string(p)

	if tl.count == tl.lines {
		// we're at the end, just start overwriting
		tl.start = (tl.start + 1) % tl.lines
	} else {
		tl.count++
	}

	return len(p), nil
}
