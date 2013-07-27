package main

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/chuckpreslar/emission"
	"github.com/golang/glog"
	"github.com/mischief/goland/client/graphics"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	"github.com/nsf/termbox-go"
	"image"
	//"io"
	"net"
	"reflect"
	"time"
)

var (
	CARDINALS = map[rune]string{
	/*
		'w': game.DIR_UP,
		'k': game.DIR_UP,
		'a': game.DIR_LEFT,
		'h': game.DIR_LEFT,
		's': game.DIR_DOWN,
		'j': game.DIR_DOWN,
		'd': game.DIR_RIGHT,
		'l': game.DIR_RIGHT,
		',': game.ACTION_ITEM_PICKUP,
		'x': game.ACTION_ITEM_DROP,
		'i': game.ACTION_ITEM_LIST_INVENTORY,
	*/
	}
)

type Game struct {
	scene *game.Scene

	rsys *graphics.RenderSystem
	msys *game.MovementSystem
	nsys *ClientNetworkSystem

	// Event emitter/handler
	em *emission.Emitter

	closechan chan bool

	//Objects *game.GameObjectMap
	//Map     *game.MapChunk
	terrain *game.TerrainChunk

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

func (g *Game) Quit() {
	g.closechan <- true
}

func (g *Game) Run() {
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

	// camera
	c := g.scene.Add("camera")
	cpp := g.msys.Pos()
	cpp.Set(image.Pt(0, 0))
	c.Add(cpp)
	cc := g.rsys.Cam(c.ID)
	cc.SetCenter(<-cpp.Get()) // is this necessary?
	c.Add(cc)

	// graphics panels
  sp := NewStatsPanel(g)
  sp.Title("stats").TitleStyle(graphics.TitleStyle)
  sp.Pos(0.5, 0.05).Size(1.0, 0.1)
  sp.SetLimits(image.Rect(0, 1, 0, 0))
  sp.Activate()
	g.rsys.AddPanel("stats", sp)

  vp := NewViewPanel(g, c)
  vp.Title("view").TitleStyle(graphics.TitleStyle)
  vp.Pos(0.5, 0.5).Size(1.0, 0.55)
  vp.Activate()

	g.rsys.AddPanel("view", vp)

	logpanel := NewLogPanel(g)
	g.rsys.AddPanel("log", logpanel)
	g.rsys.AddPanel("player", NewPlayerPanel(g))
	g.rsys.AddPanel("chat", NewChatPanel(g, g.nsys))
	g.rsys.AddPanel("console", NewConsolePanel(g))

  qp := NewQuitPanel(g)
  qp.Title("quit").TitleStyle(graphics.TitleStyle)
  qp.Pos(0.5, 0.5).Size(0.15, 0.2)

	g.rsys.AddPanel("quit", qp)

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
	// ESC to quit
	g.rsys.HandleKey(termbox.KeyEsc, func(ev termbox.Event) {
		g.rsys.PushActivePanelName("quit")
	})

	// Enter to chat
	g.rsys.HandleKey(termbox.KeyEnter, func(ev termbox.Event) {
		g.rsys.PushActivePanelName("chat")
	})

	g.rsys.HandleRune('`', func(ev termbox.Event) {
		g.rsys.PushActivePanelName("console")
	})

	// convert to func SetupDirections()
	for k, v := range CARDINALS {
		func(c rune, a string) {
			g.rsys.HandleRune(c, func(_ termbox.Event) {
				// lol collision
				g.nsys.SendPacket("Taction", a)
				/*
					offset := game.DirTable[d]
					g.pm.Lock()
					defer g.pm.Unlock()
					oldposx, oldposy := g.player.GetPos()
					newpos := image.Pt(oldposx+offset.X, oldposy+offset.Y)
					if g.Map.CheckCollision(nil, newpos) {
						g.player.SetPos(newpos.X, newpos.Y)
					}*/
			})

			/*
				      scale := PLAYER_RUN_SPEED
							upperc := unicode.ToUpper(c)
							g.HandleRune(upperc, func(_ termbox.Event) {
								for i := 0; i < scale; i++ {
									g.Player.Move(d)
								}
							})
			*/
		}(k, v)
	}
}

func (g *Game) End() {
	glog.Info("stopping systems")

	g.scene.StopSystems()

	glog.Info("ending")
	glog.Flush()
}

// deal with gnet.Packets received from the server
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
		obj := pk.Data[0].(game.Object)

		name := fmt.Sprintf("%s%d", obj.GetName(), obj.GetID())

		a := g.scene.Add(name)

		pos := g.msys.Pos()
		pos.Set(image.Pt(obj.GetPos()))

		sp := game.NewStaticSprite(name, obj.GetGlyph())

		a.Add(pos)
		a.Add(sp)

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
