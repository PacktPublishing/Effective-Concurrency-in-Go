package main

import (
	"fmt"
	"time"
)

func main() {
	// timer will be used to cancel work after 100 msec
	timer := time.NewTimer(100 * time.Millisecond)
	// Close the timeout channel after 100 msec
	timeout := make(chan struct{})
	go func() {
		<-timer.C
		close(timeout)
		fmt.Println("Timeout")
	}()
	// A more convenient way would have been:
	// time.AfterFunc(100*time.Millisecond,func() {close(timeout)})
	// Do some work until it times out
	x := 0
	done := false
	for !done {
		// Check if timed out
		select {
		case <-timeout:
			done = true
		default:
		}
		time.Sleep(time.Millisecond)
		x++
	}
	fmt.Println(x)
}
