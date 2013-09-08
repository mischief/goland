// PacketRouter: proc to handle actions for different known Packet tags
package main

import (
	"github.com/trustmaster/goflow"

//	"log"
)

type PacketRouter struct {
	flow.Component
	In  <-chan *ClientPacket
	Log chan<- *ClientPacket

	world *GameServer
}

func NewPacketRouter(w *GameServer) *PacketRouter {
	pr := new(PacketRouter)

	pr.world = w

	return pr
}

func (pr *PacketRouter) OnIn(p *ClientPacket) {

	defer func() {
		//if err := recover(); err != nil {
		//	log.Println("PacketRouter: OnIn: panic:", err)
		//}
	}()

	// log all packets
	pr.Log <- p

	pr.world.HandlePacket(p)

	/*
		switch p.Tag {
		case "move":
			v := p.Data.(game.Direction)
			log.Printf("PacketRouter: OnIn: %s %s %s", p.Client.Con.RemoteAddr(), p.Tag, v)
		default:
			log.Printf("PacketRouter: OnIn: unknown packet type %s", p.Tag)
		}
	*/
}
