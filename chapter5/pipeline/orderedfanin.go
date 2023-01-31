package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

type sequenced interface {
	getSequence() int
}

type fanInRecord[T sequenced] struct {
	index int
	data  T
	pause chan struct{}
}

func orderedFanIn[T sequenced](done <-chan struct{}, channels ...<-chan T) <-chan T {
	fanInCh := make(chan fanInRecord[T])
	wg := sync.WaitGroup{}
	for i := range channels {
		pauseCh := make(chan struct{})
		wg.Add(1)
		go func(index int, pause chan struct{}) {
			defer wg.Done()
			for {
				var ok bool
				var data T
				select {
				case data, ok = <-channels[index]:
					if !ok {
						return
					}
					fanInCh <- fanInRecord[T]{
						index: index,
						data:  data,
						pause: pause,
					}
				case <-done:
					return
				}
				select {
				case <-pause:
				case <-done:
					return
				}
			}
		}(i, pauseCh)
	}
	go func() {
		wg.Wait()
		close(fanInCh)
	}()
	outputCh := make(chan T)
	go func() {
		defer close(outputCh)
		// The next record expected
		expected := 1
		queuedData := make([]*fanInRecord[T], len(channels))
		for in := range fanInCh {
			// If this input is what is expected, send it to the output
			if in.data.getSequence() == expected {
				select {
				case outputCh <- in.data:
					in.pause <- struct{}{}
					expected++
					allDone := false
					// Send all queued data
					for !allDone {
						allDone = true
						for i, d := range queuedData {
							if d != nil && d.data.getSequence() == expected {
								select {
								case outputCh <- d.data:
									queuedData[i] = nil
									d.pause <- struct{}{}
									expected++
									allDone = false
								case <-done:
									return
								}
							}
						}
					}
				case <-done:
					return
				}
			} else {
				// This is out-of-order, queue it
				in := in
				queuedData[in.index] = &in
			}
		}
	}()
	return outputCh
}
func orderedFanOutFanIn(input *csv.Reader) {
	fmt.Println("--Ordered Fan-Out - Fan-In----")

	done := make(chan struct{})

	// single input channel to the parse stage
	parseInputCh := make(chan []string)
	convertInputCh := cancelablePipelineStage(parseInputCh, done, parse)

	numWorkers := 2
	fanInChannels := make([]<-chan Record, 0)
	for i := 0; i < numWorkers; i++ {
		// Fan-out: multiple workers read from convertInputCh
		convertOutputCh := cancelablePipelineStage(convertInputCh, done, convert)
		fanInChannels = append(fanInChannels, convertOutputCh)
	}
	convertOutputCh := orderedFanIn(done, fanInChannels...)
	outputCh := cancelablePipelineStage(convertOutputCh, done, encode)
	// Start a goroutine to read pipeline output
	go func() {
		for data := range outputCh {
			fmt.Println(string(data))
		}
		close(done)
	}()

	// Ignore the first row
	input.Read()
	for {
		rec, err := input.Read()
		if err == io.EOF {
			close(parseInputCh)
			break
		}
		if err != nil {
			panic(err)
		}
		// Send input to pipeline
		parseInputCh <- rec
	}
	// Wait until the last output is printed
	<-done
}
