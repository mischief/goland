package gutil

import (
	termbox "github.com/nsf/termbox-go"
)

var attrmap = map[string]termbox.Attribute{
	// Colors
	"default": termbox.ColorDefault,
	"black":   termbox.ColorBlack,
	"red":     termbox.ColorRed,
	"green":   termbox.ColorGreen,
	"yellow":  termbox.ColorYellow,
	"blue":    termbox.ColorBlue,
	"magenta": termbox.ColorMagenta,
	"cyan":    termbox.ColorCyan,
	"white":   termbox.ColorWhite,

	// Attributes
	"bold":      termbox.AttrBold,
	"underline": termbox.AttrUnderline,
	"reverse":   termbox.AttrReverse,
}

func StrToTermboxAttr(str string) termbox.Attribute {
	attr, ok := attrmap[str]
	if ok {
		return attr
	}

	return attrmap["default"]
}

var keymap = map[string]termbox.Key{
	"enter": termbox.KeyEnter,
	"esc":   termbox.KeyEsc,
	"space": termbox.KeySpace,
}

func StrToKey(str string) termbox.Key {
	return keymap[str]
}
