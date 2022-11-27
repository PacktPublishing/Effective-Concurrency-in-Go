package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

func cancelablePipelineStage[IN any, OUT any](input <-chan IN, done <-chan struct{}, process func(IN) OUT) <-chan OUT {
	outputCh := make(chan OUT)
	go func() {
		for {
			select {
			case data, ok := <-input:
				if !ok {
					close(outputCh)
					return
				}
				outputCh <- process(data)
			case <-done:
				return
			}
		}
	}()
	return outputCh
}

func fanIn[T any](done <-chan struct{}, channels ...<-chan T) <-chan T {
	outputCh := make(chan T)
	wg := sync.WaitGroup{}
	for _, ch := range channels {
		wg.Add(1)
		go func(input <-chan T) {
			defer wg.Done()
			for {
				select {
				case data, ok := <-input:
					if !ok {
						return
					}
					outputCh <- data
				case <-done:
					return
				}
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(outputCh)
	}()
	return outputCh
}

func fanOutFanIn(input *csv.Reader) {
	fmt.Println("--Fan-Out - Fan-In----")

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
	convertOutputCh := fanIn(done, fanInChannels...)
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
