package main

import (
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"strings"
	"time"
)

var (
	art string = `
             ..ooo@@@XXX%%%xx..            
          .oo@@XXX%x%xxx..     ' .         
        .o@XX%%xx..               ' .      
      o@X%..                  ..ooooooo    
    .@X%x.                 ..o@@^^   ^^@@o 
  .ooo@@@@@@ooo..      ..o@@^          @X% 
  o@@^^^     ^^^@@@ooo.oo@@^             % 
 xzI    -*--      ^^^o^^        --*-     % 
 @@@o     ooooooo^@@^o^@X^@oooooo     .X%x 
I@@@@@@@@@XX%%xx  ( o@o ) %x@@@@@xXX@@@@X%x
I@@@@XX%%xx  oo@@@@X% @@X%x   ^^^@@@@@@@X%x
 @X%xx     o@@@@@@@X% @@XX%%x  )    ^^@X%x 
  ^   xx o@@@@@@@@Xx  ^ @XX%%x    xxx      
        o@@^^^ooo I^^ I^o ooo   .  x       
        oo @^ IX      I   ^X  @^ oo        
        IX     U  .        V     IX        
         V     .           .     V         

goland

enter to play
q to exit
`
)

type IntroPanel struct {
	do chan func(*IntroPanel)
	*graphics.BasePanel
	g *Game
}

func (g *Game) IntroPanel() *IntroPanel {
	ip := &IntroPanel{
		do:        make(chan func(*IntroPanel), 10),
		BasePanel: graphics.NewPanel(),
		g:         g,
	}

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		ip.do <- func(ip *IntroPanel) {
			ip.Resize(ev.Width, ev.Height)
		}
	})

	return ip
}

func (ip *IntroPanel) Draw() {
	if ip.IsActive() {
		ip.Clear()

		for i, l := range strings.Split(art, "\n") {
			graphics.WriteCenteredLine(ip, l, i, graphics.TextStyle.Fg, graphics.TextStyle.Bg)
		}

		ip.BasePanel.Draw()
	}
}

func (ip *IntroPanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-ip.do:
			f(ip)
		default:
			return
		}
	}
}

func (ip *IntroPanel) HandleInput(ev termbox.Event) {
	ip.do <- func(ip *IntroPanel) {
		if ev.Ch != 0 {
			switch ev.Ch {
			case 'q':
				ip.g.Quit()
			}
		} else {
			switch ev.Key {
			case termbox.KeyEnter:
				ip.g.rsys.PopInputHandler()
			}
		}
	}
}
