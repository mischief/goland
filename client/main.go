package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	logfile    = flag.String("log", "goland.log", "log file")
	debug      = flag.Bool("debug", false, "print debugging info")
)

func main() {
	flag.Parse()

	f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	log.SetOutput(f)
	log.Print("Logging started")

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

		log.Println("Starting profiling in file %s", *cpuprofile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.Println("Creating game instance")
	g := NewGame()

	// do the good stuff
	log.Println("Beginning game loop")
	g.Run()

	log.Println("Done")
}
