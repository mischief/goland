package main

import (
	"github.com/golang/glog"
	"github.com/mischief/gochanio"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"net"
	"sync/atomic"
	"time"
)

type ClientNetworkSystem struct {
	do      chan func(*ClientNetworkSystem)
	running int32
	g       *Game
	scene   *game.Scene

	server      string             // server dial string
	servercon   net.Conn           // server tcp connection
	serverrchan <-chan interface{} // server read channel
	serverwchan chan<- interface{} // server write channel
}

func NewClientNetworkSystem(g *Game, s *game.Scene, server string) (*ClientNetworkSystem, error) {
	sys := &ClientNetworkSystem{
		do:     make(chan func(*ClientNetworkSystem), 10),
		g:      g,
		scene:  s,
		server: server,
	}

	if err := game.StartSystem(sys, true); err != nil {
		return nil, err
	}
	sys.scene.AddSystem(sys)
	sys.Syn()
	return sys, nil
}

func (sys ClientNetworkSystem) String() string {
  return "clientnetwork"
}

func (sys *ClientNetworkSystem) DoOne() error {
	f := <-sys.do
	f(sys)

	return nil
}

func (sys *ClientNetworkSystem) Syn() {
	ack := make(chan bool)
	sys.do <- func(sys *ClientNetworkSystem) {
		ack <- true
	}
	<-ack
	close(ack)
}

func (sys *ClientNetworkSystem) Running(r bool) {
	if r {
		atomic.CompareAndSwapInt32(&sys.running, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&sys.running, 1, 0)
	}
}

func (sys *ClientNetworkSystem) IsRunning() bool {
	if atomic.LoadInt32(&sys.running) == 1 {
		return true
	}

	return false
}

func (sys *ClientNetworkSystem) Stop() {
	if sys.IsRunning() {
		sys.do <- func(sys *ClientNetworkSystem) {
			sys.Running(false)
		}
	}
}

func (sys *ClientNetworkSystem) Frequency() time.Duration {
	return 100 * time.Millisecond
}

func (sys *ClientNetworkSystem) Setup() error {
	glog.Info("setup: begin")

	con, err := net.Dial("tcp", sys.server)
	if err != nil {
		glog.Info(err)
		return err
	} else if glog.V(1) {
		glog.Infof("connected to %s", sys.server)
	}

	sys.servercon = con
	sys.serverrchan = chanio.NewReader(sys.servercon)
	sys.serverwchan = chanio.NewWriter(sys.servercon)

  sys.g.em.On("packet", func(i ...interface{}) {
    pkt := i[0].(*gnet.Packet)
    sys.onpacket(pkt)
  })

	glog.Info("setup: complete")
	return nil
}

func (sys *ClientNetworkSystem) Tick(timestep time.Duration) {
	ts := timestep
	sys.do <- func(sys *ClientNetworkSystem) {
		sys.Update(ts)
	}
}

func (sys *ClientNetworkSystem) TearDown() {
	// TODO: cleanup networking
	glog.Info("teardown: complete")
	sys.scene.RemoveSystem(sys)
}

func (sys *ClientNetworkSystem) Update(delta time.Duration) {
	// TODO: send/recieve packets
loop:
	for {
		select {
		case i := <-sys.serverrchan:
			pkt, ok := i.(*gnet.Packet)
			if !ok {
				glog.Warning("update: bogus packet %#v from %s", i, sys.server)
			}

			if glog.V(1) {
				glog.Infof("update: got packet %s", pkt)
			}

      sys.g.em.Emit("packet", pkt)

		default:
			break loop
		}
	}
}

func (sys *ClientNetworkSystem) SendPacket(tag string, data interface{}) {
	sys.do <- func(sys *ClientNetworkSystem) {
		pkt := gnet.NewPacket(tag, data)
		glog.Infof("sendpacket: sending %s", pkt)
		sys.serverwchan <- pkt
	}
}

func (sys *ClientNetworkSystem) onpacket(pkt *gnet.Packet) {
    switch pkt.Tag {
    case "Rgetterrain":
      mapname := pkt.Data[0].(string)
      ter := pkt.Data[1].(*game.TerrainChunk)
      sys.g.rsys.SetTerrainChunk(ter)
      glog.Infof("loaded map %s", mapname)
    case "newactor":
      id := pkt.Data[0].(string)
      glog.Info("adding actor ", id)
      _ = sys.scene.Add(id)
    case "propposadd":
      idpos := pkt.Data[0].(*gnet.PropPosAdd)
      pos := sys.g.msys.Pos()
      pos.Set(idpos.Pos)
      sys.scene.Actors[idpos.Id].Add(pos)
    case "propspriteadd":
      idsprite := pkt.Data[0].(*gnet.PropSpriteAdd)
      sprite := game.NewStaticSprite(idsprite.Id, idsprite.Sprite)
      sys.scene.Actors[idsprite.Id].Add(sprite)
    }
}
