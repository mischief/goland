// GameServer: main gameserver struct and functions
// does not know or care about where Packets come from,
// they just arrive on our In port.
package main

import (
	"fmt"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/trustmaster/goflow"
	"image"
	"log"
	"net"
)

var (
	Actions = map[game.Action]func(*GameServer, *ClientPacket){
		game.ACTION_ITEM_PICKUP:         Action_ItemPickup,
		game.ACTION_ITEM_DROP:           Action_ItemDrop,
		game.ACTION_ITEM_LIST_INVENTORY: Action_Inventory,
	}
)

type GameServer struct {
	flow.Graph // graph for our procs; see goflow

	Listener   net.Listener       // acceptor of client connections
	PacketChan chan *ClientPacket // channel where clients packets arrive

	*game.DefaultSubject

	Sessions map[uuid.UUID]*WorldSession //client list

	Objects    game.GameObjectMap
	Map        *game.MapChunk
	Parameters *gutil.LuaParMap
}

func NewGameServer(params *gutil.LuaParMap) *GameServer {

	// flow network setup
	gs := new(GameServer)
	gs.Parameters = params
	gs.InitGraphState()

	// add nodes
	gs.Add(NewPacketRouter(gs), "router")
	gs.Add(new(PacketLogger), "logger")

	// connect processes
	gs.Connect("router", "Log", "logger", "In", make(chan *ClientPacket))

	// map external ports
	gs.MapInPort("In", "router", "In")

	gs.PacketChan = make(chan *ClientPacket, 5)
	gs.SetInPort("In", gs.PacketChan)

	// observers setup
	gs.DefaultSubject = game.NewDefaultSubject()

	// objects setup
	gs.Objects = game.NewGameObjectMap()

	return gs
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
	gs.LoadAssets()

	// setup tcp listener
	log.Printf("GameServer: Starting listener")

	dialstr := ":61507"
	if dialstr, ok := gs.Parameters.Get("listener"); !ok {
		log.Println("GameServer: 'listen' not found in config. defaulting to ", dialstr)
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

func (gs *GameServer) LoadItems() {
	newi := game.NewItem("flag")
	newi.SetTag("visible", true)
	newi.SetTag("gettable", true)
	newi.SetTag("item", true)

	// set the flag's init position
	//newi.SetPos(gs.Map.RandCell()) // maybe later player hater
	newi.SetPos(image.Pt(256/2-5, 256/2+5))

	// add flag to the game
	gs.Objects.Add(newi.GameObject)
}

func (gs *GameServer) LoadAssets() {
	mapfile, ok := gs.Parameters.Get("map")
	if !ok {
		log.Fatal("GameServer: LoadAssets: No map file specified")
	}

	log.Printf("GameServer: LoadAssets: Loading map chunk file: %s", mapfile)
	if gs.Map = game.MapChunkFromFile(mapfile); gs.Map == nil {
		log.Fatal("GameServer: LoadAssets: Can't open map chunk file")
	}
	gs.LoadItems()
}

func (gs *GameServer) SendPacketAll(pk *gnet.Packet) {
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

		//
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
		newplayer.SetPos(image.Pt(256/2, 256/2))

		// set the session's object
		cp.Client.Player = newplayer

		// put player object in world
		gs.Objects.Add(newplayer)

		// tell client about all other objects
		for _, o := range gs.Objects {
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
		gs.Objects.RemoveObject(cp.Client.Player)
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

	for _, o := range gs.Objects {
		// if same pos.. and gettable
		if o.GetPos() == p.GetPos() && o.GetTag("gettable") {
			// pickup item.
			log.Printf("GameServer: Action_ItemPickup: %s picking up %s", p, o)
			o.SetTag("visible", false)
			o.SetTag("gettable", false)
			o.SetPos(image.ZP)
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
	for _, sub := range p.GetSubObjects() {
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

	inv := plobj.GetSubObjects()

	if len(inv) == 0 {
		cp.Reply(gnet.NewPacket("Rchat", "You have 0 items."))
	} else {
		for _, sub := range cp.Client.Player.GetSubObjects() {
			cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You have a %s.", sub.GetName())))
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
	newpos := p.GetPos().Add(game.DirTable[action])
	valid := true

	// check terrain collision
	if !gs.Map.CheckCollision(nil, newpos) {
		valid = false
		cp.Reply(gnet.NewPacket("Rchat", "Ouch! You bump into a wall."))
	}

	// check gameobject collision
	for _, o := range gs.Objects {

		// check if collision with Item and item name is flag
		if o.GetPos() == newpos {
			if o.GetTag("player") {
				cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("Ouch! You bump into %s.", o.GetName())))

				// check if other player's got the goods
				for _, sub := range o.GetSubObjects() {
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
		cp.Client.Player.SetPos(newpos)
	}
}
