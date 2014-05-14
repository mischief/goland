package main

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/chuckpreslar/emission"
	"github.com/golang/glog"
	"github.com/mischief/goland/client/graphics"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gid"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gobj"
	"github.com/mischief/goland/game/gterrain"
	"github.com/mischief/goland/game/gutil"
	"github.com/nsf/termbox-go"
	"github.com/stevedonovan/luar"
	"image"
	"net"
	"reflect"
	"time"
)

type Game struct {
	scene *game.Scene

	rsys *graphics.RenderSystem
	msys *game.MovementSystem
	nsys *ClientNetworkSystem

	// Event emitter/handler
	em *emission.Emitter

	closechan chan bool

	me      *gobj.Object
	objects map[gid.Gid]gobj.Object
	//Objects *game.GameObjectMap
	//Map     *game.MapChunk
	terrain *gterrain.TerrainChunk

	// config
	config *gutil.LuaConfig

	// lua
	lua *lua.State

	// server connection
	ServerCon net.Conn

	// read/write chanio on server connection
	ServerRChan <-chan interface{}
	ServerWChan chan<- interface{}
}

func NewGame(config *gutil.LuaConfig, lua *lua.State) *Game {
	g := Game{
		scene:     game.NewScene(),
		em:        emission.NewEmitter(),
		objects:   make(map[gid.Gid]gobj.Object),
		config:    config,
		lua:       lua,
		closechan: make(chan bool, 1),
	}

	g.em.SetMaxListeners(10000)

	// make player

	/*
	  pl := g.scene.Add("player")

	  pos := g.msys.Pos()
	  pos.Set(image.Pt(0,0))

	  sp := game.NewStaticSprite("player", termbox.Cell{'@', 0, 0})

	  pl.Add(pos)
	  pl.Add(sp)
	*/

	return &g
}

// Broadcast an error message and log it.
func (g *Game) Log(s string) {
	g.em.Emit("log", s)
	glog.Info(s)
}

func (g *Game) SendChat(c string) {
	g.nsys.SendPacket("Tchat", c)
}

func (g *Game) Quit() {
	g.nsys.SendPacket("Tquit", nil)
	g.closechan <- true
}

func (g *Game) Run() {
	if err := g.Config(); err != nil {
		glog.Fatalf("config: %s", err)
	}

	g.BindLua()

	g.Start()

	ticker := time.NewTicker(time.Second)

	run := true

	for run {
		select {
		case <-ticker.C:
			glog.Flush()

		case <-g.closechan:
			glog.Infof("got close signal")
			run = false
		}
	}

	g.End()

}

