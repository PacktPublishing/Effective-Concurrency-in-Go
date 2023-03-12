package filters

import (
	"streams/store"
)

type AboveThresholdEntry struct {
	store.Entry
	Avg float64
}

func MovingAvg(threshold float64, windowSize int, in <-chan store.Entry) <-chan AboveThresholdEntry {
	// A channel can be used as a circular/FIFO buffer
	window := make(chan float64, windowSize)
	out := make(chan AboveThresholdEntry)
	go func() {
		defer close(out)
		var runningTotal float64
		for input := range in {
			if len(window) == windowSize {
				avg := runningTotal / float64(windowSize)
				if avg > threshold {
					out <- AboveThresholdEntry{
						Entry: input,
						Avg:   avg,
					}
				}
				// Drop the last in window
				runningTotal -= <-window
			}
			// Add value to window
			window <- input.Value
			runningTotal += input.Value
		}
	}()
	return out
}
