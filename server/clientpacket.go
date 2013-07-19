package main

import (
	"fmt"
	"github.com/golang/glog"
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
	if glog.V(2) {
		glog.Infof("reply: %s -> %s %s", cp.Packet, cp.Client.Con.RemoteAddr(), pk)
	}

	defer func() {
		if err := recover(); err != nil {
			glog.Error("reply: error: ", err)
		}
	}()

	cp.Client.ClientWChan <- pk
}
