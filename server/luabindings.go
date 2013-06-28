package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game"
	"github.com/nsf/termbox-go"
	"github.com/stevedonovan/luar"
	"reflect"
	"unicode/utf8"
)

// make a new GameObject
func Lua_NewGameObject(L *lua.State) int {
	name := L.CheckString(-1)

	newobj := game.NewGameObject(name)
	luar.GoToLua(L, nil, reflect.ValueOf(newobj), true)

	return 1
}

var Lua_GameObjectLibf luar.Map = map[string]interface{}{
	"New": Lua_NewGameObject,
}

func Lua_NewGlyph(L *lua.State) int {
	ch := L.CheckString(-1)
	//	/*fg*/ _ := L.CheckString(-2)
	//	/*bg*/ _ := L.CheckString(-3)

	r, _ := utf8.DecodeRuneInString(ch)

	newch := termbox.Cell{Ch: r}

	luar.GoToLua(L, nil, reflect.ValueOf(newch), true)

	return 1
}

var Lua_UtilLibf luar.Map = map[string]interface{}{
	"NewGlyph": Lua_NewGlyph,
}

func Lua_OpenObjectLib(L *lua.State) bool {

	luar.Register(L, "object", Lua_GameObjectLibf)
	luar.Register(L, "util", Lua_UtilLibf)

	return true
}
