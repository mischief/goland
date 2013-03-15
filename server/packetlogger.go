// PacketLogger: flow proc to log Packets
package main

import (
	"github.com/trustmaster/goflow"
	"log"
)

type PacketLogger struct {
	flow.Component
	In <-chan *ClientPacket // channel of packets to log
}

func (l *PacketLogger) OnIn(p *ClientPacket) {
	log.Printf("PacketLogger: OnIn: %s", p)
}
