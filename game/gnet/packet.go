// Packet: message type for flow procs and chanio connections
package gnet

import "encoding/gob"
import "fmt"
import "net"

type Packet struct {
	Con  *net.Conn   // which connection this packet is tied to
	Tag  string      // packet tag identifying the operation
	Data interface{} // packet payload
}

func (p Packet) String() string {
	return fmt.Sprintf("%v", p)
}

func init() {
	gob.Register(&Packet{})
}
