package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/mischief/gochanio"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gobj"
	"net"
	"sync/atomic"
)

type WorldSession struct {
	con         net.Conn             // connection to the client
	clientrchan <-chan interface{}   // client's incoming message channel
	clientwchan chan<- interface{}   // client's outgoing message channel
	werr        <-chan error         // write error
	rerr        <-chan error         // read error
	Username    string               // associated username
	Player      gobj.Object          // object this client controls
	nsys        *ServerNetworkSystem // server's network system
	gs          *GameServer          // world reference

	alive int32
}

func (ws *WorldSession) String() string {
	return fmt.Sprintf("addr %s user %s", ws.con.RemoteAddr(), ws.Username)
}

func NewWorldSession(gs *GameServer, nsys *ServerNetworkSystem, c net.Conn) *WorldSession {
	n := &WorldSession{
		con:      c,
		Username: "(unknown)",
		nsys:     nsys,
		gs:       gs,
	}

	atomic.StoreInt32(&n.alive, 1)

	n.clientrchan, n.rerr = chanio.NewReader(n.con)
	n.clientwchan, n.werr = chanio.NewWriter(n.con)

	n.gs.em.On("newactor", func(i ...interface{}) {
		actor := i[0].(*game.Actor)
		n.SendPacket(gnet.NewPacket("newactor", actor.ID))
	})
	n.gs.em.On("propposadd", func(i ...interface{}) {
		id := i[0].(string)
		pos := i[1].(*game.Pos)
		n.SendPacket(gnet.NewPacket("propposadd", gnet.PropPosAdd{id, <-pos.Get()}))
	})
	n.gs.em.On("propspriteadd", func(i ...interface{}) {
		/*
			id := i[0].(string)
			sprite := i[1].(*game.StaticSprite)
			n.SendPacket(gnet.NewPacket("propspriteadd", gnet.PropSpriteAdd{id, sprite.GetCell()}))
		*/
	})

	return n
}

// handle per-client packets
func (ws *WorldSession) ReceiveProc() {
	/*
		defer func() {
			if err := recover(); err != nil {
				glog.Warning("receiveproc: ", err)
			}
		}()
	*/
	var err error

	for atomic.LoadInt32(&ws.alive) == 1 && err == nil {
		select {
		case err = <-ws.rerr:
			// we're done for
			ws.gs.em.Emit("disconnect", ws)

			atomic.StoreInt32(&ws.alive, 0)
			close(ws.clientwchan)

			glog.Infof("receiveproc: channel closed %s", ws)

		case payload := <-ws.clientrchan:
			packet, ok := payload.(*gnet.Packet)

			if !ok {
				glog.Warning("receiveproc: bogus packet %#v from %s", payload, ws.con.RemoteAddr())
				continue
			}

			ws.nsys.HandlePacket(&ClientPacket{ws, packet})
		}
	}

}

// send packet to this client
func (ws *WorldSession) SendPacket(pk *gnet.Packet) error {
	if atomic.LoadInt32(&ws.alive) == 1 {

		if glog.V(2) {
			glog.Infof("sendpacket: %s %s", ws.con.RemoteAddr(), pk)
		}

		defer func() {
			if err := recover(); err != nil {
				glog.Error("sendpacket: error: ", err)
			}
		}()

		select {
		case e := <-ws.werr:
			return e
		default:
		}

		ws.clientwchan <- pk

		select {
		case e := <-ws.werr:
			return e
		default:
		}
	}

	return fmt.Errorf("not running")
}

func (ws *WorldSession) Update() {
}
