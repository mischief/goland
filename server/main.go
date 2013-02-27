package main

import (
	"encoding/gob"
	"flag"
	"github.com/mischief/goland/game/gnet"
	chanio "github.com/nu7hatch/gochanio"
	"github.com/trustmaster/goflow"
	"image"
	"log"
	"net"
)

var (
	proto      = "tcp"
	listenhost = ":61507"
)

func init() {
	gob.Register(&image.Point{})
}

func main() {
	flag.Parse()

	inPacket := make(chan *gnet.Packet, 5)

	gs := NewGameServer()

	gs.SetInPort("In", inPacket)

	log.Print("Starting network")
	flow.RunNet(gs)

	// funnel messages to inline
	l, err := net.Listen(proto, listenhost)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("new client %s", conn.RemoteAddr())

		read := chanio.NewReader(conn)

		if read == nil {
			panic("no reader")
		}

		go func(reader <-chan interface{}) {
			log.Printf("Reading from %s", conn.RemoteAddr())
			for x := range reader {
				log.Printf("%#v", x)
				p := x.(*gnet.Packet)
				p.Con = &conn

				inPacket <- p
			}
			log.Printf("Done with %s", conn.RemoteAddr())
		}(read)

	}

}
