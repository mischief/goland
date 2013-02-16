package main

import (
	"flag"
	//"flag"
	//"fmt"
	//"time"
	//"os"

	//"github.com/nsf/termbox-go"
)

func main() {
	flag.Parse()

	g := NewGame()

	g.Run()
}
