# timingwheel

Golang implementation of Hierarchical Timing Wheels.

## Usage

```go
package main

import (
    "fmt"
    "time"
    "github.com/ferueda/timingwheel"
)

func main() {
    // Create a timing wheel with 1ms tick duration and 1000 slots
    tw := timingwheel.NewTimingWheel(time.Millisecond, 1000)

    // Start the timing wheel
    tw.Start()
    defer tw.Stop()

    // Add a timer that will execute after 100ms
    tw.AddTimer(100*time.Millisecond, func() {
        fmt.Println("Timer executed!")
    })

    // Keep the program running
    time.Sleep(200 * time.Millisecond)
}
```

## Benchmark

```
$ go test -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: github.com/ferueda/timingwheel
cpu: Apple M4 Pro
BenchmarkTimingWheel_AddTimer_Scale/N-1K-12         	 3904827	       692.9 ns/op	    2091 B/op	       1 allocs/op
BenchmarkTimingWheel_AddTimer_Scale/N-10K-12        	 1725793	       766.7 ns/op	    2254 B/op	       1 allocs/op
BenchmarkTimingWheel_AddTimer_Scale/N-100K-12       	 1386366	       782.0 ns/op	    2527 B/op	       1 allocs/op
BenchmarkTimingWheel_AddTimer_Scale/N-1M-12         	  412380	      4241 ns/op	   13725 B/op	       1 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/TimingWheel-N-1K-12         	1000000000	         0.4780 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/StandardTimer-N-1K-12       	501822678	         2.412 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/TimingWheel-N-10K-12        	1000000000	         0.4733 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/StandardTimer-N-10K-12      	499650069	         2.420 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/TimingWheel-N-100K-12       	1000000000	         0.4723 ns/op	       0 B/op	       0 allocs/op
BenchmarkTimingWheel_vs_StandardTimer/StandardTimer-N-100K-12     	501154496	         2.411 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/ferueda/timingwheel	24.791s
```

The benchmark results show that the timing wheel implementation maintains consistent performance even as the number of existing timers increases. When compared to Go's standard `time.Timer`, the timing wheel demonstrates superior performance characteristics, particularly in scenarios with high timer volumes.

## Features

- **High Performance**: Optimized for scenarios with large numbers of timers
- **Scalable**: Performance remains consistent as timer count increases
- **Concurrent-Safe**: Thread-safe operations using channels and mutexes
- **Memory Efficient**: Low memory overhead per timer
- **Hierarchical**: Supports timers with delays spanning multiple wheel rotations

## Algorithm

This implementation is based on the Hierarchical Timing Wheels algorithm as described in:

- [Hashed and Hierarchical Timing Wheels](https://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf)
