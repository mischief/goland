package main

import (
	"fmt"
	"github.com/mischief/gochanio"
	goland "github.com/mischief/goland/game"
	"github.com/mischief/goland/game/gnet"
	"github.com/mischief/goland/game/gutil"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	uuid "github.com/nu7hatch/gouuid"
	"image"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

const (
	FPS_SAMPLES = 64
	FPS_LIMIT   = 23
)

var (
	CARDINALS = map[rune]goland.Direction{
		'w': goland.DIR_UP,
		'k': goland.DIR_UP,
		'a': goland.DIR_LEFT,
		'h': goland.DIR_LEFT,
		's': goland.DIR_DOWN,
		'j': goland.DIR_DOWN,
		'd': goland.DIR_RIGHT,
		'l': goland.DIR_RIGHT,
	}
)

type Stats struct {
	Samples  [64]float64
	Current  int
	MemStats runtime.MemStats

	fps float64
}

func (s Stats) String() string {
	return fmt.Sprintf("%5.2f FPS %5.2f MB %d GC %d GR", s.fps, float64(s.MemStats.HeapAlloc)/1000000.0, s.MemStats.NumGC, runtime.NumGoroutine())
}

func (s *Stats) Update(delta time.Duration) {

	runtime.ReadMemStats(&s.MemStats)

	s.Samples[s.Current%FPS_SAMPLES] = 1.0 / delta.Seconds()
	s.Current++

	for i := 0; i < FPS_SAMPLES; i++ {
		s.fps += s.Samples[i]
	}

	s.fps /= FPS_SAMPLES
}

type Game struct {
	Player *goland.Player

	Terminal
	*TermLog
	CloseChan chan bool

	stats Stats

	Objects goland.GameObjectMap
	//	Objects []*goland.GameObject // all known objects
	Map *goland.MapChunk

	Parameters *gutil.LuaParMap

	ServerCon net.Conn

	ServerRChan <-chan interface{}
	ServerWChan chan<- interface{}
}

func NewGame(params *gutil.LuaParMap) *Game {
	g := Game{}
	g.Objects = goland.NewGameObjectMap()
	g.Parameters = params

	g.CloseChan = make(chan bool, 1)

	//g.Objects = append(g.Objects, g.Player.GameObject)

	return &g
}

func (g *Game) SendPacket(p *gnet.Packet) {
	log.Printf("Game: SendPacket: %s", p)
	g.ServerWChan <- p
}

func (g *Game) Run() {

	g.Start()

	timer := goland.NewDeltaTimer()
	ticker := time.NewTicker(time.Second / FPS_LIMIT)

	run := true

	for run {
		select {
		case <-ticker.C:
			// frame tick
			delta := timer.DeltaTime()

			if delta.Seconds() > 0.25 {
				delta = time.Duration(250 * time.Millisecond)
			}

			g.Update(delta)
			g.Draw()

			g.Flush()

		case <-g.CloseChan:
			run = false
		}
	}

	g.End()

}

func (g *Game) Start() {
	log.Print("Game: Starting")

	// network setup
	server, ok1 := g.Parameters.Get("server")
	if !ok1 {
		log.Fatal("Game: Start: missing server in config")
	}

	con, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatalf("Game: Start: Dial: %s", err)
	}

	g.ServerCon = con

	g.ServerRChan = chanio.NewReader(g.ServerCon)
	g.ServerWChan = chanio.NewWriter(g.ServerCon)

	if g.ServerRChan == nil || g.ServerWChan == nil {
		log.Fatal("Game: Start: can't establish channels")
	}

	// login
	username, ok2 := g.Parameters.Get("username")
	if !ok2 {
		log.Fatal("Game: Start: missing username in config")
	}

	g.ServerWChan <- gnet.NewPacket("Tconnect", username)

	// request the map from server
	g.ServerWChan <- gnet.NewPacket("Tloadmap", nil)

	// request the object we control
	g.ServerWChan <- gnet.NewPacket("Tgetplayer", nil)

	// anonymous function that reads packets from the server
	go func(r <-chan interface{}) {
		for x := range r {
			p, ok := x.(*gnet.Packet)
			if !ok {
				log.Printf("Game: Read: Bogus server packet %#v", x)
				continue
			}

			g.HandlePacket(p)
		}
		log.Println("Game: Read: Disconnected from server!")
		io.WriteString(g.TermLog, "Disconnected from server!")
	}(g.ServerRChan)

	// terminal/keyhandling setup
	g.Terminal.Start()

	// chat dialog
	g.TermLog = NewTermLog(image.Pt(g.Terminal.Rect.Width-VIEW_START_X-VIEW_PAD_X, 5))

	// ESC to quit
	g.HandleKey(termbox.KeyEsc, func(ev termbox.Event) { g.CloseChan <- false })

	// convert to func SetupDirections()
	for k, v := range CARDINALS {
		func(c rune, d goland.Direction) {
			g.HandleRune(c, func(_ termbox.Event) {
				// lol collision
				p := &gnet.Packet{"Taction", CARDINALS[c]}
				g.SendPacket(p)
				newpos := g.Player.GetPos().Add(goland.DirTable[d])
				if g.Map.CheckCollision(nil, newpos) {
					g.Player.SetPos(newpos)
				}
			})

			/*
				      scale := PLAYER_RUN_SPEED
							upperc := unicode.ToUpper(c)
							g.HandleRune(upperc, func(_ termbox.Event) {
								for i := 0; i < scale; i++ {
									g.Player.Move(d)
								}
							})
			*/
		}(k, v)
	}

}

