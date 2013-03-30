package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game"
	"github.com/stevedonovan/luar"
	"reflect"
)

// make a new GameObject
func Lua_NewGameObject(L *lua.State) int {
	name := L.CheckString(-1)

	newobj := game.NewGameObject(name)
	luar.GoToLua(L, nil, reflect.ValueOf(newobj))

	return 1
}

var Lua_GameObjectLibf luar.Map = map[string]interface{}{
	"New": Lua_NewGameObject,
}

func Lua_OpenObjectLib(L *lua.State) bool {

	luar.Register(L, "object", Lua_GameObjectLibf)

	return true
}
