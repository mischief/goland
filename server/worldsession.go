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
	Con         net.Conn           // connection to the client
	ClientRChan <-chan interface{} // client's incoming message channel
	ClientWChan chan<- interface{} // client's outgoing message channel
	ID          uuid.UUID          // client's id (account id?)
	Username    string             // associated username
	Pos         image.Point        // XXX: what's this for?
	Player      game.Object        // object this client controls
	World       *GameServer        // world reference
}

func (ws *WorldSession) String() string {
	return fmt.Sprintf("%s %s %s %s", ws.Con.RemoteAddr(), ws.ID, ws.Pos, ws.Player)
}

func NewWorldSession(w *GameServer, c net.Conn) *WorldSession {
	var err error
	var id *uuid.UUID

	n := new(WorldSession)

	n.Con = c

	n.ClientRChan = chanio.NewReader(n.Con)
	n.ClientWChan = chanio.NewWriter(n.Con)

	if id, err = uuid.NewV4(); err != nil {
		log.Printf("NewWorldSession: %s", err)
		return nil
	} else {
		n.ID = *id
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
