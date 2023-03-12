package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	base := context.Background()
	c1, c1cancel := context.WithTimeout(base, 3*time.Second)
	defer c1cancel()
	c2, c2cancel := context.WithTimeout(base, 2*time.Second)
	defer c2cancel()
	c3, c3cancel := context.WithTimeout(c1, 1*time.Second)
	defer c3cancel()
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		<-c1.Done()
		fmt.Println("c1 done")
	}()
	go func() {
		defer wg.Done()
		<-c2.Done()
		fmt.Println("c2 done")
	}()
	go func() {
		defer wg.Done()
		<-c3.Done()
		fmt.Println("c3 done")
	}()
	wg.Wait()
}