func (g *Game) Start() {
	var err error

	glog.Info("starting")

	// config items
	// systems
	if g.rsys, err = graphics.NewRenderSystem(g.scene, g.em); err != nil {
		glog.Fatalf("rendersystem: %s", err)
	}

	if g.msys, err = game.NewMovementSystem(g.scene); err != nil {
		glog.Fatalf("movementsystem: %s", err)
	}

	// Intro panel
	intro := g.IntroPanel()
	intro.Title("intro").TitleStyle(graphics.TitleStyle)
	g.rsys.AddPanel(intro)
	g.rsys.PushActivePanelName("intro")

	// camera
	c := g.scene.Add("camera")
	cpp := g.msys.Pos()
	cpp.Set(image.Pt(128, 128))
	c.Add(cpp)
	cc := g.rsys.Cam(c.ID)
	cc.SetCenter(<-cpp.Get()) // is this necessary?
	c.Add(cc)

	// graphics panels
	sp := NewStatsPanel(g)
	sp.Title("stats").TitleStyle(graphics.TitleStyle)
	sp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(1, 1, w-1, 2)
	})
	sp.Activate()
	g.rsys.AddPanel(sp)

	// viewport
	vp := NewViewPanel(g, c)
	vp.Title("view").TitleStyle(graphics.TitleStyle)
	vp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(1, 3, w-1, h-8)
	})
	vp.Activate()
	g.rsys.AddPanel(vp)

	// log/chat
	logp := NewLogPanel(g)
	logp.Title("log").TitleStyle(graphics.TitleStyle)
	logp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(1, h-7, w-1, h-3)
	})
	logp.Activate()
	g.rsys.AddPanel(logp)

	// player stats
	playerp := NewPlayerPanel(g)
	playerp.Title("player").TitleStyle(graphics.TitleStyle)
	playerp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(1, h-2, w/2, h-1)
	})
	playerp.Activate()
	g.rsys.AddPanel(playerp)

	// chat box
	chatp := NewChatPanel(g)
	chatp.Title("chat").TitleStyle(graphics.TitleStyle)
	chatp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(w-1, h-2, w/2, h-1)
	})
	chatp.Activate()
	g.rsys.AddPanel(chatp)

	// console
	cons := NewConsolePanel(g)
	cons.Title("console").TitleStyle(graphics.TitleStyle)
	cons.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(4, 4, w-4, h-4)
	})
	g.rsys.AddPanel(cons)

	// quit dialog
	qp := NewQuitPanel(g)
	qp.Title("quit").TitleStyle(graphics.TitleStyle)
	qp.SizeFn(func(w, h int) image.Rectangle {
		return image.Rect(w/2-10, h/2-2, w/2+10, h/2+2)
	})
	g.rsys.AddPanel(qp)

	// resize all panels
	w, h := termbox.Size()
	g.em.Emit("resize", termbox.Event{Type: termbox.EventResize, Width: w, Height: h})

	// do keybindings
	g.SetupKeys()

	// network setup
	sconf := "server"
	sc, err := g.config.Get(sconf, reflect.String)
	if err != nil {
		g.Log(fmt.Sprintf("no %q in config: %s", sconf, err))
		//glog.Errorf("start: missing server in config: %s", err)
		//logpanel.addline(fmt.Sprintf("connection to the server failed: %s", err))
	}

	server := sc.(string)

	// setup network
	if g.nsys, err = NewClientNetworkSystem(g, g.scene, server); err != nil {
		g.Log(fmt.Sprintf("clientnetworksystem: %s", err))
	} else {
		if err := g.Login(); err != nil {
			g.Log(fmt.Sprintf("login error: %s", err))
		}
	}

	glog.Info("start complete")
	glog.Flush()
}

// load things from config
// caller should probably bail on error
// dont return error if the problematic option is not fatal, just print to log
func (g *Game) Config() error {
	glog.Info("reading config items")

	if txtstyle, err := g.config.Get("theme.textfg", reflect.String); err == nil {
		graphics.TextStyle.Fg = gutil.StrToTermboxAttr(txtstyle.(string))
	} else {
		glog.Infof("config: %s", err)
	}
	if tstyle, err := g.config.Get("theme.titlefg", reflect.String); err == nil {
		graphics.TitleStyle.Fg = gutil.StrToTermboxAttr(tstyle.(string))
	} else {
		glog.Infof("config: %s", err)
	}
	if bstyle, err := g.config.Get("theme.borderfg", reflect.String); err == nil {
		graphics.BorderStyle.Fg = gutil.StrToTermboxAttr(bstyle.(string))
	} else {
		glog.Infof("config: %s", err)
	}
	if pstyle, err := g.config.Get("theme.promptfg", reflect.String); err == nil {
		graphics.PromptStyle.Fg = gutil.StrToTermboxAttr(pstyle.(string))
	} else {
		glog.Infof("config: %s", err)
	}

	return nil
}

func (g *Game) BindLua() {
	luar.Register(g.lua, "", luar.Map{
		"g": g,
	})
}

func (g *Game) Login() error {

	// login
	username, err := g.config.Get("username", reflect.String)
	if err != nil {
		glog.Infof("start: missing username in config: %s", err)
		return fmt.Errorf("missing username name in config, can't log in!")
	}

	g.nsys.SendPacket("Tlogin", username)

	// request the map from server
	g.nsys.SendPacket("Tgetterrain", nil)

	// request the object we control
	// XXX: the delay is to fix a bug regarding ordering of packets.
	// if the client gets the response to this before he is notified
	// that the object exists, it will barf, so we delay this request.
	time.AfterFunc(50*time.Millisecond, func() {
		g.nsys.SendPacket("Tgetplayer", nil)
	})

	return nil
}

