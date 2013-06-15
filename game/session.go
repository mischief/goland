package game

import (
	"github.com/mischief/goland/game/gnet"
	uuid "github.com/nu7hatch/gouuid"
)

type Session interface {
	SendPacket(pk *gnet.Packet)
	GetPlayer() *Player
	ID() uuid.UUID
	Username() string
}
