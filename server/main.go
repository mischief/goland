package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/mischief/goland/game/gutil"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	configfile = flag.String("config", "config.lua", "configuration file")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	defer glog.Flush()

	lua := gutil.LuaInit()
	if lua == nil {
		glog.Fatal("error initializing lua state")
	}
	// load configuration
	config, err := gutil.NewLuaConfig(lua, *configfile)
	if err != nil {
		glog.Fatalf("error loading configuration file %s: %s", *configfile, err)
	}

	/*
		// setup logging
		if lf, err := config.Get("logfile", reflect.String); err != nil {
			glog.Infof("main: Error reading logfile from config, using stdout: %s", err)
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
	*/

	// log panics
	/*
		defer func() {
			if r := recover(); r != nil {
				glog.Fatalf("recovered from %v", r)
			}
		}()*/

	glog.Infof("config loaded from %s", *configfile)

	if glog.V(2) {
		// dump config
		for ce := range config.Chan() {
			glog.Infof("config: %s -> '%s'", ce.Key, ce.Value)
		}
	}

	// enable profiling
	if cpuprof, err := config.Get("cpuprofile", reflect.String); err == nil {
		fname := cpuprof.(string)
		glog.Infof("starting profiling in file %s", fname)
		f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			glog.Infof("can't open profiling file: %s", err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gs, err := NewGameServer(config, lua)
	if err != nil {
		glog.Infoln(err)
	} else {
		gs.Run()
	}

	glog.Info("done")
}
