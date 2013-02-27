// GameServer: main gameserver struct and functions
// does not know or care about where Packets come from,
// they just arrive on our In port.
package main

import (
	"github.com/mischief/goland/game/gnet"
	"github.com/trustmaster/goflow"
)

type GameServer struct {
	flow.Graph // graph for our procs; see goflow
}

func NewGameServer() *GameServer {
	gs := new(GameServer)
	gs.InitGraphState()

	// add nodes
	gs.Add(new(PacketRouter), "router")
	gs.Add(new(PacketLogger), "logger")

	// connect processes
	gs.Connect("router", "Log", "logger", "In", make(chan *gnet.Packet))

	// map external ports
	gs.MapInPort("In", "router", "In")

	return gs
}
