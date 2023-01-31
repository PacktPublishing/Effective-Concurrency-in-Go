package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	done := time.After(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("Tick: %d\n", time.Since(start).Milliseconds())
		case <-done:
			return
		}
	}
}
