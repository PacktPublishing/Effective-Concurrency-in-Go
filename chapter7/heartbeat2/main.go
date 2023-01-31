package main

import (
	"fmt"
	"math/rand"
	"time"
)

func longRunningProcess(heartbeat, done chan struct{}) {
	for {
		n := time.Millisecond * time.Duration(rand.Intn(500)-350)
		// Do something for a long time
		time.Sleep(time.Second + n)
		select {
		case <-done:
			return
		case heartbeat <- struct{}{}:
		default:
		}
	}
}

func monitor(heartbeat, done chan struct{}, tick <-chan time.Time) {
	var numTicks int
	var lastHeartbeat time.Time
	for {
		select {
		case <-tick:

			numTicks++
			if numTicks >= 2 {
				fmt.Printf("No progress since %s, terminating\n", lastHeartbeat)
				close(done)
				return
			} else {
				fmt.Printf("Tick\n")
			}

		case <-heartbeat:
			lastHeartbeat = time.Now()
			numTicks = 0
			fmt.Printf("Heartbeat received %s\n", lastHeartbeat)
		}
	}
}

func main() {
	heartbeat := make(chan struct{})
	done := make(chan struct{})
	// Expect a heartbeat at least every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	go longRunningProcess(heartbeat, done)
	go monitor(heartbeat, done, ticker.C)
	<-done
}
