// PacketLogger: flow proc to log Packets
package main

import (
	"github.com/mischief/goland/game/gnet"
	"github.com/trustmaster/goflow"
	"log"
)

type PacketLogger struct {
	flow.Component
	In <-chan *gnet.Packet // channel of packets to log
}

func (l *PacketLogger) OnIn(p *gnet.Packet) {
	log.Printf("%#v", p)
}
