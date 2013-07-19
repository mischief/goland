package game

import (
	"fmt"
	"github.com/golang/glog"
	"image"
	"sync/atomic"
	"time"
)

type MovementSystem struct {
	do      chan func(*MovementSystem)
	running int32
	scene   *Scene
}

func NewMovementSystem(s *Scene) (*MovementSystem, error) {
	sys := &MovementSystem{
		do:    make(chan func(*MovementSystem)),
		scene: s,
	}

	sys.scene.Wg.Add(1)

	if err := StartSystem(sys, true); err != nil {
		return nil, err
	}

	sys.Syn()
	return sys, nil
}

func (sys *MovementSystem) DoOne() error {
	f := <-sys.do
	f(sys)

	return nil
}

func (sys *MovementSystem) Syn() {
	ack := make(chan bool)
	sys.do <- func(sys *MovementSystem) {
		ack <- true
	}
	<-ack
	close(ack)
}

func (sys *MovementSystem) Running(r bool) {
	if r {
		atomic.CompareAndSwapInt32(&sys.running, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&sys.running, 1, 0)
	}
}

func (sys *MovementSystem) IsRunning() bool {
	if atomic.LoadInt32(&sys.running) == 1 {
		return true
	}

	return false
}

func (sys *MovementSystem) Stop() {
	if sys.IsRunning() {
		sys.do <- func(sys *MovementSystem) {
			sys.Running(false)
		}
	}
}

func (sys *MovementSystem) Frequency() time.Duration {
	return 500 * time.Millisecond
}

func (sys *MovementSystem) Setup() error {
	glog.Info("setup: complete")
	return nil
}

func (sys *MovementSystem) Tick(timestep time.Duration) {
	ts := timestep
	sys.do <- func(sys *MovementSystem) {
		sys.Update(ts)
	}
}

func (sys *MovementSystem) TearDown() {
	glog.Info("teardown: complete")
	sys.scene.Wg.Done()
}

func (sys *MovementSystem) Update(delta time.Duration) {
	//glog.Printf("updating %v", delta)

}

func (sys *MovementSystem) Pos() *Pos {
	p := &Pos{do: make(chan func(*Pos))}
	p.start()
	return p
}

type Pos struct {
	do  chan func(*Pos)
	pos image.Point
}

func (p Pos) String() string {
	return fmt.Sprintf("%s", p.pos)
}

func (p *Pos) Type() PropType {
	return PropPos
}

func (p *Pos) start() {
	go func() {
		for f := range p.do {
			f(p)
		}
	}()
}

func (p *Pos) Set(newpos image.Point) {
	np := newpos
	p.do <- func(pp *Pos) {
		pp.pos = np
	}
}

func (p *Pos) Get() <-chan image.Point {
	ch := make(chan image.Point)
	p.do <- func(pp *Pos) {
		ch <- pp.pos
		close(ch)
	}

	return ch
}
