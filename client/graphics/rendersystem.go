package graphics

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/errnoh/termbox/panel"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gutil"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"runtime"
	"sync/atomic"
	"time"
  "github.com/golang/glog"
)

type KeyHandler func(ev termbox.Event)

// RenderSystem controls graphics in the game client.
// TODO: move termbox stuffs here
type RenderSystem struct {
	// action queue
	do chan func(*RenderSystem)

	// atomic running flag
	running int32

	// scene containing actors
	scene *game.Scene

	// chan of input events from termbox
	termchan chan termbox.Event

	// display panels
  mainpanel *panel.Buffered
	panels map[string]panel.Panel

	// Stack of input panels
	inputstack *list.List

	// key bindings
	runehandlers map[rune]KeyHandler
	keyhandlers  map[termbox.Key]KeyHandler
}

func NewRenderSystem(s *game.Scene) (*RenderSystem, error) {
	sys := &RenderSystem{
		do:           make(chan func(*RenderSystem), 5),
		scene:        s,
		termchan:     make(chan termbox.Event),
    mainpanel:    panel.MainScreen(),
    panels:       make(map[string]panel.Panel),
		inputstack:   list.New(),
		runehandlers: make(map[rune]KeyHandler),
		keyhandlers:  make(map[termbox.Key]KeyHandler),
	}

  sys.scene.Wg.Add(1)
	//game.StartSystem(sys, true)
  if err := sys.Start(true); err != nil {
    return nil, err
  }

	sys.Syn()
	return sys, nil
}

func (sys *RenderSystem) DoOne() error {
	f := <-sys.do
	f(sys)

	return nil
}

func (sys *RenderSystem) Syn() {
	ack := make(chan bool)
	sys.do <- func(sys *RenderSystem) {
		ack <- true
	}
	<-ack
	close(ack)
}

func (sys *RenderSystem) Running(r bool) {
	if r {
		atomic.CompareAndSwapInt32(&sys.running, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&sys.running, 1, 0)
	}
}

func (sys *RenderSystem) IsRunning() bool {
	if atomic.LoadInt32(&sys.running) == 1 {
		return true
	}

	return false
}

func (sys *RenderSystem) Stop() {
	if sys.IsRunning() {
		sys.do <- func(sys *RenderSystem) {
			sys.Running(false)
		}
	}
}

func (sys *RenderSystem) Frequency() time.Duration {
	return 40 * time.Millisecond
}

func (sys *RenderSystem) Setup() error {
	glog.Info("setup: starting")

	if err := termbox.Init(); err != nil {
		return fmt.Errorf("setup: termbox: %s", err)
	}

	termbox.HideCursor()

	glog.Info("setup: complete")
	return nil
}

func (sys *RenderSystem) Tick(delta time.Duration) {
	d := delta
	sys.do <- func(sys *RenderSystem) {
		sys.updateGraphics(d)
	}
}

func (sys *RenderSystem) TearDown() {
	termbox.Close()
	glog.Info("teardown: complete")
  sys.scene.Wg.Done()
}

func (sys *RenderSystem) Start(lockOsThread bool) (err error) {
	if sys.IsRunning() {
		return errors.New("system already started")
	} else {

		go func() {
			if lockOsThread {
				runtime.LockOSThread()
			}

			if e := sys.Setup(); e != nil {
				err = e
				return
			}

			sys.Running(true)

			for sys.IsRunning() {
				sys.DoOne()
			}

      sys.TearDown()
		}()

		sys.Syn()

		// graphics
		gticker := time.Tick(sys.Frequency())
		gtimer := game.NewDeltaTimer()

		go func() {
			for sys.IsRunning() {
				<-gticker
				sys.Tick(gtimer.DeltaTime())
			}
		}()

		// input
		iticker := time.Tick(sys.Frequency())
		itimer := game.NewDeltaTimer()

		// goroutine reading directly from termbox
		go func() {
			for sys.IsRunning() {
				ev := termbox.PollEvent()
				if ev.Type == termbox.EventError {
					glog.Infof("termbox poller: error: %s", ev.Err)
					close(sys.termchan)
				} else {
					sys.termchan <- ev
				}
			}
		}()

		// goroutine regularly reading event queue and
		// sending messages to the rendersystem
		go func() {
			for sys.IsRunning() {
				<-iticker
				sys.do <- func(sys *RenderSystem) {
					sys.updateInput(itimer.DeltaTime())
				}
			}
		}()

	}

	return nil
}

