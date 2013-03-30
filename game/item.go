//Item is any construct which can be stored in a Unit's inventory
package game

import (
	"encoding/gob"
	"fmt"
)

func init() {
	gob.Register(&Item{})
}

type Item struct {
	Object

	Desc     string
	Weight   int
	Modifier int
}

func NewItem(name string) *Item {
	i := &Item{Object: NewGameObject(name)}
	//	i := BootstrapItem(o)
	//	i.Tags["visible"] = true
	//	i.Tags["gettable"] = true
	return i
}

func (i Item) String() string {
	return fmt.Sprintf("%s: <weight: %s, mod: %s, desc: %s, %s>",
		i.GetName(), i.Weight, i.Modifier, i.Desc, i.Object)
}
