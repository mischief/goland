// provides some handy lua functions
package gutil

import (
	"errors"
	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

// function that lua will call before aborting.
// we override this because normally lua longjmps,
// but that breaks go's defer/panic. so we just panic.
func LuaAtPanic(L *lua.State) int {
	panic(errors.New(L.ToString(-1)))
	return 0
}

// TODO: better error handling
func LuaInit() *lua.State {
	L := luar.Init()
	//L.AtPanic(LuaAtPanic)

	L.OpenLibs()
	L.DoString("math.randomseed( os.time() )")

	return L
}
