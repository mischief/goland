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

type LuaConfig struct {
	file string //
	conf map[string]interface{}
	m    sync.Mutex
}

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
				// TODO: advance one map down
				return nil, fmt.Errorf("I can't walk yet!")
			}
		}
	}

	return
}

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
