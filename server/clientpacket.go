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
	return fmt.Sprintf("%s %s", cp.Client.con.RemoteAddr(), cp.Packet)
}

func (cp *ClientPacket) Reply(tag string, data... interface{}) {
  pk := gnet.NewPacket(tag, data...)

	if glog.V(2) {
		glog.Infof("reply: %s -> %s %s", cp.Packet, cp.Client.con.RemoteAddr(), pk)
	}

	defer func() {
		if err := recover(); err != nil {
			glog.Error("reply: error: ", err)
		}
	}()

	cp.Client.clientwchan <- pk
}
