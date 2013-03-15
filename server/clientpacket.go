package main

import (
	"fmt"
	"github.com/mischief/goland/game/gnet"
)

type ClientPacket struct {
	Client *WorldSession
	*gnet.Packet
}

func (cp ClientPacket) String() string {
	return fmt.Sprintf("%s %s", cp.Client.Con.RemoteAddr(), cp.Packet)
}

func (cp *ClientPacket) Reply(pk *gnet.Packet) {
	cp.Client.ClientWChan <- pk
}
