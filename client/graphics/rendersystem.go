package graphics

import (
	"container/list"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/errnoh/termbox/panel"
	"github.com/golang/glog"
	"github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gterrain"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"sync/atomic"
	"time"
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

	// event emitter
	emitter *emission.Emitter

	// display panels
	mainpanel *panel.Buffered
	panels    map[string]GamePanel

	// Stack of 'active' panels
	activestack *list.List

	// key bindings
	runehandlers map[rune]KeyHandler
	keyhandlers  map[termbox.Key]KeyHandler

	// current terrain. TODO: fix this
	Terrain *gterrain.TerrainChunk
}

func NewRenderSystem(s *game.Scene, em *emission.Emitter) (*RenderSystem, error) {
	sys := &RenderSystem{
		do:           make(chan func(*RenderSystem), 10),
		scene:        s,
		emitter:      em,
		mainpanel:    panel.MainScreen(),
		panels:       make(map[string]GamePanel),
		activestack:  list.New(),
		runehandlers: make(map[rune]KeyHandler),
		keyhandlers:  make(map[termbox.Key]KeyHandler),
	}

	if err := game.StartSystem(sys, true); err != nil {
		return nil, err
	}

	sys.scene.AddSystem(sys)

	sys.Syn()
	return sys, nil
}

func (sys RenderSystem) String() string {
	return "render"
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
	glog.Info("setup: begin")

	if err := termbox.Init(); err != nil {
		return fmt.Errorf("setup: termbox: %s", err)
	}

	termbox.HideCursor()

	go func() {
		for sys.IsRunning() {
			ev := termbox.PollEvent()
			if ev.Type == termbox.EventError {
				glog.Infof("termbox poller: error: %s", ev.Err)
			} else {
				switch ev.Type {
				case termbox.EventKey:
					sys.emitter.Emit("key", ev)
				case termbox.EventResize:
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
					sys.emitter.Emit("resize", ev)
				}
			}
		}
	}()

	// TODO: remove this and make better keyevent to game event translator
	sys.emitter.On("key", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		if glog.V(2) {
			glog.Infof("key event: %s", tulib.KeyToString(ev.Key, ev.Ch, ev.Mod))
		}

		sys.do <- func(sys *RenderSystem) {
			sys.handleInput(ev)
		}
	})

	glog.Info("setup: complete")
	return nil
}

func (sys *RenderSystem) Tick(delta time.Duration) {
	d := delta
	sys.do <- func(sys *RenderSystem) {
		sys.Update(d)
	}
}

func (sys *RenderSystem) TearDown() {
	termbox.Close()
	glog.Info("teardown: complete")
	sys.scene.RemoveSystem(sys)
}

// Draw stuff here
func (sys *RenderSystem) Update(delta time.Duration) {
	if glog.V(3) {
		glog.Infof("rendersystem: updating %v", delta)
	}

	sys.mainpanel.Clear()

	for _, p := range sys.panels {
		p.Update(delta)

		if v, ok := p.(panel.Drawer); ok {
			v.Draw()
		}
	}

	// draw the 'active' panels over everything else
	for act := sys.activestack.Back(); act != nil; act = act.Prev() {
		if v, ok := act.Value.(panel.Drawer); ok {
			v.Draw()
		}
	}

	// update camera positions
	for _, actor := range sys.scene.Find(game.PropPos, game.PropCamera) {
		pos := actor.Get(game.PropPos).(*game.Pos)
		cam := actor.Get(game.PropCamera).(*Camera)
		cam.SetCenter(<-pos.Get())
	}

	// drawing game stuff is handled by viewpanel for now
	/*
		for _, actor := range sys.scene.Find(game.PropPos, game.PropStaticSprite) {
			p := actor.Get(game.PropPos).(*game.Pos)
			sp := actor.Get(game.PropStaticSprite).(*StaticSprite)
		}
	*/

	termbox.Flush()
}

// Read all pending input and send off to handlers
func (sys *RenderSystem) handleInput(ev termbox.Event) {
	if sys.activestack.Len() > 0 {
		e := sys.activestack.Front()
		ih := e.Value.(InputHandler)
		ih.HandleInput(ev)
	} else if ev.Ch != 0 {

		if h, ok := sys.runehandlers[ev.Ch]; ok {
			//sys.do <- func(sys *RenderSystem) {
			h(ev)
			//}
		}

	} else {

		if h, ok := sys.keyhandlers[ev.Key]; ok {
			//sys.do <- func(sys *RenderSystem) {
			h(ev)
			//}
		}
	}

}

// Push an input handler on the stack.
// Recieves input until it is popped.
func (sys *RenderSystem) PushActivePanel(p GamePanel) {
	sys.do <- func(sys *RenderSystem) {
		sys.activestack.PushFront(p)

		if a, ok := p.(Activator); ok {
			a.Activate()
		}
	}
}

// Push an input handler by panel name.
// Only used for chatpanel for now.
func (sys *RenderSystem) PushActivePanelName(n string) {
	sys.do <- func(sys *RenderSystem) {
		if p, ok := sys.panels[n]; ok {
			glog.Infof("switching to panel %s", n)
			sys.activestack.PushFront(p)
			if a, ok := p.(Activator); ok {
				a.Activate()
			}
		}
	}
}

// Pop one input handler off the stack
// Will only pop if len > 0
func (sys *RenderSystem) PopInputHandler() {
	sys.do <- func(sys *RenderSystem) {
		if sys.activestack.Len() > 0 {
			e := sys.activestack.Front()
			sys.activestack.Remove(e)
			if a, ok := e.Value.(Activator); ok {
				a.Deactivate()
			}
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
func (sys *RenderSystem) AddPanel(p GamePanel) {
	sys.do <- func(sys *RenderSystem) {
		sys.panels[p.GetTitle()] = p
	}
}

func (sys *RenderSystem) SetTerrainChunk(tc *gterrain.TerrainChunk) {
	sys.do <- func(sys *RenderSystem) {
		sys.Terrain = tc
	}
}