func (g *Game) SetupKeys() {

	keyskey := "keys"

	// gives us a map[string] interface{}
	kbconf, err := g.config.RawGet(keyskey)
	if err != nil {
		glog.Infof("%q not found in config, no key bindings", keyskey)
		return
	}
	keybinds := kbconf.(map[string]interface{})

	keyactions := map[string]func(g *Game, a string){
		// movement
		"moveup": func(g *Game, a string) {
			g.nsys.SendPacket("Tmove", a)
		},
		"moveleft": func(g *Game, a string) {
			g.nsys.SendPacket("Tmove", a)
		},
		"movedown": func(g *Game, a string) {
			g.nsys.SendPacket("Tmove", a)
		},
		"moveright": func(g *Game, a string) {
			g.nsys.SendPacket("Tmove", a)
		},

		// actions
		"pickup": func(g *Game, a string) {
			g.nsys.SendPacket("Taction", a)
		},
		"drop": func(g *Game, a string) {
			g.nsys.SendPacket("Taction", a)
		},

		// make this pop up a dialog
		"inventory": func(g *Game, a string) {
			g.nsys.SendPacket("Taction", a)
		},

		// dialogs
		"quit": func(g *Game, a string) {
			g.rsys.PushActivePanelName(a)
		},
		"chat": func(g *Game, a string) {
			g.rsys.PushActivePanelName(a)
		},
		"console": func(g *Game, a string) {
			g.rsys.PushActivePanelName(a)
		},
	}

	for key, actconf := range keybinds {
		act := actconf.(string)
		glog.Infof("binding %q to %q", key, act)

		// check if it's just a rune. if not, it's a 'key' like esc or space
		if len(key) == 1 {
			g.rsys.HandleRune(rune(key[0]), func(ev termbox.Event) {
				keyactions[act](g, act)
			})
		} else {
			termkey := gutil.StrToKey(key)
			if termkey == termbox.Key(0) {
				glog.Infof("invalid key %s, not binding", key)
			} else {
				g.rsys.HandleKey(termkey, func(ev termbox.Event) {
					keyactions[act](g, act)
				})
			}
		}
	}

}

func (g *Game) End() {
	glog.Info("stopping systems")

	g.scene.StopSystems()

	glog.Info("ending")
	glog.Flush()
}

// deal with gnet.Packets received from the server
// XXX: old, remove
func (g *Game) HandlePacket(pk *gnet.Packet) {
	defer func() {
		if err := recover(); err != nil {
			glog.Infof("handlepacket: %s", err)
		}
	}()

	if glog.V(1) {
		glog.Infof("handlepacket: got packet %s", pk)
	}
	switch pk.Tag {

	// Rchat: we got a text message
	case "Rchat":
		//chatline := pk.Data.(string)
		//io.WriteString(g.logpanel, chatline)

	// Raction: something moved on the server
	// Need to update the objects (sync client w/ srv)
	case "Raction":
		//robj := pk.Data.(game.Object) // remote object

		/*for o := range g.Objects.Chan() {
			if o.GetID() == robj.GetID() {
				o.SetPos(robj.GetPos())
			} /*else if o.GetTag("item") {
				item := g.Objects.FindObjectByID(o.GetID())
				if item.GetTag("gettable") {
					item.SetPos(o.GetPos())
				} else {
					g.Objects.RemoveObject(item)
				}
			}
		}*/

		// Rnewobject: new object we need to track
	case "Rnewobject":
		obj := pk.Data[0].(gobj.Object)

		//name := fmt.Sprintf("%s%d", obj.GetName(), obj.GetID())

		//a := g.scene.Add(name)

		pos := g.msys.Pos()
		pos.Set(image.Pt(obj.GetPos()))

		//sp := game.NewStaticSprite(name, obj.GetGlyph())

		//a.Add(pos)
		//a.Add(sp)

		// Rdelobject: some object went away
	case "Rdelobject":
		//obj := pk.Data.(game.Object)
		//g.Objects.RemoveObject(obj)

		// Rgetplayer: find out who we control
	case "Rgetplayer":
		//playerid := pk.Data.(int)

		/*pl := g.Objects.FindObjectByID(playerid)
		if pl != nil {
			g.pm.Lock()
			g.player = pl
			g.pm.Unlock()
		} else {
			glog.Infof("Game: HandlePacket: can't find our player %s", playerid)

			// just try again
			// XXX: find a better way
			time.AfterFunc(50*time.Millisecond, func() {
				g.ServerWChan <- gnet.NewPacket("Tgetplayer", nil)
			})
		}*/

	default:
		glog.Infof("bad packet tag %s", pk.Tag)
	}

}
