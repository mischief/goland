// GameServer: main gameserver struct and functions
// does not know or care about where Packets come from,
// they just arrive on our In port.
package main

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	"github.com/stevedonovan/luar"
	"github.com/trustmaster/goflow"
	"image"
	"log"
	"net"
	"reflect"
)

var (
	Actions = map[game.Action]func(*GameServer, *ClientPacket){
		game.ACTION_ITEM_PICKUP:         Action_ItemPickup,
		game.ACTION_ITEM_DROP:           Action_ItemDrop,
		game.ACTION_ITEM_LIST_INVENTORY: Action_Inventory,
	}

	GS *GameServer
)

type GameServer struct {
	flow.Graph // graph for our procs; see goflow

	Listener   net.Listener       // acceptor of client connections
	PacketChan chan *ClientPacket // channel where clients packets arrive

	*game.DefaultSubject

	Sessions map[int]*WorldSession //client list

	Objects *game.GameObjectMap
	Map     *game.MapChunk

	config *gutil.LuaConfig

	Lua *lua.State
}

func NewGameServer(config *gutil.LuaConfig, ls *lua.State) (*GameServer, error) {
	gs := &GameServer{
		config: config,
	}

	GS = gs

	gs.InitGraphState()

	// add nodes
	gs.Add(NewPacketRouter(gs), "router")
	gs.Add(new(PacketLogger), "logger")

	// connect processes
	gs.Connect("router", "Log", "logger", "In")

	// map external ports
	gs.MapInPort("In", "router", "In")

	gs.PacketChan = make(chan *ClientPacket, 5)
	gs.SetInPort("In", gs.PacketChan)

	// observers setup
	gs.DefaultSubject = game.NewDefaultSubject()

	// objects setup
	gs.Objects = game.NewGameObjectMap()

	// lua state
	gs.Lua = ls

	return gs, nil
}

func (gs *GameServer) Debug() bool {
	if debug, err := gs.config.Get("debug", reflect.Bool); err != nil {
		log.Println("GameServer: 'debug' not found in config. defaulting to false")
		return false
	} else {
		return debug.(bool)
	}
}

func (gs *GameServer) Run() {
	gs.Start()

	for {
		conn, err := gs.Listener.Accept()
		if err != nil {
			log.Println("GameServer: acceptor: ", err)
			continue
		}

		ws := NewWorldSession(gs, conn)
		gs.Attach(ws)

		log.Printf("GameServer: New connection from %s", ws.Con.RemoteAddr())

		go ws.ReceiveProc()
	}

	gs.End()
}

func (gs *GameServer) Start() {
	var err error

	// load assets
	log.Print("GameServer: Loading assets")
	if gs.LoadAssets() != true {
		log.Printf("GameServer: LoadAssets failed")
		return
	}

	// setup tcp listener
	log.Printf("GameServer: Starting listener")

	var dialstr string
	defaultdialstr := ":61507"
	if dialconf, err := gs.config.Get("listener", reflect.String); err != nil {
		log.Println("GameServer: 'listen' not found in config. defaulting to ", defaultdialstr)
		dialstr = defaultdialstr
	} else {
		dialstr = dialconf.(string)
	}

	if gs.Listener, err = net.Listen("tcp", dialstr); err != nil {
		log.Fatalf("GameServer: %s", err)
	}

	// setup goflow network
	log.Print("GameServer: Starting flow")

	flow.RunNet(gs)
}

func (gs *GameServer) End() {
}

func (gs *GameServer) LoadMap(file string) bool {
	if gs.Map = game.MapChunkFromFile(file); gs.Map == nil {
		log.Printf("GameServer: LoadMap: failed loading %s", file)
		return false
	}

	log.Printf("GameServer: LoadMap: loaded map %s", file)
	return true
}

func (gs *GameServer) AddObject(obj game.Object) {
	log.Printf("Adding object %s", obj)

	// tell clients about new object
	gs.SendPkStrAll("Rnewobject", obj)
	gs.Objects.Add(obj)
}

func (gs *GameServer) LuaLog(fmt string, args ...interface{}) {
	log.Printf("GameServer: Lua: "+fmt, args...)
}

func (gs *GameServer) GetScriptPath() string {
	defaultpath := "../scripts/?.lua"
	if scriptconf, err := gs.config.Get("scriptpath", reflect.String); err != nil {
		log.Printf("GameServer: GetScriptPath defaulting to %s: %s", defaultpath, err)
		return defaultpath
	} else {
		return scriptconf.(string)
	}
}

