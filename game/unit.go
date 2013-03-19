// Unit is a non-static game object
package game

import (
	"fmt"
)

const (
	DEFAULT_HP = 10
)

type Unit struct {
	*GameObject
	*Inventory

	Level     int
	Hp, HpMax int
}

func NewUnit(name string) *Unit {
	u := &Unit{Level: 1,
		Hp:    DEFAULT_HP,
		HpMax: DEFAULT_HP,
	}
	u.Inventory = NewInventory()
	u.GameObject = NewGameObject(name)

	return u
}

// Checks if a Unit HasItem *Item
func (u Unit) HasItem(i *Item) bool {
	if u.Inventory.ContainsItem(i) {
		return true
	}
	return false
}

func (u Unit) String() string {
	return fmt.Sprintf("%s: Hp: %d(%d) %s", u.Name, u.Hp, u.HpMax, u.GameObject)
}
