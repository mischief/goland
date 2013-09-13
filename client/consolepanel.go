package main

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/mischief/goland/client/graphics"
	"github.com/nsf/termbox-go"
	"github.com/stevedonovan/luar"
	"os/exec"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

type ConsoleCommand struct {
	Name string                     // how to call
	Desc string                     // short description
	Help string                     // long help
	Fn   func(args ...string) error // function
}

var (
	// informational crap
	info = map[string][]string{
		"keybindings": []string{"keybinding information", "esc: quit menu", "wasd or hjkl: move", "`: open console", "enter: start chatting"},
		"source":      []string{"source of the engine", "  https://github.com/mischief/goland"},
		"irc":         []string{"irc channel of the engine", "  ircs://offblast.org/#goland"},
	}
)

const (
	NLINES = 100
)

type ConsoleLine struct {
	Cell termbox.Cell
	Line string
}

// ConsolePanel is a panel which runs the in-game chat and command input.
type ConsolePanel struct {
	do                  chan func(*ConsolePanel)
	*graphics.BasePanel              // Panel
	buf                 bytes.Buffer // Buffer for keyboard input
	g                   *Game

	// commands
	prompt string
	cmds   map[string]ConsoleCommand

	// input history
	history []string
	hidx    int

	// backlog
	messages     []ConsoleLine
	start, count int
}

