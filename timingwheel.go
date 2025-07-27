// Package timingwheel provides a timing wheel implementation for efficient timer management.
package timingwheel

import (
	"sync"
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

// TimingWheel is an implementation of Hierarchical Timing Wheels.
type TimingWheel struct {
	tickDuration time.Duration // duration of each tick
	slots        [][]*Timer    // wheel slots
	currentPos   int           // current tick position
	wheelSize    int           // number of slots
	addTimerCh   chan *Timer   // channel for incoming timers
	stopChan     chan struct{} // channel to stop the wheel
	mu           sync.RWMutex  // protect concurrent access
}

// NewTimingWheel creates a new TimingWheel with the given tick duration and wheel size.
func NewTimingWheel(tickDuration time.Duration, wheelSize int) *TimingWheel {
	slots := make([][]*Timer, wheelSize)
	for i := range slots {
		slots[i] = []*Timer{}
	}
	return &TimingWheel{
		tickDuration: tickDuration,
		slots:        slots,
		wheelSize:    wheelSize,
		addTimerCh:   make(chan *Timer),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the TimingWheel.
func (tw *TimingWheel) Start() {
	ticker := time.NewTicker(tw.tickDuration)
	go func() {
		for {
			select {
			case <-ticker.C:
				tw.handleTick()
			case timer := <-tw.addTimerCh:
				tw.addTimer(timer)
			case <-tw.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the TimingWheel.
func (tw *TimingWheel) Stop() {
	close(tw.stopChan)
}

// AddTimer adds a new timer to the TimingWheel.
func (tw *TimingWheel) AddTimer(delay time.Duration, task func()) {
	timer := &Timer{
		Delay: delay,
		Task:  task,
	}
	tw.addTimerCh <- timer
}

func (tw *TimingWheel) addTimer(timer *Timer) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	ticks := int(timer.Delay / tw.tickDuration)
	if ticks == 0 {
		ticks = 1
	}

	slot := (tw.currentPos + ticks) % tw.wheelSize
	round := ticks / tw.wheelSize

	timer.Slot = slot
	timer.Round = round

	tw.slots[slot] = append(tw.slots[slot], timer)
}

func (tw *TimingWheel) handleTick() {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	slot := tw.slots[tw.currentPos]
	var remainingTimers []*Timer

	for _, timer := range slot {
		if timer.Round <= 0 {
			// Timer has expired, execute the task
			go timer.Task()
		} else {
			// Still has rounds to go, decrement and keep it
			timer.Round--
			remainingTimers = append(remainingTimers, timer)
		}
	}
	// Replace current slot with remaining timers
	tw.slots[tw.currentPos] = remainingTimers
	// Move to the next slot
	tw.currentPos = (tw.currentPos + 1) % tw.wheelSize
}
