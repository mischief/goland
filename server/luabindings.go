package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gutil"
	"github.com/nsf/termbox-go"
	"github.com/stevedonovan/luar"
	"unicode/utf8"
  "image"
)

// make a new GameObject
func LuaNewGameObject(name string) game.Object {
	return game.NewGameObject(name)
}

var LuaGameObjectLib luar.Map = map[string]interface{}{
	"New": LuaNewGameObject,
}

func NewGlyph(ch string, fg string, bg string) termbox.Cell {
	newfg := gutil.StrToTermboxAttr(fg)
	newbg := gutil.StrToTermboxAttr(bg)

	r, _ := utf8.DecodeRuneInString(ch)

	newch := termbox.Cell{Ch: r, Fg: newfg, Bg: newbg}

	return newch
}

var LuaUtilLib luar.Map = map[string]interface{}{
	"NewGlyph": NewGlyph,
  "NewStaticSprite": game.NewStaticSprite,
  "IDGen": game.IDGen,
  "Pt": image.Pt,
}

func Lua_OpenObjectLib(L *lua.State) bool {
	luar.Register(L, "object", LuaGameObjectLib)
	luar.Register(L, "util", LuaUtilLib)

	return true
}
