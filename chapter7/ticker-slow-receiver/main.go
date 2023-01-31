package main

import (
	"fmt"
	"time"
)

func main() {
	// Ticker ticks every 100 msec
	dur := 100 * time.Millisecond
	start := time.Now()
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	n := 0
	go func() {
		for {
			<-ticker.C
			// Receiver works for 50 msecs, but every 5 ticks, it works for
			// 180 msecs
			n++
			var sleep time.Duration
			if n >= 5 {
				sleep = 180 * time.Millisecond
				n = 0
			} else {
				sleep = 10 * time.Millisecond
			}
			fmt.Printf("Tick at %d, delaying for %d msecs\n", time.Since(start).Milliseconds(), sleep.Milliseconds())
			time.Sleep(sleep)
		}
	}()
	select {}
}
