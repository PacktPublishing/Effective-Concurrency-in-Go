package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Race-free use of atomic as a synchronization tool. The number of
// times the program will print 1 will be different at each run. It
// will never print 0.
func main() {
	for i := 0; i < 1000000; i++ {
		var done atomic.Bool
		var a int
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a = 1
			done.Store(true)
		}()
		if done.Load() {
			fmt.Println(a)
		}
		wg.Wait()
	}
}
