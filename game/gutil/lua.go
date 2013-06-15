// provides some handy lua functions
package gutil

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

// TODO: remove this luaparmap in favor of luar's proxies

// LuaParMap: go proxy for lua string-value table
// currently supporting anything tostring()'able
type LuaParMap struct {
	ls  *lua.State
	ref int
}

// create a new LuaParMap from table at idx
// return nil, error on failure
func NewLuaParMap(L *lua.State, idx int) (*LuaParMap, error) {
	pm := LuaParMap{L, 0}

	if err := LuaCheckIs(L, -1, "table"); err != nil {
		return nil, err
	}

	pm.ref = L.Ref(lua.LUA_REGISTRYINDEX)

	return &pm, nil
}

// get value for key in stored table reference
// returns "", false on failure
func (pm *LuaParMap) Get(key string) (value string, ok bool) {
	if pm.ls == nil || pm.ref == lua.LUA_NOREF {
		return "", false
	}

	pm.ls.RawGeti(lua.LUA_REGISTRYINDEX, pm.ref)

	if err := LuaCheckIs(pm.ls, -1, "table"); err != nil {
		return "", false
	}

	pm.ls.PushString(key)
	pm.ls.RawGet(-2)

	//	if err := LuaCheckIs(pm.ls, -1, "string"); err != nil {
	//		return "", false
	//	}

	value = pm.ls.ToString(-1)
	if value == "" {
		return value, false
		//return nil, errors.New(fmt.Sprintf("Key %s: Can't convert value %s to string", key, pm.ls.LTypename(-1)))
	}

	pm.ls.Pop(1)

	return value, true
}

// type of iterator closure function
type IterFunc func() (key, value string, ok bool)

// default IterFunc for bad calls
var badf = func() (string, string, bool) { return "", "", false }

// generate an IterFunc that returns succesive key-value pairs
// of the table stored in the registry at pm.ref
func (pm *LuaParMap) Iter() IterFunc {
	if pm.ls == nil || pm.ref == lua.LUA_NOREF {
		return badf
	}

	pm.ls.RawGeti(lua.LUA_REGISTRYINDEX, pm.ref)

	if err := LuaCheckIs(pm.ls, -1, "table"); err != nil {
		return badf
	}

	pm.ls.PushNil()

	return func() (string, string, bool) {
		if pm.ls.Next(-2) != 0 {
			if err := LuaCheckIs(pm.ls, -2, "string"); err != nil {
				return "", "", false
			}

			key := pm.ls.ToString(-2)
			value := pm.ls.ToString(-1)

			pm.ls.Pop(1)

			return key, value, true
		}

		return "", "", false

	}
}

// make a new LuaParMap from a file
// returns nil, error on failure
func LuaParMapFromFile(L *lua.State, filename string) (*LuaParMap, error) {
	if L.LoadFile(filename) != 0 {
		return nil, errors.New(L.CheckString(-1))
	}

	if err := L.Call(0, 1); err != nil {
		return nil, err
	}

	return NewLuaParMap(L, -1)
}

// check if type at idx on the stack is expected
// returns error on failure
func LuaCheckIs(L *lua.State, idx int, expected string) error {
	got := L.LTypename(idx)
	if got != expected {
		return errors.New(fmt.Sprintf("%s: %s expected, got %s", L.ToString(idx), expected, got))
	}

	return nil
}

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
	L.AtPanic(LuaAtPanic)

	L.OpenLibs()
	L.DoString("math.randomseed( os.time() )")

	return L
}

// Wrapper around lua_call that traps panics as errors
// returns nil if no error
func LuaSafeCall(L *lua.State, nargs, nresults int) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			if _, ok := err2.(error); ok {
				err = err2.(error)
			}
			return
		}
	}()

	err = nil

	L.Call(nargs, nresults)

	return

}
