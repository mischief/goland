//
package game

import (
	"encoding/gob"
	"fmt"
)

func init() {
	gob.Register(&Item{})
}

type Item struct {
	*GameObject

	Desc string
	Weight int
	Modifier int
}

func NewItem(name string) *Item {
	o := NewGameObject(name)
	o.Glyph = GLYPH_ITEM
	i := &Item{Desc: "",
	        GameObject: o,
	        Weight: 1,
	        Modifier: 0,	
	}
	i.Tags["item"] = true
	return i
}

func (i Item) String() string {
	return fmt.Sprintf("%s: <weight: %s, mod: %s, desc: %s, %s>", i.Name, i.Weight, i.Modifier, i.Desc, i.GameObject)
}