func (g *Game) End() {
	log.Print("Game: Ending")
	g.Terminal.End()
}

func (g *Game) Update(delta time.Duration) {
	// collect stats
	g.stats.Update(delta)

	g.RunInputHandlers()

	for _, o := range g.Objects {
		o.Update(delta)
	}

}

func (g *Game) Draw() {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// construct a current view of the 2d world and blit it
	viewwidth := g.Terminal.Rect.Width - VIEW_START_X - VIEW_PAD_X
	viewheight := g.Terminal.Rect.Height - VIEW_START_Y - VIEW_PAD_Y
	viewrect := tulib.Rect{VIEW_START_X, VIEW_START_Y, viewwidth, viewheight}
	viewbuf := tulib.NewBuffer(viewwidth, viewheight)
	viewbuf.Fill(viewrect, termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault})

	cam := NewCamera(viewbuf)

	if g.Player != nil {
		cam.SetCenter(g.Player.GetPos())
	} else {
		cam.SetCenter(image.Pt(256/2, 256/2))
	}

	// draw terrain
	if g.Map != nil {
		for x, row := range g.Map.Locations {
			for y, terr := range row {
				pos := image.Pt(x, y)
				if cam.ContainsWorldPoint(pos) {
					cam.Draw(terr, pos)
				}
			}
		}
	}

	// draw objects
	for _, o := range g.Objects {
		if cam.ContainsWorldPoint(o.GetPos()) {
			cam.Draw(o, o.GetPos())
		}
	}

	// draw labels
	statsparams := &tulib.LabelParams{termbox.ColorRed, termbox.ColorBlack, tulib.AlignLeft, '.', false}
	statsrect := tulib.Rect{1, 0, 60, 1}

	statsstr := fmt.Sprintf("%s TERM %s", g.Terminal.Size(), g.stats)

	playerparams := &tulib.LabelParams{termbox.ColorRed, termbox.ColorBlack, tulib.AlignLeft, '.', false}
	playerrect := tulib.Rect{1, g.Terminal.Rect.Height - 1, g.Terminal.Rect.Width, 1}

	playerstr := fmt.Sprintf("%s Cam.Pos: %s Cam.Rect: %v", g.Player, cam.Pos, cam.Rect)

	g.Terminal.DrawLabel(statsrect, statsparams, []byte(statsstr))
	g.Terminal.DrawLabel(playerrect, playerparams, []byte(playerstr))

	g.TermLog.Draw(&g.Terminal.Buffer, image.Pt(1, g.Terminal.Rect.Height-6))

	// blit
	g.Terminal.Blit(viewrect, 0, 0, &viewbuf)

}

// deal with gnet.Packets received from the server
func (g *Game) HandlePacket(pk *gnet.Packet) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Game: HandlePacket: %s", err)
		}
	}()

	log.Printf("Game: HandlePacket: %s", pk)
	switch pk.Tag {

	// Rchat: we got a text message
	case "Rchat":
		chatline := pk.Data.(string)
		io.WriteString(g.TermLog, chatline)

		// Raction: something moved
	case "Raction":
		pl := pk.Data.(*goland.Player)

		for _, o := range g.Objects {
			if *o.ID == *pl.ID {
				o.SetPos(pl.GetPos())
			}
		}

		// Rnewobject: new object we need to track
	case "Rnewobject":
		obj := pk.Data.(*goland.GameObject)
		g.Objects.Add(obj)

		// Rdelobject: some object went away
	case "Rdelobject":
		obj := pk.Data.(*goland.GameObject)
		g.Objects.RemoveObject(obj)

		// Rgetplayer: find out who we control
	case "Rgetplayer":
		playerid := pk.Data.(*uuid.UUID)

		pl := g.Objects.FindObjectByID(playerid)
		if pl != nil {
			player := goland.NewPlayer("") // name is "" because the gameobject from the wire has our name
			player.GameObject = pl
			g.Player = player
		} else {
			log.Printf("Game: HandlePacket: can't find our player %s", playerid)
		}

		// Rloadmap: get the map data from the server
	case "Rloadmap":
		gmap := pk.Data.(*goland.MapChunk)
		g.Map = gmap

	default:
		log.Printf("bad packet tag %s", pk.Tag)
	}

}
