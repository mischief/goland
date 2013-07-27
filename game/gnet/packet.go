// Packet: message type for flow procs and chanio connections
package gnet

import (
	"encoding/gob"
	"fmt"
	termbox "github.com/nsf/termbox-go"
	"image"
)

func init() {
	gob.Register(&Packet{})
	gob.Register(&PropPosAdd{})
	gob.Register(&PropSpriteAdd{})
}

type Packet struct {
	Tag  string      // packet tag identifying the operation
	Data []interface{} // packet payload
}

func (p Packet) String() string {
	dat := "nil"
	if p.Data != nil {
		dat = fmt.Sprintf("%v", p.Data)
	}

	return fmt.Sprintf("(%s %s)", p.Tag, dat)
}

func NewPacket(tag string, data ...interface{}) *Packet {
	return &Packet{tag, data}
}

type PropPosAdd struct {
	Id  string
	Pos image.Point
}

type PropSpriteAdd struct {
	Id     string
	Sprite termbox.Cell
}
