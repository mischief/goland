package game

import (
	"container/list"
	"sync"
)

type Subject interface {
	Attach(obs Observer)
	Detach(obs Observer)
	Notify()
}

type Observer interface {
	Update()
}

type DefaultSubject struct {
	sync.Mutex
	Observers *list.List
}

func NewDefaultSubject() *DefaultSubject {
	return &DefaultSubject{Observers: list.New()}
}

func (sub *DefaultSubject) Attach(obs Observer) {
	sub.Lock()
	sub.Observers.PushBack(obs)
	sub.Unlock()
}

func (sub *DefaultSubject) Detach(obs Observer) {
	sub.Lock()
	for o := sub.Observers.Front(); o != nil; o = o.Next() {
		if o.Value == obs {
			sub.Observers.Remove(o)
		}
	}
	sub.Unlock()
}

func (sub *DefaultSubject) Notify() {
	for s := sub.Observers.Front(); s != nil; s = s.Next() {
		s.Value.(Observer).Update()
	}
}
