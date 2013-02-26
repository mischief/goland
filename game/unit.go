// Unit is a non-static game object
package game

import (
	"fmt"
)

const (
	DEFAULT_HP = 10
)

type Unit struct {
	GameObject

	Level     int
	Hp, HpMax int
}

func NewUnit() *Unit {
	u := &Unit{Level: 1,
		Hp:    DEFAULT_HP,
		HpMax: DEFAULT_HP,
	}

	u.GameObject = *NewGameObject()

	return u
}

func (u Unit) String() string {
	return fmt.Sprintf("Hp: %d(%d) %s", u.Hp, u.HpMax, &u.GameObject)
}
