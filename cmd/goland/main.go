package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()

	// log panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from %v", r)
		}
	}()

	// enable profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	g := NewGame()

	defer g.End()

	// do the good stuff
	g.Run()
}
