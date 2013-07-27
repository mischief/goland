package game

import (
	"fmt"
	"runtime"
	"time"
)

// The system interface is for parts of the game which wish to implement logic
// for the properties of actors.
type System interface {
  // Return a static name for this system
  String() string

  // Execute one pending task
	DoOne() error

  // Syn waits until the System has finished all tasks queed before calling Syn
	Syn()

  // Running will set the run state of the System
	Running(bool)

  // IsRunning checks if the System is running
	IsRunning() bool

  // Stop should halt the System
	Stop()

  // Frequency at which Tick is called
	Frequency() time.Duration

  // Setup any relevant data for the System
	Setup() error

  // Tick queues any regularly scheduled tasks for the System
  // at a rate of Frequency
	Tick(time.Duration)

  // TearDown should cleanup any resources the System uses
	TearDown()
}

// StartSystem starts and supervises the execution of a System.
// if lockOsThread is true, the System will be launched in a goroutine
// with runtime.LockOSThread() called.
func StartSystem(sys System, lockOsThread bool) (err error) {
	if sys.IsRunning() {
		return fmt.Errorf("startsystem: system %s already running", sys)
	} else {
    errch := make(chan error)
		go func() {
			if lockOsThread {
				runtime.LockOSThread()
			}

			sys.Running(true)

			if e := sys.Setup(); e != nil {
        errch <- e
				return
			}

      errch <- nil

			for sys.IsRunning() {
				sys.DoOne()
			}

			sys.TearDown()
		}()

    if e := <-errch; e != nil {
      err = e
      return
    }

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
