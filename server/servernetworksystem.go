package main

import (
	"github.com/golang/glog"
	//"github.com/mischief/gochanio"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"net"
	"sync/atomic"
	"time"
)

type ServerNetworkSystem struct {
	do      chan func(*ServerNetworkSystem)
	running int32
	scene   *game.Scene

	listen      string             // server listen string
	listencon   net.Listener       // server tcp connection
	serverrchan <-chan interface{} // server read channel
	serverwchan chan<- interface{} // server write channel
}

func NewServerNetworkSystem(s *game.Scene, listen string) (*ServerNetworkSystem, error) {
	sys := &ServerNetworkSystem{
		do:     make(chan func(*ServerNetworkSystem)),
		scene:  s,
		listen: listen,
	}

	sys.scene.Wg.Add(1)

	if err := game.StartSystem(sys, true); err != nil {
		return nil, err
	}
	sys.Syn()
	return sys, nil
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
	sys.scene.Wg.Done()
}

func (sys *ServerNetworkSystem) Update(delta time.Duration) {
	// TODO: send/recieve packets
}

func (sys *ServerNetworkSystem) SendPacket(tag string, data interface{}) {
	sys.do <- func(sys *ServerNetworkSystem) {
		pkt := gnet.NewPacket(tag, data)
		glog.Infof("sendpacket: sending %s", pkt)
		sys.serverwchan <- pkt
	}
}
