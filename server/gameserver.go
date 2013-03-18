// GameServer: main gameserver struct and functions
// does not know or care about where Packets come from,
// they just arrive on our In port.
package main

import (
	"fmt"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	"github.com/trustmaster/goflow"
	"image"
	"log"
	"net"
)

var (
	Actions = map[game.Action]func(*GameServer, *ClientPacket){
	        game.ACTION_ITEM_PICKUP: Action_ItemPickup,
	        game.ACTION_ITEM_DROP: Action_ItemDrop,
		game.ACTION_ITEM_LIST_INVENTORY: Action_Inventory,
	}
)

type GameServer struct {
	flow.Graph // graph for our procs; see goflow

	Listener   net.Listener       // acceptor of client connections
	PacketChan chan *ClientPacket // channel where clients packets arrive

	*game.DefaultSubject //

	Sessions map[int64]*WorldSession //client list

	Objects game.GameObjectMap
	//Objects []*game.GameObject
	Map *game.MapChunk

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

		log.Printf("New World Session: %s", ws)

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

func (gs *GameServer) LoadAssets() {
	mapfile, ok := gs.Parameters.Get("map")
	if !ok {
		log.Fatal("GameServer: LoadAssets: No map file specified")
	}

	log.Printf("GameServer: LoadAssets: Loading map chunk file: %s", mapfile)
	if gs.Map = game.MapChunkFromFile(mapfile); gs.Map == nil {
		log.Fatal("GameServer: LoadAssets: Can't open map chunk file")
	}

	newi := game.NewItem("flag")

	// set the position of the flag
	newi.SetPos(gs.Map.RandCell())
	gs.Objects.Add(newi.GameObject)
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
		newpl := game.NewPlayer(username)
		newpl.SetPos(image.Pt(256/2, 256/2))

		cp.Client.Player = newpl

		gs.Objects.Add(newpl.GameObject)

		gs.SendPacketAll(gnet.NewPacket("Rnewobject", newpl.GameObject))

		// tell client about all other objects
		for _, o := range gs.Objects {
			cp.Reply(gnet.NewPacket("Rnewobject", o))
		}

		cp.Reply(gnet.NewPacket("Rchat", "Welcome to Goland!"))

	case "Tdisconnect":
		// notify clients this player went away
		gs.Objects.RemoveObject(cp.Client.Player.GameObject)
		gs.SendPacketAll(gnet.NewPacket("Rdelobject", cp.Client.Player.GameObject))

	case "Tgetplayer":
		if cp.Client.Player != nil {
			cp.Reply(gnet.NewPacket("Rgetplayer", cp.Client.Player.ID))
		} else {
			cp.Reply(gnet.NewPacket("error", "nil Player in WorldSession"))
		}

	case "Tloadmap":
		cp.Reply(gnet.NewPacket("Rloadmap", gs.Map))

	default:
		log.Printf("GameServer: HandlePacket: unknown packet type %s", cp.Tag)
	}
}


func Action_ItemPickup(gs *GameServer, cp *ClientPacket) {
	cp.Reply(gnet.NewPacket("Rchat", "Picking up Item"))
}

func Action_ItemDrop(gs *GameServer, cp *ClientPacket) {
	cp.Reply(gnet.NewPacket("Rchat", "Dropping Item"))
}

func Action_Inventory(gs *GameServer, cp *ClientPacket) {	
	cp.Reply(gnet.NewPacket("Rchat", cp.Client.Player.Inventory.String()))
}

func (gs *GameServer) HandleActionPacket(cp *ClientPacket) {
	action := cp.Data.(game.Action)

	_, isdir := game.DirTable[action]
	if isdir {
		gs.HandleMovementPacket(cp)
	}
	
	Actions[action](gs, cp)
	
	gs.SendPacketAll(gnet.NewPacket("Raction", cp.Client.Player))
}

// Handle Directionals
func (gs *GameServer) HandleMovementPacket(cp *ClientPacket) {
	action := cp.Data.(game.Action)
	newpos := cp.Client.Player.GetPos().Add(game.DirTable[action])
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
			if o.Tags["player"] {
				cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("Ouch! You bump into %s.", o.Name)))
				valid = false
				break
			}

			if o.Tags["item"] && valid {
				cp.Reply(gnet.NewPacket("Rchat", fmt.Sprintf("You see a %s here.", o.Name)))
			}
		}
	}
	
	if valid {
		cp.Client.Player.SetPos(newpos)
	}
}