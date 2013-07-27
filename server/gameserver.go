// GameServer: main gameserver struct and functions
package main

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/golang/glog"
	"github.com/mischief/goland/game"
	//"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	"github.com/stevedonovan/luar"
	//"image"
	"github.com/chuckpreslar/emission"
	"net"
	"os"
	"os/signal"
	"reflect"
)

type GameServer struct {
	scene *game.Scene
	em    *emission.Emitter

	msys *game.MovementSystem
  tsys *game.TerrainSystem
	nsys *ServerNetworkSystem

	closechan chan bool
	sigchan   chan os.Signal

	Listener   net.Listener       // acceptor of client connections
	PacketChan chan *ClientPacket // channel where clients packets arrive

	*game.DefaultSubject

	Sessions map[int]*WorldSession //client list

	Objects *game.GameObjectMap

	config *gutil.LuaConfig

	lua *lua.State
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
		glog.Warning("'debug' not found in config. defaulting to false")
		gs.debug = false
	} else {
		gs.debug = debug.(bool)
	}

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

	var listen string
	defaultlistenstr := "127.0.0.1:61507"
	if listenconf, err := gs.config.Get("listener", reflect.String); err != nil {
		glog.Info("'listener' not found in config. defaulting to ", defaultlistenstr)
		listen = defaultlistenstr
	} else {
		listen = listenconf.(string)
	}

	if gs.msys, err = game.NewMovementSystem(gs.scene); err != nil {
		glog.Fatalf("movementsystem: %s", err)
	}

  if gs.tsys, err = game.NewTerrainSystem(gs.scene); err != nil {
		glog.Fatalf("terrainsystem: %s", err)
  }

	if gs.nsys, err = NewServerNetworkSystem(gs, gs.scene, listen); err != nil {
		glog.Fatalf("servernetworksystem: %s", err)
	}

	// load assets
	glog.Info("loading assets")
	if gs.LoadAssets() != true {
		glog.Error("loading assets failed")
		return
	}

}

func (gs *GameServer) End() {
	glog.Info("stopping systems")

	gs.scene.StopSystems()

	glog.Info("systems stopped")
}

func (gs *GameServer) AddObject(obj game.Object) {
	glog.Infof("adding object %s", obj)

	// tell clients about new object
	gs.nsys.SendPacketAll("Rnewobject", obj)
	gs.Objects.Add(obj)
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
    "nsys": gs.nsys,
  })


	// add our script path here..
	pkgpathscript := `package.path = package.path .. ";" .. gs.GetScriptPath() --";../?.lua"`
	if err := gs.lua.DoString(pkgpathscript); err != nil {
	}

	Lua_OpenObjectLib(gs.lua)
}

// load everything from lua scripts
func (gs *GameServer) LoadAssets() bool {
	gs.BindLua()

	if err := gs.lua.DoString("require('system')"); err != nil {
		glog.Errorf("loadassets: %s", err)
		return false
	}

	return true
}

func (gs *GameServer) HandlePacket(cp *ClientPacket) {

	switch cp.Tag {

	// Tchat: chat message from a client
	case "Tchat":
		// broadcast chat
		chatline := cp.Data[0].(string)
		gs.nsys.SendPacketAll("Rchat", fmt.Sprintf("[chat] %s: %s", cp.Client.Username, chatline))

		// Taction: movement request
	case "Taction":
		gs.HandleActionPacket(cp)

		// Tconnect: user establishes new connection
	case "Tconnect":
		/*
			username, ok := cp.Data.(string)

			if !ok {
				cp.Reply(gnet.NewPacket("Rerror", "invalid username or conversion failed"))
				break
			} else {
				cp.Client.Username = username
			}

			// make new player for client
			var newplayer game.Object
			newplayer = game.NewGameObject(username)
			newplayer.SetTag("player", true)
			newplayer.SetTag("visible", true)

			// setting this lets players pick up other players, lol
			//newplayer.SetTag("gettable", true)
			newplayer.SetGlyph(game.GLYPH_HUMAN)
			newplayer.SetPos(256/2, 256/2)

			// set the session's object
			cp.Client.Player = newplayer

			// put player object in world
			gs.Objects.Add(newplayer)

			// tell client about all other objects
			for o := range gs.Objects.Chan() {
				if o.GetID() != newplayer.GetID() {
					cp.Reply(gnet.NewPacket("Rnewobject", o))
				}
			}

			// tell all clients about the new player
			gs.SendPacketAll(gnet.NewPacket("Rnewobject", newplayer))

			// greet our new player
			cp.Reply(gnet.NewPacket("Rchat", "Welcome to Goland!"))
		*/

	case "Tdisconnect":
		// notify clients this player went away
		/*
			Action_ItemDrop(gs, cp)
			gs.Objects.RemoveObject(cp.Client.Player)
			gs.Detach(cp.Client)
			gs.SendPacketAll(gnet.NewPacket("Rdelobject", cp.Client.Player))
		*/

	case "Tgetplayer":
		/*
			if cp.Client.Player != nil {
				cp.Reply(gnet.NewPacket("Rgetplayer", cp.Client.Player.GetID()))
			} else {
				cp.Reply(gnet.NewPacket("Rerror", "nil Player in WorldSession"))
			}
		*/

	case "Tloadmap":
		//cp.Reply(gnet.NewPacket("Rloadmap", gs.Map))

	default:
		glog.Infof("handlepacket: unknown packet type %s", cp.Tag)
	}
}

