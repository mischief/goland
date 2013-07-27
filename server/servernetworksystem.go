package main

import (
	"github.com/golang/glog"
	//"github.com/mischief/gochanio"
	"fmt"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	termbox "github.com/nsf/termbox-go"
	"image"
	"net"
	"sync/atomic"
	"time"
)

type ServerNetworkSystem struct {
	do      chan func(*ServerNetworkSystem)
	running int32
	scene   *game.Scene
	gs      *GameServer

	listen      string             // server listen string
	listencon   net.Listener       // server tcp connection
	serverrchan <-chan interface{} // server read channel
	serverwchan chan<- interface{} // server write channel

	clients map[string]*WorldSession
}

func NewServerNetworkSystem(gs *GameServer, s *game.Scene, listen string) (*ServerNetworkSystem, error) {
	sys := &ServerNetworkSystem{
		do:      make(chan func(*ServerNetworkSystem), 10),
		scene:   s,
		gs:      gs,
		listen:  listen,
		clients: make(map[string]*WorldSession),
	}

	if err := game.StartSystem(sys, true); err != nil {
		return nil, err
	}

	sys.scene.AddSystem(sys)

	sys.Syn()
	return sys, nil
}

func (sys ServerNetworkSystem) String() string {
	return "servernetwork"
}

func (sys *ServerNetworkSystem) DoOne() error {
	f := <-sys.do
	f(sys)

	return nil
}

func (sys *ServerNetworkSystem) Syn() {
	ack := make(chan bool)
	sys.do <- func(sys *ServerNetworkSystem) {
		ack <- true
	}
	<-ack
	close(ack)
}

func (sys *ServerNetworkSystem) Running(r bool) {
	if r {
		atomic.CompareAndSwapInt32(&sys.running, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&sys.running, 1, 0)
	}
}

func (sys *ServerNetworkSystem) IsRunning() bool {
	if atomic.LoadInt32(&sys.running) == 1 {
		return true
	}

	return false
}

func (sys *ServerNetworkSystem) Stop() {
	if sys.IsRunning() {
		sys.do <- func(sys *ServerNetworkSystem) {
			sys.Running(false)
		}
	}
}

func (sys *ServerNetworkSystem) Frequency() time.Duration {
	return 100 * time.Millisecond
}

func (sys *ServerNetworkSystem) Setup() error {
	glog.Info("setup: begin")

	if listener, err := net.Listen("tcp", sys.listen); err != nil {
		glog.Error(err)
		return err
	} else {
		sys.listencon = listener
		if glog.V(1) {
			glog.Infof("listening on %s", listener.Addr())
		}
	}

	go sys.listenproc()

	sys.gs.em.On("disconnect", func(i ...interface{}) {
		sys.do <- func(sys *ServerNetworkSystem) {
			ws := i[0].(*WorldSession)
			glog.Infof("removing client %s", ws)
			delete(sys.clients, ws.con.RemoteAddr().String())
			glog.Flush()
		}
	})

	//sys.serverrchan = chanio.NewReader(sys.servercon)
	//sys.serverwchan = chanio.NewWriter(sys.servercon)

	glog.Info("setup: complete")
	return nil
}

func (sys *ServerNetworkSystem) Tick(timestep time.Duration) {
	ts := timestep
	sys.do <- func(sys *ServerNetworkSystem) {
		sys.Update(ts)
	}
}

func (sys *ServerNetworkSystem) TearDown() {
	// TODO: cleanup networking
	glog.Info("teardown: complete")
	sys.scene.RemoveSystem(sys)
}

func (sys *ServerNetworkSystem) Update(delta time.Duration) {
	// TODO: send/recieve packets
}

func (sys *ServerNetworkSystem) SendPacketAll(tag string, data interface{}) {
	pkt := gnet.NewPacket(tag, data)
	sys.do <- func(sys *ServerNetworkSystem) {
		for _, ws := range sys.clients {
			ws.SendPacket(pkt)
		}
	}
}

func (sys *ServerNetworkSystem) HandlePacket(pkt *ClientPacket) {
	sys.do <- func(sys *ServerNetworkSystem) {
		sys.dispatchpacket(pkt)
	}
}

// TODO: handle packets
func (sys *ServerNetworkSystem) dispatchpacket(pkt *ClientPacket) {
	if glog.V(2) {
		glog.Infof("processing packet %s", pkt)
	}

	sys.gs.em.Emit("packet", pkt.Data)

	switch pkt.Tag {
	case "Tlogin":
		if user, ok := pkt.Data[0].(string); !ok {
			glog.Error("client did not send string in Tconnect")
			// TODO: disconnect the client
		} else {
			pkt.Client.Username = user

			sys.gs.em.Emit("login", pkt.Client)

			id := fmt.Sprintf("%s%d", user, game.IDGen())

			a := sys.scene.Add(id)
			pkt.Client.Player = a
			sys.gs.em.Emit("newactor", a)

			pos := sys.gs.msys.Pos()
			//pos.Set(image.Pt(256/2, 256/2))
			pos.Set(image.Pt(0, 0))
			a.Add(pos)
			sys.gs.em.Emit("propposadd", a.ID, pos)

			sp := game.NewStaticSprite(id, termbox.Cell{'@', 0, 0})
			a.Add(sp)
			sys.gs.em.Emit("propspriteadd", a.ID, sp)
		}

		// text chat message
	case "Tchat":
		line, ok := pkt.Data[0].(string)
		if ok {
			sys.SendPacketAll("Rchat", fmt.Sprintf("[chat] %s: %s", pkt.Client.Username, line))
		}

		// event
	case "Taction":
	case "Tdisconnect":
	case "Tgetplayer":
		pkt.Reply("Rgetplayer", pkt.Client.Player.ID)

	case "Tgetterrain":
		m, ok := sys.gs.tsys.Get("map1")
		if !ok {
			glog.Info("map not ok!!")
		} else {
			pkt.Reply("Rgetterrain", "map1", m)
		}
	default:
	}
}

func (sys *ServerNetworkSystem) listenproc() {
	for sys.IsRunning() {
		if con, err := sys.listencon.Accept(); err != nil {
			glog.Warningf("%s", err)
		} else {
			ws := NewWorldSession(sys.gs, sys, con)
			sys.gs.em.Emit("connect", ws)

			sys.do <- func(sys *ServerNetworkSystem) {
				glog.Infof("adding session %s", ws)
				sys.clients[con.RemoteAddr().String()] = ws
				go ws.ReceiveProc()
			}
		}
	}
}
