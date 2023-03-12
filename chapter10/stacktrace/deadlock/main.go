package main

import (
	"sync"
)

func philosopher(firstFork, secondFork *sync.Mutex) {
	for {
		firstFork.Lock()
		secondFork.Lock()
		secondFork.Unlock()
		firstFork.Unlock()
	}
}

func main() {
	forks := [2]sync.Mutex{}
	go philosopher(&forks[1], &forks[0])
	go philosopher(&forks[0], &forks[1])
	select {}
}
