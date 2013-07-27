package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/mischief/gochanio"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	uuid "github.com/nu7hatch/gouuid"
	"image"
	"net"
)

type WorldSession struct {
	con         net.Conn           // connection to the client
	clientrchan <-chan interface{} // client's incoming message channel
	clientwchan chan<- interface{} // client's outgoing message channel
	ID          uuid.UUID          // client's id (account id?)
	Username    string             // associated username
	Pos         image.Point        // XXX: what's this for?
	Player      *game.Actor         // object this client controls
	nsys        *ServerNetworkSystem
	gs          *GameServer // world reference
}

func (ws *WorldSession) String() string {
	return fmt.Sprintf("addr %s user %s", ws.con.RemoteAddr(), ws.Username)
}

func NewWorldSession(gs *GameServer, nsys *ServerNetworkSystem, c net.Conn) *WorldSession {
	var err error
	var id *uuid.UUID

	n := &WorldSession{
		con:      c,
		Username: "(unknown)",
		nsys:     nsys,
		gs:       gs,
	}

	n.clientrchan = chanio.NewReader(n.con)
	n.clientwchan = chanio.NewWriter(n.con)

	if id, err = uuid.NewV4(); err != nil {
		glog.Error("uuid.NewV4: ", err)
		return nil
	} else {
		n.ID = *id
	}

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
    id := i[0].(string)
    sprite := i[1].(*game.StaticSprite)
    n.SendPacket(gnet.NewPacket("propspriteadd", gnet.PropSpriteAdd{id, sprite.GetCell()}))
  })

	return n
}

// handle per-client packets
func (ws *WorldSession) ReceiveProc() {
	defer func() {
		if err := recover(); err != nil {
			glog.Warning("receiveproc: ", err)
		}
	}()

	for x := range ws.clientrchan {
		p, ok := x.(*gnet.Packet)
		if !ok {
			glog.Warning("receiveproc: bogus packet %#v from %s", x, ws.con.RemoteAddr())
			continue
		}

		// TODO: handle this client's packets
		ws.nsys.HandlePacket(&ClientPacket{ws, p})
	}

	ws.gs.em.Emit("disconnect", ws)

	glog.Infof("receiveproc: channel closed %s", ws)
}

// send packet to this client
func (ws *WorldSession) SendPacket(pk *gnet.Packet) {
	if glog.V(2) {
		glog.Infof("sendpacket: %s %s", ws.con.RemoteAddr(), pk)
	}

	defer func() {
		if err := recover(); err != nil {
			glog.Error("sendpacket: error: ", err)
		}
	}()
	ws.clientwchan <- pk
}

func (ws *WorldSession) Update() {
}
