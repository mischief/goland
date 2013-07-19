package game

import (
	"fmt"
	"runtime"
	"time"
)

type System interface {
	DoOne() error
	Syn()
	Running(bool)
	IsRunning() bool
	Stop()
	Frequency() time.Duration
	Setup() error
	Tick(time.Duration)
	TearDown()
}

func StartSystem(sys System, lockOsThread bool) (err error) {
	if sys.IsRunning() {
		return fmt.Errorf("startsystem: system %s already running", sys)
	} else {
		go func() {
			if lockOsThread {
				runtime.LockOSThread()
			}
			if e := sys.Setup(); e != nil {
				err = e
				return
			}

			sys.Running(true)
			for sys.IsRunning() {
				sys.DoOne()
			}

			sys.TearDown()
		}()

		sys.Syn()

		ticker := time.Tick(sys.Frequency())
		timer := NewDeltaTimer()

		go func() {
			for sys.IsRunning() {
				<-ticker
				sys.Tick(timer.DeltaTime())
			}
		}()
	}

	return nil
}
