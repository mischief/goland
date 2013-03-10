// Time wrapper for easy delta timers
package game

import (
	"time"
)

type DeltaTimer struct {
	t time.Time
}

func NewDeltaTimer() *DeltaTimer {
	return &DeltaTimer{t: time.Now()}
}

// Reset resets sets the time of the timer to now.
func (dt *DeltaTimer) Reset() {
	dt.t = time.Now()
}

// DeltaTime returns the time between now and the last time DeltaTime was called.
// If DeltaTime is called the first time, the last time is the time DeltaTimer was created.
// DeltaTime also resets the timer.
func (dt *DeltaTimer) DeltaTime() time.Duration {
	now := time.Now()
	deltaTime := now.Sub(dt.t)
	dt.t = now
	return deltaTime
}

// GetDeltaTime returns the time between now and the last time DeltaTime was called.
// If DeltaTime is called the first time, the last time is the time DeltaTimer was created.
// GetDeltaTime does not reset the timer.
func (dt *DeltaTimer) GetDeltaTime() time.Duration {
	return time.Since(dt.t)
}
