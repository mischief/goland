// Inventory is a collection of items
package game

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
)

const (
	// weight? #items?
	DEFAULT_INVENTORY_CAP = 10
)

type Inventory struct {
	Items map[uuid.UUID]*Item
	Capacity int
}

func NewInventory() *Inventory {
	i := &Inventory{Capacity: DEFAULT_INVENTORY_CAP,
		Items: make(map[uuid.UUID]*Item),
	}
	return i
}

func (inv Inventory) AddItem(i *Item) {
	inv.Items[*i.ID] = i
}

// Removes an item from an Invetory yet 
// returns the dropped item to the caller
// for further processing
func (inv Inventory) DropItem(i *Item) *Item{
	delete(inv.Items, *i.ID)
	return i
}

func (inv Inventory) DestroyItem(i *Item) {
	delete(inv.Items, *i.ID)
}

// Iterates over Inventory items and prints their attrs
// Should intelligently handle printing items in qty > 1
// assuming the items also have the same properties (modifiers, etc)
func (u Inventory) String() string {
	return fmt.Sprintf("%s", "XXX Placeholder")
}