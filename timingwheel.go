// Package timingwheel provides a timing wheel implementation for efficient timer management.
package timingwheel

import (
	"sync"
	"time"
)

type TimingWheel struct {
	tickDuration time.Duration // duration of each tick
	slots        [][]*Timer    // wheel slots
	currentPos   int           // current tick position
	wheelSize    int           // number of slots
	addTimerCh   chan *Timer   // channel for incoming timers
	stopChan     chan struct{} // channel to stop the wheel
	mu           sync.RWMutex  // protect concurrent access
}

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
