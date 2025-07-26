package timingwheel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimingWheel_AddTimer(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	durations := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	for _, d := range durations {
		t.Run("", func(t *testing.T) {
			exitC := make(chan time.Time)

			start := time.Now().UTC()
			tw.AddTimer(d, func() {
				exitC <- time.Now().UTC()
			})

			got := (<-exitC).Truncate(time.Millisecond)
			min := start.Add(d).Truncate(time.Millisecond)

			err := 10 * time.Millisecond // Allow for some timing variance
			if got.Before(min) || got.After(min.Add(err)) {
				t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", d, min, min.Add(err), got)
			}
		})
	}
}

func TestTimingWheel_MultipleTimers(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	var executed int32
	var wg sync.WaitGroup

	// Add multiple timers with different delays
	for i := 0; i < 10; i++ {
		wg.Add(1)
		delay := time.Duration(i+1) * 10 * time.Millisecond
		tw.AddTimer(delay, func() {
			atomic.AddInt32(&executed, 1)
			wg.Done()
		})
	}

	// Wait for all timers to execute
	wg.Wait()

	if got := atomic.LoadInt32(&executed); got != 10 {
		t.Errorf("Expected 10 timers to execute, got %d", got)
	}
}

func TestTimingWheel_ConcurrentAddTimer(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	var executed int32
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Add timers concurrently from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tw.AddTimer(10*time.Millisecond, func() {
				atomic.AddInt32(&executed, 1)
			})
		}()
	}

	wg.Wait()

	// Wait for all timers to execute
	time.Sleep(50 * time.Millisecond)

	if got := atomic.LoadInt32(&executed); got != numGoroutines {
		t.Errorf("Expected %d timers to execute, got %d", numGoroutines, got)
	}
}

func TestTimingWheel_ZeroDelay(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	executed := make(chan bool, 1)
	start := time.Now()

	tw.AddTimer(0, func() {
		executed <- true
	})

	select {
	case <-executed:
		elapsed := time.Since(start)
		// Should execute within one tick (plus some margin)
		if elapsed > 5*time.Millisecond {
			t.Errorf("Zero delay timer took too long: %v", elapsed)
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("Zero delay timer did not execute within timeout")
	}
}

func TestTimingWheel_LongDelay(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	executed := make(chan bool, 1)
	start := time.Now()
	delay := 100 * time.Millisecond // This will require multiple rounds

	tw.AddTimer(delay, func() {
		executed <- true
	})

	select {
	case <-executed:
		elapsed := time.Since(start)
		// Should execute close to the expected delay
		if elapsed < delay || elapsed > delay+10*time.Millisecond {
			t.Errorf("Long delay timer execution time: want ~%v, got %v", delay, elapsed)
		}
	case <-time.After(delay + 20*time.Millisecond):
		t.Error("Long delay timer did not execute within timeout")
	}
}

func TestTimingWheel_StartStop(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()
	
	executed := make(chan bool, 1)
	tw.AddTimer(5*time.Millisecond, func() {
		executed <- true
	})
	
	select {
	case <-executed:
		// Timer executed successfully
	case <-time.After(20 * time.Millisecond):
		t.Error("Timer did not execute after start")
	}
}

func TestTimingWheel_StopBeforeExecution(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	
	executed := make(chan bool, 1)
	tw.AddTimer(20*time.Millisecond, func() {
		executed <- true
	})
	
	// Stop the wheel before timer should execute
	tw.Stop()
	
	select {
	case <-executed:
		t.Error("Timer executed after wheel was stopped")
	case <-time.After(30 * time.Millisecond):
		// Expected: timer should not execute
	}
}

func TestTimingWheel_TaskExecution(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	results := make(chan int, 5)
	
	// Add timers that will execute in order
	for i := 0; i < 5; i++ {
		value := i
		delay := time.Duration(i+1) * 5 * time.Millisecond
		tw.AddTimer(delay, func() {
			results <- value
		})
	}
	
	// Collect results
	var executed []int
	for i := 0; i < 5; i++ {
		select {
		case val := <-results:
			executed = append(executed, val)
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Timer did not execute within timeout")
		}
	}
	
	// Results should be in order (0, 1, 2, 3, 4)
	for i, val := range executed {
		if val != i {
			t.Errorf("Expected timer %d to execute, got %d", i, val)
		}
	}
}

func TestTimingWheel_RoundCalculation(t *testing.T) {
	tickDuration := 10 * time.Millisecond
	wheelSize := 10
	tw := NewTimingWheel(tickDuration, wheelSize)
	tw.Start()
	defer tw.Stop()

	// Test timer that requires multiple rounds
	delay := time.Duration(wheelSize) * tickDuration * 2 // 2 full rounds
	
	executed := make(chan bool, 1)
	start := time.Now()
	
	tw.AddTimer(delay, func() {
		executed <- true
	})
	
	select {
	case <-executed:
		elapsed := time.Since(start)
		// Should execute after approximately 2 full rounds
		expectedMin := delay
		expectedMax := delay + 20*time.Millisecond
		if elapsed < expectedMin || elapsed > expectedMax {
			t.Errorf("Round calculation timing: want [%v, %v], got %v", expectedMin, expectedMax, elapsed)
		}
	case <-time.After(delay + 50*time.Millisecond):
		t.Error("Multi-round timer did not execute within timeout")
	}
}

func BenchmarkTimingWheel_AddTimer(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 1000)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.AddTimer(time.Duration(i%100)*time.Millisecond, func() {})
	}
}

func BenchmarkTimingWheel_ConcurrentAddTimer(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 1000)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tw.AddTimer(time.Duration(i%100)*time.Millisecond, func() {})
			i++
		}
	})
}