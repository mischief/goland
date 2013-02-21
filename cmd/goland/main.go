package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()

	// log panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from %v", r)
		}
	}()

	g := NewGame()

	defer g.End()

	// do the good stuff
	g.Run()
}

