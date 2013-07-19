package main

import (
	"fmt"
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

const (
	FPS_LIMIT = 23
)

var (
	CARDINALS = map[rune]game.Action{
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
	}
)

type Game struct {
	scene *game.Scene

	rsys *graphics.RenderSystem
	msys *game.MovementSystem
	nsys *ClientNetworkSystem

	closechan chan bool

	//Objects *game.GameObjectMap
	//Map     *game.MapChunk

	config *gutil.LuaConfig

	ServerCon net.Conn

	ServerRChan <-chan interface{}
	ServerWChan chan<- interface{}
}

func NewGame(config *gutil.LuaConfig) *Game {
	g := Game{
		scene:     game.NewScene(),
		config:    config,
		closechan: make(chan bool, 1),
	}

	// make player
	/*
	  pl := g.scene.Add("player")

	  pos := g.msys.Pos()
	  pos.Set(image.Pt(0,0))

	  sp := g.rsys.StaticSprite("Human", termbox.Cell{'@', 0, 0})

	  pl.Add(pos)
	  pl.Add(sp)
	*/

	return &g
}

func (g *Game) Run() {

	g.Start()

	timer := game.NewDeltaTimer()
	ticker := time.NewTicker(time.Second / FPS_LIMIT)

	run := true

	for run {
		select {
		case <-ticker.C:
			// frame tick
			delta := timer.DeltaTime()

			if delta.Seconds() > 0.25 {
				delta = time.Duration(250 * time.Millisecond)
			}

			g.Update(delta)
			g.Draw()

			glog.Flush()

			//g.Flush()

		case <-g.closechan:
			glog.Infof("got close signal")
			run = false
		}
	}

	g.End()

}

func (g *Game) Start() {
	glog.Info("starting")

	// config items
	// network setup
	sc, err := g.config.Get("server", reflect.String)
	if err != nil {
		glog.Fatal("start: missing server in config: %s", err)
	}

	server := sc.(string)

	// systems
	if g.rsys, err = graphics.NewRenderSystem(g.scene); err != nil {
		glog.Fatalf("rendersystem: %s", err)
	}

	if g.msys, err = game.NewMovementSystem(g.scene); err != nil {
		glog.Fatalf("movementsystem: %s", err)
	}

	if g.nsys, err = NewClientNetworkSystem(g.scene, server); err != nil {
		glog.Fatalf("clientnetworksystem: %s", err)
	}

	// camera
	c := g.scene.Add("camera")
	cpp := g.msys.Pos()
	cpp.Set(image.Pt(128, 128))
	c.Add(cpp)
	cc := g.rsys.Cam(c.ID)
	cc.SetCenter(<-cpp.Get()) // is this necessary?
	c.Add(cc)

	// graphics panels
	g.rsys.AddPanel("stats", NewStatsPanel())
	g.rsys.AddPanel("view", graphics.NewViewPanel(g.rsys, c))
	g.rsys.AddPanel("log", NewLogPanel())
	g.rsys.AddPanel("player", NewPlayerPanel(g))
	g.rsys.AddPanel("chat", NewChatPanel(g, g.nsys))

	// login
	username, err := g.config.Get("username", reflect.String)
	if err != nil {
		glog.Fatal("start: missing username in config: %s", err)
	}

	g.nsys.SendPacket("Tconnect", username)

	// request the map from server
	g.nsys.SendPacket("Tloadmap", nil)

	// request the object we control
	// XXX: the delay is to fix a bug regarding ordering of packets.
	// if the client gets the response to this before he is notified
	// that the object exists, it will barf, so we delay this request.
	time.AfterFunc(50*time.Millisecond, func() {
		g.nsys.SendPacket("Tgetplayer", nil)
	})

	// anonymous function that reads packets from the server
	/*
		go func(r <-chan interface{}) {
			for x := range r {
				p, ok := x.(*gnet.Packet)
				if !ok {
					glog.Warningf("server read: bogus server packet %#v", x)
					continue
				}

				g.HandlePacket(p)
			}
			glog.Warning("server read: Disconnected from server!")
			//io.WriteString(g.logpanel, "Disconnected from server!")
		}(g.ServerRChan)
	*/

	// ESC to quit
	g.rsys.HandleKey(termbox.KeyEsc, func(ev termbox.Event) { g.closechan <- false })

	// Enter to chat
	g.rsys.HandleKey(termbox.KeyEnter, func(ev termbox.Event) {
		g.rsys.PushPanelInputName("chat")
	})

	// convert to func SetupDirections()
	for k, v := range CARDINALS {
		func(c rune, d game.Action) {
			g.rsys.HandleRune(c, func(_ termbox.Event) {
				// lol collision
				g.nsys.SendPacket("Taction", CARDINALS[c])
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

	g.rsys.Stop()
	g.msys.Stop()
	g.nsys.Stop()
	g.scene.Wg.Wait()

	glog.Info("ending")
	glog.Flush()
}

func (g *Game) Update(delta time.Duration) {
	// collect stats
	/*
		for _, p := range g.panels {
			if v, ok := p.(InputHandler); ok {
				v.HandleInput(termbox.Event{Type: termbox.EventResize})
			}

			if v, ok := p.(gutil.Updater); ok {
				v.Update(delta)
			}
		}

		//g.RunInputHandlers()

		//for o := range g.Objects.Chan() {
		//	o.Update(delta)
		//}
	*/

}

func (g *Game) Draw() {

	//g.Terminal.Clear()
	//g.mainpanel.Clear()

	// draw panels
	/*
		for _, p := range g.panels {
			if v, ok := p.(panel.Drawer); ok {
				v.Draw()
			}
		}
	*/

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
		obj := pk.Data.(game.Object)

		name := fmt.Sprintf("%s%d", obj.GetName(), obj.GetID())

		a := g.scene.Add(name)

		pos := g.msys.Pos()
		pos.Set(image.Pt(obj.GetPos()))

		sp := g.rsys.StaticSprite(name, obj.GetGlyph())

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

		// Rloadmap: get the map data from the server
	case "Rloadmap":
		//gmap := pk.Data.(*game.MapChunk)
		//g.Map = gmap

	default:
		glog.Infof("bad packet tag %s", pk.Tag)
	}

}
