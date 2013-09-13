package game

import (
	"fmt"
	"sync"
)

type Actor struct {
	ID         string
	properties map[PropType]Property
	scene      *Scene

	m sync.Mutex
}

func (a Actor) String() string {
	return fmt.Sprintf("%s", a.ID)
}

func NewActor(id string) *Actor {
	return &Actor{ID: id, properties: make(map[PropType]Property)}
}

func (a *Actor) Add(p Property) error {
	a.m.Lock()
	defer a.m.Unlock()

	t := p.Type()
	if _, ok := a.properties[t]; ok {
		return fmt.Errorf("actor: add: %v already has property type %v", a.ID, t)
	}

	a.properties[t] = p
	a.scene.cache(a, t)

	return nil
}

func (a *Actor) Get(t PropType) Property {
	a.m.Lock()
	defer a.m.Unlock()

	if prop, ok := a.properties[t]; !ok {
		panic(fmt.Sprintf("%s no such property %v", a, t))
	} else {
		return prop
	}
}

func (a *Actor) Remove(t PropType) (rem Property, present bool) {
	a.m.Lock()
	defer a.m.Unlock()

	rem, present = a.properties[t]
	delete(a.properties, t)
	a.scene.uncache(a, t)
	return
}
