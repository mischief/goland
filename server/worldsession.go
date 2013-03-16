package main

import (
	"fmt"
	"github.com/mischief/gochanio"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	uuid "github.com/nu7hatch/gouuid"
	"image"
	"log"
	"net"
)

type WorldSession struct {
	Con net.Conn

	ClientRChan <-chan interface{}
	ClientWChan chan<- interface{}

	ID       *uuid.UUID
	Username string

	Pos image.Point

	*game.Player

	World *GameServer
}

func (ws *WorldSession) String() string {
	return fmt.Sprintf("%s %s %s %s", ws.Con.RemoteAddr(), ws.ID, ws.Pos, ws.Player)
}

func NewWorldSession(w *GameServer, c net.Conn) *WorldSession {
	var err error

	n := new(WorldSession)

	n.Con = c

	n.ClientRChan = chanio.NewReader(n.Con)
	n.ClientWChan = chanio.NewWriter(n.Con)

	if n.ID, err = uuid.NewV4(); err != nil {
		log.Printf("NewWorldSession: %s", err)
		return nil
	}

	n.World = w

	return n
}

// handle per-client packets
func (ws *WorldSession) ReceiveProc() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("WorldSession: ReceiveProc: %s", err)
		}
	}()

	for x := range ws.ClientRChan {
		p, ok := x.(*gnet.Packet)
		if !ok {
			log.Printf("WorldSession: ReceiveProc: bad packet %#v from %s", x, ws.Con.RemoteAddr())
			continue
		}

		cp := &ClientPacket{ws, p}

		ws.World.PacketChan <- cp
	}

	dis := &ClientPacket{ws, gnet.NewPacket("Tdisconnect", nil)}

	ws.World.PacketChan <- dis

	log.Printf("WorldSession: ReceiveProc: Channel closed %s", ws)
}

// send packet to this client
func (ws *WorldSession) SendPacket(pk *gnet.Packet) {
	ws.ClientWChan <- pk
}

func (ws *WorldSession) Update() {
}
