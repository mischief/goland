// Inventory is a collection of items
package game

import (
	"fmt"

//	uuid "github.com/nu7hatch/gouuid"
)

const (
	// weight? #items?
	DEFAULT_INVENTORY_CAP = 10
)

type Inventory struct {
	Items    map[int]*Item
	Capacity int
}

func NewInventory() *Inventory {
	i := &Inventory{Capacity: DEFAULT_INVENTORY_CAP,
		Items: make(map[int]*Item),
	}
	return i
}

// Returns an item with .Name == name if it exists, otherwise false
func (inv Inventory) ContainsItemNamed(name string) bool {
	for _, i := range inv.Items {
		if i.GetName() == name {
			return true
		}
	}
	return false
}

func (inv Inventory) ContainsItem(i *Item) bool {
	_, exists := inv.Items[i.GetID()]
	return exists
}

func (inv Inventory) AddItem(i *Item) {
	inv.Items[i.GetID()] = i
}

// Removes an item from an Invetory yet
// returns the dropped item to the caller
// for further processing
func (inv Inventory) DropItem(i *Item) *Item {
	delete(inv.Items, i.GetID())
	return i
}

// Assumes Item exists in Inventory (or panics)
func (inv Inventory) GetItemNamed(name string) *Item {
	for _, i := range inv.Items {
		if i.GetName() == name {
			return i
		}
	}
	return NewItem("error") // XXX!
}

func (inv Inventory) DestroyItem(i *Item) {
	delete(inv.Items, i.GetID())
}

// Iterates over Inventory items and prints their attrs
// Should intelligently handle printing items in qty > 1
// assuming the items also have the same properties (modifiers, etc)
func (i Inventory) String() string {
	return fmt.Sprintf("Inventory: %s", i.Items)
}
