package game

import "container/list"

type Subject interface {
	Attach(obs Observer)
	Detach(obs Observer)
	Notify()
}

type Observer interface {
	Update()
}

type DefaultSubject struct {
	Observers *list.List
}

func NewDefaultSubject() *DefaultSubject {
	return &DefaultSubject{list.New()}
}

func (sub *DefaultSubject) Attach(obs Observer) {
	sub.Observers.PushBack(obs)
}

func (sub *DefaultSubject) Detach(obs Observer) {
	for o := sub.Observers.Front(); o != nil; o = o.Next() {
		if o.Value == obs {
			sub.Observers.Remove(o)
		}
	}
}

func (sub *DefaultSubject) Notify() {
	for s := sub.Observers.Front(); s != nil; s = s.Next() {
		s.Value.(Observer).Update()
	}
}
