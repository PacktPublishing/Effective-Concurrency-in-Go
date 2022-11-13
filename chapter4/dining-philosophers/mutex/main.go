package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func philosopher(index int, leftFork, rightFork *sync.Mutex) {
	for {
		// Think for some time
		fmt.Printf("Philospher %d is thinking\n", index)
		time.Sleep(time.Duration(rand.Intn(1000)))
		// Get left fork
		leftFork.Lock()
		fmt.Printf("Philosopher %d got left fork\n", index)
		// Get right fork
		if rightFork.TryLock() {
			fmt.Printf("Philosopher %d got right fork\n", index)
			// Eat
			fmt.Printf("Philosopher %d is eating\n", index)
			time.Sleep(time.Duration(rand.Intn(1000)))
			rightFork.Unlock()
		}
		leftFork.Unlock()
	}
}

func main() {
	forks := [5]sync.Mutex{}
	go philosopher(0, &forks[4], &forks[0])
	go philosopher(1, &forks[0], &forks[1])
	go philosopher(2, &forks[1], &forks[2])
	go philosopher(3, &forks[2], &forks[3])
	go philosopher(4, &forks[3], &forks[4])
	select {}
}
