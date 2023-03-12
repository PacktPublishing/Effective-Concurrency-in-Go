package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func CancelSupport() (cancel func(), isCancelled func() bool) {
	v := atomic.Bool{}
	cancel = func() {
		v.Store(true)
	}
	isCancelled = func() bool {
		return v.Load()
	}
	return
}

func main() {
	cancel, isCanceled := CancelSupport()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(100 * time.Millisecond)
			if isCanceled() {
				fmt.Println("Cancelled")
				return
			}
		}
	}()
	time.AfterFunc(5*time.Second, cancel)
	wg.Wait()
}
