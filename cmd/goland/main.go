package main

import (
	"flag"
  "log"
	//"flag"
	//"time"
	//"os"

	//"github.com/nsf/termbox-go"
)

func main() {
	flag.Parse()

	g := NewGame()

  defer func() {
    if r := recover(); r != nil {
      log.Printf("Recovered from %v", r)
      g.End()
    }
  }()

	g.Run()
}

