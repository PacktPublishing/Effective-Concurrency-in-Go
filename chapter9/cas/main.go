package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type SomeStruct struct {
	v int
}

var sharedValue atomic.Pointer[SomeStruct]

func computeNewCopy(in SomeStruct) SomeStruct {
	return SomeStruct{v: in.v + 1}
}

func updateSharedValue(index int) {
	myCopy := sharedValue.Load()
	newCopy := computeNewCopy(*myCopy)
	if sharedValue.CompareAndSwap(myCopy, &newCopy) {
		fmt.Printf("Set value %d\n", index)
	} else {
		fmt.Printf("Cannot set value %d\n", index)
	}
}

func main() {
	sharedValue.Store(&SomeStruct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			updateSharedValue(i)
		}()
	}
	wg.Done()
}