// Prevent User from re-adding / picking up item
// Disassociate item with map after action successful
func Action_ItemPickup(gs *GameServer, cp *ClientPacket) {
	/*
		p := cp.Client.Player

		// we assume our cp.Data is a game.Action of type ACTION_ITEM_PICKUP
		// act accordingly

		for o := range gs.Objects.Chan() {
			// if same pos.. and gettable
			if game.SamePos(o, p) && o.GetTag("gettable") {
				// pickup item.
				glog.Infof("Action_ItemPickup: %s picking up %s", p, o)
				o.SetTag("visible", false)
				o.SetTag("gettable", false)
				o.SetPos(0, 0)
				p.AddSubObject(o)

				// update clients with the new state of this object
				gs.SendPacketAll(gnet.NewPacket("Raction", o))
				cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You pick up a %s.", o.GetName())))
			}
		}
	*/
}

// Player drops the item indicated by the ID from their inventory
// TODO: this drops all items right now. make it drop individual items
func Action_ItemDrop(gs *GameServer, cp *ClientPacket) {
	/*
		p := cp.Client.Player
		for sub := range p.GetSubObjects().Chan() {
			glog.Infof("Action_ItemDrop: %s dropping %s", p, sub)

			// remove item from player
			p.RemoveSubObject(sub)
			// put it where the player was
			sub.SetPos(p.GetPos())
			// make it visible
			sub.SetTag("visible", true)
			sub.SetTag("gettable", true)

			// update clients with the new state of this object
			gs.SendPacketAll(gnet.NewPacket("Raction", sub))
			cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You drop a %s.", sub.GetName())))
		}
	*/
}

// List items in Player's inventory
func Action_Inventory(gs *GameServer, cp *ClientPacket) {
	/*
		plobj := cp.Client.Player

		inv := plobj.GetSubObjects().Chan()

		if len(inv) == 0 {
			cp.Reply(gnet.NewPacket("Rchat", "You have 0 items."))
		} else {
			counts := make(map[string]int)
			for sub := range inv {
				n := sub.GetName()
				if _, ok := counts[n]; ok {
					counts[n]++
				} else {
					counts[n] = 1
				}
			}

			for n, c := range counts {
				if c == 1 {
					cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You have a %s.", n)))
				} else {
					cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You have %d %ss.", c, n)))
				}

			}
		}

	*/
}

// Top level handler for Taction packets
func (gs *GameServer) HandleActionPacket(cp *ClientPacket) {
  /*
	action := cp.Data[0].(game.Action)
	p := cp.Client.Player

	_, isdir := game.DirTable[action]
	if isdir {
		gs.HandleMovementPacket(cp)
	}

	// check if this action is in our Actions table, if so execute it
	if f, ok := Actions[action]; ok {
		f(gs, cp)
	}

	gs.nsys.SendPacketAll("Raction", p)
  */
}

// Handle Directionals
func (gs *GameServer) HandleMovementPacket(cp *ClientPacket) {
	/*
		action := cp.Data.(game.Action)
		p := cp.Client.Player
		offset := game.DirTable[action]
		oldposx, oldposy := p.GetPos()
		newpos := image.Pt(oldposx+offset.X, oldposy+offset.Y)
		valid := true

		// check terrain collision
		if !gs.Map.CheckCollision(nil, newpos) {
			valid = false
			cp.Reply(gnet.NewPacket("Rchat", "Ouch! You bump into a wall."))
		}

		// check gameobject collision
		for o := range gs.Objects.Chan() {

			// check if collision with Item and item name is flag
			px, py := o.GetPos()
			if px == newpos.X && py == newpos.Y {
				collfn := luar.NewLuaObjectFromName(gs.lua, "collide")
				res, err := collfn.Call(p, o)
				if err != nil {
					glog.Infof("GameServer: HandleMovementPacket: Lua error: %s", err)
					return
				}

				// only update position if collide returns true
				if thebool, ok := res.(bool); !ok || !thebool {
					glog.Infof("GameServer: HandleMovementPacket: Lua collision failed")
					valid = false
				} else {
					// tell everyone that the colliders changed
					gs.SendPacketAll(gnet.NewPacket("Raction", o))
				}

				if o.GetTag("player") {
					cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("Ouch! You bump into %s.", o.GetName())))

					// check if other player's got the goods
					for sub := range o.GetSubObjects().Chan() {
						if sub.GetTag("item") == true {
							// swap pop'n'lock

							// remove item from player
							swap := o.RemoveSubObject(sub)
							p.AddSubObject(swap)
							cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You steal a %s!", swap.GetName())))
						}
					}
				}

				if o.GetTag("item") && o.GetTag("gettable") && valid {
					cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You see a %s here.", o.GetName())))
				}
			}
		}

		if valid {
			cp.Client.Player.SetPos(newpos.X, newpos.Y)
			//gs.SendPacketAll(gnet.NewPacket("Raction", p))
		}
	*/

}
