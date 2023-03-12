package main

import (
	"fmt"
	"math/rand"
	"time"
)

func longRunning(done, heartBeat chan struct{}) {
	for {
		select {
		case <-done:
			return
		case heartBeat <- struct{}{}:
		default:
		}
		time.Sleep(100 * time.Millisecond)
		// Sometimes goroutine fails
		if rand.Intn(100) > 96 {
			fmt.Println("Func fails")
			return
		}
	}
}

func restart(done chan struct{}, f func(done, heartBeat chan struct{}), timeout time.Duration) {
	for {
		funcDone := make(chan struct{})
		heartBeat := make(chan struct{})
		fmt.Println("starting func")
		go func() {
			f(funcDone, heartBeat)
		}()
		timer := time.NewTimer(timeout)
		retry := false
		for !retry {
			select {
			case <-done:
				close(funcDone)
				return
			case <-heartBeat:
				fmt.Println("Heartbeat")
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				fmt.Println("Timeout, stopping func")
				close(funcDone)
				retry = true
			}
		}
	}
}

func main() {
	done := make(chan struct{})
	restart(done, longRunning, 1000*time.Millisecond)
}
