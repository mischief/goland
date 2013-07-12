package main

import (
	"flag"
	"github.com/aarzilli/golua/lua"
	"github.com/mischief/goland/game/gutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	configfile = flag.String("config", "config.lua", "configuration file")

	Lua *lua.State
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	//	Lua = lua.NewState()
	Lua = gutil.LuaInit()
}

func main() {
	flag.Parse()

	// load configuration
	config, err := gutil.NewLuaConfig(Lua, *configfile)
	if err != nil {
		log.Fatalf("Error loading configuration file %s: %s", *configfile, err)
	}

	// setup logging
	if lf, err := config.Get("logfile", reflect.String); err != nil {
		log.Printf("main: Error reading logfile from config, using stdout: %s", err)
	} else {
		// open log file
		file := lf.(string)
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Print("main: Logging started")

	// log panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("main: Recovered from %v", r)
		}
	}()

	log.Printf("main: Config loaded from %s", *configfile)

	// dump config
	for ce := range config.Chan() {
		log.Printf("main: config: %s -> '%s'", ce.Key, ce.Value)
	}

	// enable profiling
	if cpuprof, err := config.Get("cpuprofile", reflect.String); err == nil {
		fname := cpuprof.(string)
		log.Printf("main: Starting profiling in file %s", fname)
		f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("main: Can't open profiling file: %s", err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gs, err := NewGameServer(config, Lua)
	if err != nil {
		log.Println(err)
	} else {
		gs.Run()
	}

	log.Println("main: Logging ended")
}
