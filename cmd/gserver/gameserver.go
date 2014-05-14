// GameServer: main gameserver struct and functions
package main

import (
	"os"
	"os/signal"
	"reflect"

	"github.com/golang/glog"

	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"

	"code.google.com/p/go9p/p"

	"github.com/chuckpreslar/emission"

	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gid"
	"github.com/mischief/goland/game/gobj"
	"github.com/mischief/goland/game/gterrain"
	"github.com/mischief/goland/game/gutil"
)

type GameServer struct {
	scene *game.Scene
	em    *emission.Emitter

	msys *game.MovementSystem
	tsys *gterrain.TerrainSystem

	DBM *DBManager

	fsys *GameFs

	closechan chan bool
	sigchan   chan os.Signal

	Objects *gobj.GameObjectMap

	// game object -> owner
	Owners map[gid.Gid]p.User

	config *gutil.LuaConfig

	lua   *lua.State
	debug bool
}

func NewGameServer(config *gutil.LuaConfig, ls *lua.State) (*GameServer, error) {
	gs := &GameServer{
		scene:     game.NewScene(),
		em:        emission.NewEmitter(),
		closechan: make(chan bool, 1),
		sigchan:   make(chan os.Signal, 1),
		config:    config,
		lua:       ls,
	}

	if debug, err := gs.config.Get("debug", reflect.Bool); err != nil {
		glog.Warning("'debug' not found in config or not a boolean. defaulting to false.")
		gs.debug = false
	} else {
		gs.debug = debug.(bool)
	}

	// objects setup
	gs.Objects = gobj.NewGameObjectMap()

	gs.Owners = make(map[gid.Gid]p.User)

	// lua state
	gs.lua = ls

	return gs, nil
}

func (gs *GameServer) Debug() bool {
	return gs.debug
}

func (gs *GameServer) Run() {
	gs.Start()

	run := true

	for run {
		select {
		case <-gs.closechan:
			glog.Info("got close signal")
			run = false
		case <-gs.sigchan:
			gs.closechan <- true
		}
	}

	gs.End()
}

func (gs *GameServer) Start() {
	var err error

	glog.Info("starting")

	if glog.V(2) {
		glog.Info("hooking signals")
	}

	signal.Notify(gs.sigchan, os.Interrupt)

	gs.DBM = NewDBManager(gs)

	if gs.msys, err = game.NewMovementSystem(gs.scene); err != nil {
		glog.Fatalf("movementsystem: %s", err)
	}

	gs.tsys = gterrain.NewTerrainSystem()

	gs.BindLua()

	// load assets
	glog.Info("loading assets")
	if gs.LoadAssets() != true {
		glog.Error("loading assets failed")
		return
	}

	gs.fsys = NewGameFs(gs)

	go func() {
		err := gs.fsys.Run()
		if err != nil {
			glog.Error(err)
		}
	}()

}

func (gs *GameServer) End() {
	glog.Info("stopping systems")

	gs.scene.StopSystems()

	glog.Info("systems stopped")
}

func (gs *GameServer) AddObject(obj gobj.Object, owner p.User) {
	glog.Infof("adding object %s owner %s", obj, owner)

	// tell clients about new object
	gs.Objects.Add(obj)
	gs.Owners[obj.GetID()] = owner
}

func (gs *GameServer) LuaLog(fmt string, args ...interface{}) {
	glog.Infof("lua: "+fmt, args...)
}

func (gs *GameServer) GetScriptPath() string {
	defaultpath := "../scripts/?.lua"
	if scriptconf, err := gs.config.Get("scriptpath", reflect.String); err != nil {
		glog.Warningf("GetScriptPath defaulting to %s: %s", defaultpath, err)
		return defaultpath
	} else {
		return scriptconf.(string)
	}
}

// TODO: move these bindings into another file
func (gs *GameServer) BindLua() {
	luar.Register(gs.lua, "", luar.Map{
		"gs": gs,
	})

	luar.Register(gs.lua, "sys", luar.Map{
		"msys": gs.msys,
		"tsys": gs.tsys,
	})

	// add our script path here..
	pkgpathscript := `package.path = package.path .. ";" .. gs.GetScriptPath() --";../?.lua"`
	if err := gs.lua.DoString(pkgpathscript); err != nil {
	}

	Lua_OpenObjectLib(gs.lua)
}

// load everything from lua scripts
func (gs *GameServer) LoadAssets() bool {
	if err := gs.lua.DoString("require('system')"); err != nil {
		luaerr := err.(*lua.LuaError)
		glog.Errorf("loadassets: %s", err)
		for _, f := range luaerr.StackTrace() {
			glog.Errorf("	%v", f)
		}
		return false
	}

	return true
}
