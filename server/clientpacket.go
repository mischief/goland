package main

import (
	"fmt"
	"github.com/mischief/goland/game/gnet"
	"log"
)

type ClientPacket struct {
	Client *WorldSession
	*gnet.Packet
}

func (cp ClientPacket) String() string {
	return fmt.Sprintf("%s %s", cp.Client.Con.RemoteAddr(), cp.Packet)
}

func (cp *ClientPacket) Reply(pk *gnet.Packet) {
	log.Printf("ClientPacket: Reply: %s -> %s %s", cp.Packet, cp.Client.Con.RemoteAddr(), pk)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("ClientPacket: Reply: error: %s", err)
		}
	}()

	cp.Client.ClientWChan <- pk
}
