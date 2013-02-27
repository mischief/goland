// PacketRouter: proc to handle actions for different known Packet tags
package main

import (
	"github.com/mischief/goland/game/gnet"
	"github.com/trustmaster/goflow"
	"image"
	"log"
)

type PacketRouter struct {
	flow.Component
	In  <-chan *gnet.Packet
	Log chan<- *gnet.Packet
}

func (pf *PacketRouter) OnIn(p *gnet.Packet) {
	// log all packets
	pf.Log <- p

	switch p.Tag {
	case "move":
		v := p.Data.(*image.Point)
		log.Printf("PacketRouter: %s %s", p.Tag, v)
	default:
		log.Printf("PacketRouter: unknown packet type %s", p.Tag)
	}
}
