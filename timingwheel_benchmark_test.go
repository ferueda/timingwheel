package timingwheel

import (
	"sync"
	"testing"
	"time"
)

func genD(i int) time.Duration {
	return time.Duration(i%10000) * time.Millisecond
}

func BenchmarkTimingWheel_AddTimer_Scale(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	cases := []struct {
		name string
		N    int // the data size (i.e. number of existing timers)
	}{
		{"N-1K", 1000},
		{"N-10K", 10000},
		{"N-100K", 100000},
		{"N-1M", 1000000},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			// Create baseline timers
			for i := 0; i < c.N; i++ {
				tw.AddTimer(genD(i), func() {})
			}
			
			b.ResetTimer()
			
			// Benchmark adding new timers
			for i := 0; i < b.N; i++ {
				tw.AddTimer(time.Second, func() {})
			}
		})
	}
}

func BenchmarkStandardTimer_Scale(b *testing.B) {
	cases := []struct {
		name string
		N    int // the data size (i.e. number of existing timers)
	}{
		{"N-1K", 1000},
		{"N-10K", 10000},
		{"N-100K", 100000},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			base := make([]*time.Timer, c.N)
			for i := 0; i < len(base); i++ {
				base[i] = time.AfterFunc(genD(i), func() {})
			}
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				timer := time.AfterFunc(time.Second, func() {})
				timer.Stop()
			}
			
			b.StopTimer()
			for i := 0; i < len(base); i++ {
				base[i].Stop()
			}
		})
	}
}

func BenchmarkTimingWheel_AddTimer_Concurrent(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 1000)
	tw.Start()
	defer tw.Stop()

	cases := []struct {
		name       string
		goroutines int
	}{
		{"Goroutines-1", 1},
		{"Goroutines-10", 10},
		{"Goroutines-100", 100},
		{"Goroutines-1000", 1000},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			var wg sync.WaitGroup
			
			b.ResetTimer()
			
			for g := 0; g < c.goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N/c.goroutines; i++ {
						tw.AddTimer(time.Duration(i%1000)*time.Millisecond, func() {})
					}
				}()
			}
			
			wg.Wait()
		})
	}
}

func BenchmarkTimingWheel_AddTimer_DifferentDelays(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 1000)
	tw.Start()
	defer tw.Stop()

	cases := []struct {
		name  string
		delay time.Duration
	}{
		{"Delay-1ms", 1 * time.Millisecond},
		{"Delay-10ms", 10 * time.Millisecond},
		{"Delay-100ms", 100 * time.Millisecond},
		{"Delay-1s", 1 * time.Second},
		{"Delay-10s", 10 * time.Second},
		{"Delay-1m", 1 * time.Minute},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				tw.AddTimer(c.delay, func() {})
			}
		})
	}
}

func BenchmarkTimingWheel_Memory_Usage(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 1000)
	tw.Start()
	defer tw.Stop()

	cases := []struct {
		name string
		N    int
	}{
		{"Timers-1K", 1000},
		{"Timers-10K", 10000},
		{"Timers-100K", 100000},
		{"Timers-1M", 1000000},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Create many timers to test memory usage
				for j := 0; j < c.N/b.N; j++ {
					tw.AddTimer(time.Duration(j)*time.Millisecond, func() {})
				}
			}
		})
	}
}

func BenchmarkTimingWheel_TickProcessing(b *testing.B) {
	cases := []struct {
		name         string
		tickDuration time.Duration
		wheelSize    int
	}{
		{"Tick-1ms-WS-100", 1 * time.Millisecond, 100},
		{"Tick-1ms-WS-1000", 1 * time.Millisecond, 1000},
		{"Tick-10ms-WS-100", 10 * time.Millisecond, 100},
		{"Tick-10ms-WS-1000", 10 * time.Millisecond, 1000},
		{"Tick-100ms-WS-100", 100 * time.Millisecond, 100},
		{"Tick-100ms-WS-1000", 100 * time.Millisecond, 1000},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			tw := NewTimingWheel(c.tickDuration, c.wheelSize)
			tw.Start()
			defer tw.Stop()

			// Add some timers to process
			for i := 0; i < 1000; i++ {
				tw.AddTimer(time.Duration(i)*c.tickDuration, func() {})
			}

			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				tw.AddTimer(c.tickDuration, func() {})
			}
		})
	}
}

func BenchmarkTimingWheel_vs_StandardTimer(b *testing.B) {
	cases := []struct {
		name string
		N    int
	}{
		{"N-1K", 1000},
		{"N-10K", 10000},
		{"N-100K", 100000},
	}

	for _, c := range cases {
		b.Run("TimingWheel-"+c.name, func(b *testing.B) {
			tw := NewTimingWheel(time.Millisecond, 1000)
			tw.Start()
			defer tw.Stop()

			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for j := 0; j < c.N/b.N; j++ {
					tw.AddTimer(time.Duration(j)*time.Millisecond, func() {})
				}
			}
		})

		b.Run("StandardTimer-"+c.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				timers := make([]*time.Timer, c.N/b.N)
				for j := 0; j < c.N/b.N; j++ {
					timers[j] = time.AfterFunc(time.Duration(j)*time.Millisecond, func() {})
				}
				// Stop timers to clean up
				for _, timer := range timers {
					timer.Stop()
				}
			}
		})
	}
}

func BenchmarkTimingWheel_RoundRobin(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 100)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Test round-robin distribution across slots
		delay := time.Duration(i%100) * time.Millisecond
		tw.AddTimer(delay, func() {})
	}
}

func BenchmarkTimingWheel_LongDelays(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 100)
	tw.Start()
	defer tw.Stop()

	cases := []struct {
		name  string
		delay time.Duration
	}{
		{"Delay-1Round", 100 * time.Millisecond},   // 1 full round
		{"Delay-10Rounds", 1000 * time.Millisecond}, // 10 full rounds
		{"Delay-100Rounds", 10000 * time.Millisecond}, // 100 full rounds
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				tw.AddTimer(c.delay, func() {})
			}
		})
	}
}

func BenchmarkTimingWheel_TaskExecution(b *testing.B) {
	tw := NewTimingWheel(time.Millisecond, 100)
	tw.Start()
	defer tw.Stop()

	var counter int64
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tw.AddTimer(time.Millisecond, func() {
			counter++
		})
	}
	
	// Wait for tasks to complete
	time.Sleep(10 * time.Millisecond)
}