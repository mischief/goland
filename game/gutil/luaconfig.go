package gutil

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
	"reflect"
	"strings"
	"sync"
)

type LuaConfigElement struct {
	Key   string
	Value interface{}
}

// LuaConfig conveniently wraps a lua file as a go map,
// and config elements can be accesses through Get().
type LuaConfig struct {
	file string //
	conf map[string]interface{}
	m    sync.Mutex
}

// Construct a new LuaConfig given the lua state and file name.
// Returns *LuaConfig, nil on success and nil, error on error.
// Expects that the file provided is a lua script that will return a
// table of strings to values, for example:
// config = {
//   key = "value",
//   boolean = false,
// }
// return config
//
func NewLuaConfig(lua *lua.State, file string) (*LuaConfig, error) {
	lc := &LuaConfig{file: file}

	if err := lua.DoFile(file); err != nil {
		return nil, fmt.Errorf("NewLuaConfig: Can't load %s: %s", file, err)
	} else {
		m := luar.CopyTableToMap(lua, nil, -1)
		lc.conf = m.(map[string]interface{})
	}

	return lc, nil
}

// Get will walk the config for key, and assert that its value is of Kind expected.
// Returns value, nil on success and nil, error on error.
// Get will accept keys like "table.subtable.key", and will walk tables until it finds
// the last element in the key, or abort on error.
func (lc *LuaConfig) Get(key string, expected reflect.Kind) (res interface{}, err error) {
	lc.m.Lock()
	defer lc.m.Unlock()

	parts := strings.Split(key, ".")

	m := lc.conf

	for i, p := range parts {
		val, ok := m[p]
		if !ok {
			return nil, fmt.Errorf("LuaConfig: %s: no key named %s", lc.file, p)
		} else {
			kind := reflect.TypeOf(val).Kind()
			if i+1 == len(parts) {
				if kind != expected {
					return nil, fmt.Errorf("LuaConfig: %s: key %s is type %s, expected %s", lc.file, kind, expected)
				} else {
					return val, nil
				}
			} else {
				if newm, ok := val.(map[string]interface{}); ok {
					m = newm
				} else {
					return nil, fmt.Errorf("LuaConfig: %s: key %s is not a table", lc.file, p)
				}
			}
		}
	}

	return
}

// Gives a channel which will contain top-level config elements
func (lc *LuaConfig) Chan() <-chan LuaConfigElement {
	lc.m.Lock()
	defer lc.m.Unlock()

	ch := make(chan LuaConfigElement, len(lc.conf))

	for k, v := range lc.conf {
		ch <- LuaConfigElement{k, v}
	}

	close(ch)

	return ch
}
