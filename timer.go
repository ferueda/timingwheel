package timingwheel

import (
	"time"
)

// Timer represents a single delayed task. When it expires, the task will be
// executed.
type Timer struct {
	Delay time.Duration // how long to wait before execution
	Round int           // how many full rotations left
	Slot  int           // which slot it's placed in
	Task  func()        // the actual function to run
}
