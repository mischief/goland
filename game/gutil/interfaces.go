// This file describes general interfaces that game objects
// should try to take advantage of.
package gutil

import (
	"time"
)

type Updater interface {
	Update(delta time.Duration)
}