// TODO: Draw stuff here
func (sys *RenderSystem) updateGraphics(delta time.Duration) {
	//glog.Infof("rendersystem: updating %v", delta)

  sys.mainpanel.Clear()

  for _, p := range sys.panels {
    if v, ok := p.(InputHandler); ok {
      v.HandleInput(termbox.Event{Type: termbox.EventResize})
    }

    if v, ok := p.(gutil.Updater); ok {
      v.Update(delta)
    }

    if v, ok := p.(panel.Drawer); ok {
      v.Draw()
    }
  }

  for _, actor := range sys.scene.Find(game.PropPos, game.PropCamera) {
    pos := actor.Get(game.PropPos).(*game.Pos)
    cam := actor.Get(game.PropCamera).(*Camera)
    cam.SetCenter(<-pos.Get())
  }
	/*
		for _, actor := range sys.scene.Find(game.PropPos, game.PropStaticSprite) {
			p := actor.Get(game.PropPos).(*game.Pos)
			sp := actor.Get(game.PropStaticSprite).(*StaticSprite)
		}
	*/

  termbox.Flush()

}

// Read all pending input and send off to handlers
func (sys *RenderSystem) updateInput(delta time.Duration) {
stop:
	for {
		select {
		case ev := <-sys.termchan:
			glog.Infof("termbox event: %s", tulib.KeyToString(ev.Key, ev.Ch, ev.Mod))
			switch ev.Type {

			case termbox.EventKey:

				sys.do <- func(sys *RenderSystem) {

					if ev.Ch != 0 {

						if h, ok := sys.runehandlers[ev.Ch]; ok {
							sys.do <- func(sys *RenderSystem) {
								h(ev)
							}
						}

					} else {

						if h, ok := sys.keyhandlers[ev.Key]; ok {
							sys.do <- func(sys *RenderSystem) {
								h(ev)
							}
						}

					}
				}

			case termbox.EventResize:
				// TODO: fix
			}

		default:
			break stop
		}
	}
}

// Push an input handler on the stack.
// Recieves input until it is popped.
func (sys *RenderSystem) PushInputHandler(i interface{}) {
	sys.do <- func(sys *RenderSystem) {
		sys.inputstack.PushFront(i)
	}
}

func (sys *RenderSystem) PushPanelInputName(n string) {
  sys.do <- func(sys *RenderSystem) {
    if p, ok := sys.panels[n]; ok {
      sys.inputstack.PushFront(p)
    }
  }
}

// Pop one input handler off the stack
// Won't pop if len < 1
func (sys *RenderSystem) PopInputHandler() {
	sys.do <- func(sys *RenderSystem) {
		if sys.inputstack.Len() > 1 {
			e := sys.inputstack.Front()
			sys.inputstack.Remove(e)
		}
	}
}

// Hook a rune to a handler
func (sys *RenderSystem) HandleRune(r rune, h KeyHandler) {
	sys.do <- func(sys *RenderSystem) {
		sys.runehandlers[r] = h
	}
}

// Hook a termbox key (eg space, pageup, arrow keys) to a handler
func (sys *RenderSystem) HandleKey(k termbox.Key, h KeyHandler) {
	sys.do <- func(sys *RenderSystem) {
		sys.keyhandlers[k] = h
	}
}

// Add a panel to the render system
func (sys *RenderSystem) AddPanel(n string, p panel.Panel) {
  sys.do <- func(sys *RenderSystem) {
    sys.panels[n] = p
  }
}