// TODO: move these bindings into another file
func (gs *GameServer) BindLua() {
	luar.Register(gs.Lua, "", luar.Map{
		"gs": gs,
	})

	// add our script path here..
	pkgpathscript := `package.path = package.path .. ";" .. gs.GetScriptPath() --";../?.lua"`
	if err := gs.Lua.DoString(pkgpathscript); err != nil {
	}

	Lua_OpenObjectLib(gs.Lua)
}

// load everything from lua scripts
func (gs *GameServer) LoadAssets() bool {
	gs.BindLua()

	if err := gs.Lua.DoString("require('system')"); err != nil {
		log.Printf("GameServer: LoadAssets: %s", err)
		return false
	}

	return true
}

func (gs *GameServer) SendPkStrAll(tag string, data interface{}) {
	gs.SendPacketAll(gnet.NewPacket(tag, data))
}

// send a packet to all clients
func (gs *GameServer) SendPacketAll(pk *gnet.Packet) {
	gs.DefaultSubject.Lock()
	defer gs.DefaultSubject.Unlock()
	for s := gs.DefaultSubject.Observers.Front(); s != nil; s = s.Next() {
		s.Value.(*WorldSession).SendPacket(pk)
	}
}

func (gs *GameServer) HandlePacket(cp *ClientPacket) {

	switch cp.Tag {

	// Tchat: chat message from a client
	case "Tchat":
		// broadcast chat
		chatline := cp.Data.(string)
		gs.SendPacketAll(gnet.NewPacket("Rchat", fmt.Sprintf("[chat] %s: %s", cp.Client.Username, chatline)))

		// Taction: movement request
	case "Taction":
		gs.HandleActionPacket(cp)

		// Tconnect: user establishes new connection
	case "Tconnect":
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

	case "Tdisconnect":
		// notify clients this player went away
		Action_ItemDrop(gs, cp)
		gs.Objects.RemoveObject(cp.Client.Player)
		gs.Detach(cp.Client)
		gs.SendPacketAll(gnet.NewPacket("Rdelobject", cp.Client.Player))

	case "Tgetplayer":
		if cp.Client.Player != nil {
			cp.Reply(gnet.NewPacket("Rgetplayer", cp.Client.Player.GetID()))
		} else {
			cp.Reply(gnet.NewPacket("Rerror", "nil Player in WorldSession"))
		}

	case "Tloadmap":
		cp.Reply(gnet.NewPacket("Rloadmap", gs.Map))

	default:
		log.Printf("GameServer: HandlePacket: unknown packet type %s", cp.Tag)
	}
}

// Prevent User from re-adding / picking up item
// Disassociate item with map after action successful
func Action_ItemPickup(gs *GameServer, cp *ClientPacket) {
	p := cp.Client.Player

	// we assume our cp.Data is a game.Action of type ACTION_ITEM_PICKUP
	// act accordingly

	for o := range gs.Objects.Chan() {
		// if same pos.. and gettable
		if game.SamePos(o, p) && o.GetTag("gettable") {
			// pickup item.
			log.Printf("GameServer: Action_ItemPickup: %s picking up %s", p, o)
			o.SetTag("visible", false)
			o.SetTag("gettable", false)
			o.SetPos(0, 0)
			p.AddSubObject(o)

			// update clients with the new state of this object
			gs.SendPacketAll(gnet.NewPacket("Raction", o))
			cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You pick up a %s.", o.GetName())))
		}
	}
}

// Player drops the item indicated by the ID from their inventory
// TODO: this drops all items right now. make it drop individual items
func Action_ItemDrop(gs *GameServer, cp *ClientPacket) {
	p := cp.Client.Player
	for sub := range p.GetSubObjects().Chan() {
		log.Printf("GameServer: Action_ItemDrop: %s dropping %s", p, sub)

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
}

// List items in Player's inventory
func Action_Inventory(gs *GameServer, cp *ClientPacket) {
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

}

// Top level handler for Taction packets
func (gs *GameServer) HandleActionPacket(cp *ClientPacket) {
	action := cp.Data.(game.Action)
	p := cp.Client.Player

	_, isdir := game.DirTable[action]
	if isdir {
		gs.HandleMovementPacket(cp)
	}

	// check if this action is in our Actions table, if so execute it
	if f, ok := Actions[action]; ok {
		f(gs, cp)
	}

	gs.SendPacketAll(gnet.NewPacket("Raction", p))
}

// Handle Directionals
func (gs *GameServer) HandleMovementPacket(cp *ClientPacket) {
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
			collfn := luar.NewLuaObjectFromName(gs.Lua, "collide")
			res, err := collfn.Call(p, o)
			if err != nil {
				log.Printf("GameServer: HandleMovementPacket: Lua error: %s", err)
				return
			}

			// only update position if collide returns true
			if thebool, ok := res.(bool); !ok || !thebool {
				log.Printf("GameServer: HandleMovementPacket: Lua collision failed")
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

}
