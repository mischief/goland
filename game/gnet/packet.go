// Packet: message type for flow procs and chanio connections
package gnet

import (
	"encoding/gob"
	"fmt"
)

type Packet struct {
	Tag  string      // packet tag identifying the operation
	Data interface{} // packet payload
}

func (p Packet) String() string {
	return fmt.Sprintf("(%s %s)", p.Tag, p.Data)
}

func NewPacket(tag string, data interface{}) *Packet {
	return &Packet{tag, data}
}

func init() {
	gob.Register(&Packet{})
}