func NewConsolePanel(g *Game) *ConsolePanel {
	cp := &ConsolePanel{
		do:        make(chan func(*ConsolePanel), 100),
		BasePanel: graphics.NewPanel(),
		g:         g,
		history:   []string{""},
		messages:  make([]ConsoleLine, NLINES),
	}

	cp.prompt = "> "

	HelpCmd := ConsoleCommand{
		Name: "help",
		Desc: "get help",
		Help: "no.",
		Fn: func(args ...string) error {
			if len(args) == 0 {
				cp.AddLine(graphics.TextStyle.Fg, "help topics are:")
				cp.AddLine(graphics.TextStyle.Fg, "")
				for _, ccmd := range cp.cmds {
					cp.AddLine(graphics.TextStyle.Fg, "  %15s  -  %s", ccmd.Name, ccmd.Desc)
				}
			} else {
				if ccmd, ok := cp.cmds[args[0]]; !ok {
					return fmt.Errorf("no such command %s", args[0])
				} else {
					cp.AddLine(graphics.TextStyle.Fg, "help for %s:", args[0])
					cp.AddLine(graphics.TextStyle.Fg, "  %s", ccmd.Help)
				}
			}
			return nil
		},
	}

	ClearCmd := ConsoleCommand{
		Name: "clear",
		Desc: "clear the console",
		Help: "",
		Fn: func(args ...string) error {
			cp.do <- func(cp *ConsolePanel) {
				cp.messages = make([]ConsoleLine, NLINES)
				cp.start = 0
				cp.count = 0
			}
			return nil
		},
	}

	ExecCmd := ConsoleCommand{
		Name: "exec",
		Desc: "exec a shell command",
		Help: "usage: exec command [args...]\n\nUnfortunately, exec does not support pipes or redirections or variable expansion. (yet)",
		Fn: func(args ...string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: exec command [arguments...]")
			}
			if _, err := exec.LookPath(args[0]); err != nil {
				return fmt.Errorf("error finding program %q: %s", args[0], err)
			}
			arg0 := args[0]
			argv := args[1:]
			cmd := exec.Command(arg0, argv...)
			out, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("error executing %q %q: %s", arg0, argv, err)
			}
			cp.AddLine(graphics.TextStyle.Fg, "%s", out)
			return nil
		},
	}

	infotopics := ""
	for t, arr := range info {
		infotopics += fmt.Sprintf("  %15s - %s\n", t, arr[0])
	}

	InfoCmd := ConsoleCommand{
		Name: "info",
		Desc: "get game information",
		Help: fmt.Sprintf("available info topics:\n\n%s", infotopics),
		Fn: func(args ...string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: info topic")
			} else {
				inf := ""
				for _, l := range info[args[0]] {
					inf += fmt.Sprintf("  %s\n", l)
				}
				cp.AddLine(graphics.TextStyle.Fg, "%s info:\n%s", args[0], inf)
			}
			return nil
		},
	}

	ConfigCmd := ConsoleCommand{
		Name: "config",
		Desc: "view or modify config variables",
		Help: "run 'config get' to list all config variables",
		Fn: func(args ...string) error {
			nr := len(args)
			if nr == 0 {
				return fmt.Errorf("usage: config [get] [variable names...]")
			} else {
				switch args[0] {
				case "get":
					if nr == 2 {
						if conf, err := cp.g.config.RawGet(args[1]); err != nil {
							return fmt.Errorf("config: %s", err)
						} else {
							cp.AddLine(graphics.TextStyle.Fg, "  %10s = %v", args[1], conf)
						}
					} else {
						cp.AddLine(graphics.TextStyle.Fg, "config:")
						for ci := range cp.g.config.Chan() {
							cp.AddLine(graphics.TextStyle.Fg, "  %10s = %v", ci.Key, ci.Value)
						}
					}
				case "set":
					return fmt.Errorf("unimplemented")
				}
			}
			return nil
		},
	}

	// TODO: modify this to drop into a lua repl
	LuaCmd := ConsoleCommand{
		Name: "lua",
		Desc: "execute lua",
		Help: "usage: lua luacode...\n\nexecutes lua, must be one line",
		Fn: func(args ...string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: lua luacode...")
			}
			code := strings.Join(args, " ")
			errcode := cp.g.lua.LoadString(code)
			if errcode != 0 {
				return fmt.Errorf("lua: bad code")
			}
			ofn := luar.NewLuaObject(cp.g.lua, -1)
			if res, err := ofn.Call(); err != nil {
				return fmt.Errorf("lua: %s", err)
			} else {
				cp.AddLine(termbox.ColorBlue, "%v", res)
			}
			return nil
		},
	}

	cp.cmds = map[string]ConsoleCommand{
		HelpCmd.Name:   HelpCmd,
		ClearCmd.Name:  ClearCmd,
		ExecCmd.Name:   ExecCmd,
		InfoCmd.Name:   InfoCmd,
		ConfigCmd.Name: ConfigCmd,
		LuaCmd.Name:    LuaCmd,
	}

	// selectively enable commands
	for name, _ := range cp.cmds {
		key := "commands." + name
		if conf, err := cp.g.config.Get(key, reflect.Bool); err == nil {
			b := conf.(bool)
			if b == false {
				if glog.V(2) {
					glog.Infof("%s = %t, disabling", key, b)
				}
				delete(cp.cmds, name)
			}
		}
	}

	g.em.On("log", func(i ...interface{}) {
		glog.Info("consolepanel logging")
		cp.AddLine(termbox.ColorRed, "error: %s", i[0])
	})

	g.em.On("resize", func(i ...interface{}) {
		ev := i[0].(termbox.Event)
		cp.do <- func(cp *ConsolePanel) {
			cp.Resize(ev.Width, ev.Height)
		}
	})

	return cp
}

// turn a string into a slice of strings based on a width limit
// TODO: make utf8-safe
func wrap(str string, width int) []string {
	var out []string

	lines := strings.Split(str, "\n")

	for _, l := range lines {

		words := strings.Split(l, " ")
		if len(words) == 0 {
			return out
		}

		// current line we are making
		current := words[0]

		// # spaces left before the end
		remaining := width - len(current)

		for _, word := range words[1:] {
			if len(word)+1 > remaining {
				out = append(out, current)
				current = word
				remaining = width - len(word)
			} else {
				current += " " + word
				remaining -= 1 + len(word)
			}
		}

		out = append(out, current)
	}

	return out
}

