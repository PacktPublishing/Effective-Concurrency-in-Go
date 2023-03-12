package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type ProgressMeter struct {
	progress int64
}

func (pm *ProgressMeter) Progress() {
	atomic.AddInt64(&pm.progress, 1)
}

func (pm *ProgressMeter) Get() int64 {
	return atomic.LoadInt64(&pm.progress)
}

func longGoroutine(ctx context.Context, pm *ProgressMeter) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Canceled")
			return
		default:
		}
		time.Sleep(time.Duration(rand.Intn(120)) * time.Millisecond)
		pm.Progress()
	}
}

func observer(ctx context.Context, cancel func(), progress *ProgressMeter) {
	// Expect progress every 100 msecs. If not, cancel the goroutine
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	var lastProgress int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			p := progress.Get()
			if p == lastProgress {
				fmt.Println("No progress since last time, canceling")
				cancel()
				return
			}
			fmt.Printf("Progress: %d\n", p)
			lastProgress = p
		}
	}
}

func main() {
	var progress ProgressMeter
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		longGoroutine(ctx, &progress)
	}()
	go observer(ctx, cancel, &progress)
	wg.Wait()
}