func (cp *ConsolePanel) Draw() {
	if cp.IsActive() {
		cp.Clear()

		w := cp.Bounds().Dx()
		h := cp.Bounds().Dy()

		// draw log
		y := h - 2

		// draws multiple lines and returns how many lines the draw took
		wrappedraw := func(cl []ConsoleLine, starty int) int {
			ly := starty
			for i := len(cl) - 1; i >= 0; i-- {
				// figure out how many lines we need
				lines := wrap(cl[i].Line, w-1)
				for sl := len(lines) - 1; sl >= 0; sl-- {
					for x, r := range lines[sl] {
						cp.SetCell(x, ly, r, cl[i].Cell.Fg, cl[i].Cell.Bg)
					}
					ly--
				}
			}

			return starty - ly
		}

		if cp.start+cp.count > NLINES {

			// first part
			a1 := cp.messages[cp.start:]
			y -= wrappedraw(a1, y)

			// second
			a2 := cp.messages[:cp.start+cp.count-NLINES]
			y -= wrappedraw(a2, y)

		} else {

			// or just this
			a1 := cp.messages[cp.start : cp.start+cp.count]
			y -= wrappedraw(a1, y)

		}

		// draw prompt at bottom
		str := cp.prompt + cp.buf.String()
		for i, r := range str {
			cp.SetCell(i, h-1, r, graphics.PromptStyle.Fg, graphics.PromptStyle.Bg)
		}

		cp.Buffered.Draw()
	}
}

func (cp *ConsolePanel) Update(delta time.Duration) {
	for {
		select {
		case f := <-cp.do:
			f(cp)
		default:
			return
		}
	}
}

func (cp *ConsolePanel) HandleInput(ev termbox.Event) {
	cp.do <- func(cp *ConsolePanel) {
		switch ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				cp.buf.WriteRune(ev.Ch)
			} else {
				switch ev.Key {
				case termbox.KeySpace:
					// just add a space
					cp.buf.WriteRune(' ')

				case termbox.KeyBackspace:
					fallthrough

				case termbox.KeyBackspace2:
					// on backspace, remove the last rune in the buffer
					if cp.buf.Len() > 0 {
						_, size := utf8.DecodeLastRune(cp.buf.Bytes())
						cp.buf.Truncate(cp.buf.Len() - size)
					}

				case termbox.KeyCtrlU:
					// clear the buffer, like a UNIX terminal
					cp.buf.Reset()

				case termbox.KeyEnter:
					// input confirmed, execute
					if cp.buf.Len() > 0 {
						str := cp.buf.String()
						cp.history = append(cp.history, str)
						cp.hidx = 0
						toks := strings.Fields(str)
						cp.AddLine(graphics.PromptStyle.Fg, "%s%s", cp.prompt, str)
						if ccmd, ok := cp.cmds[toks[0]]; !ok {
							cp.AddLine(termbox.ColorRed, "not a command: %s", toks[0])
						} else {
							if err := ccmd.Fn(toks[1:]...); err != nil {
								cp.AddLine(termbox.ColorRed, "%s command error: %s", toks[0], err)
							}
						}
						cp.buf.Reset()
					}
				case termbox.KeyArrowUp:
					// get last command from history
					cp.buf.Reset()
					cp.buf.WriteString(cp.history[len(cp.history)-cp.hidx-1])
					cp.hidx = (cp.hidx + 1) % len(cp.history)

				case termbox.KeyArrowDown:
					cp.buf.Reset()

					cp.hidx--

					if cp.hidx < 0 {
						cp.hidx = len(cp.history) - 1
					}

					cp.buf.WriteString(cp.history[len(cp.history)-cp.hidx-1])

				case termbox.KeyEsc:
					// input cancelled

					//cp.buf.Reset()
					cp.g.rsys.PopInputHandler()
				}
			}

		}
	}

}

func (cp *ConsolePanel) AddLine(fg termbox.Attribute, format string, args ...interface{}) {
	cp.do <- func(cp *ConsolePanel) {
		end := (cp.start + cp.count) % NLINES
		cp.messages[end] = ConsoleLine{termbox.Cell{Fg: fg}, fmt.Sprintf(format, args...)}

		if cp.count == NLINES {
			cp.start = (cp.start + 1) % NLINES
		} else {
			cp.count++
		}
	}
}
