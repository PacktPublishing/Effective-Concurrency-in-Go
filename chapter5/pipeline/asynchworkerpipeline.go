package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

func workerPoolPipelineStage[IN any, OUT any](input <-chan IN, output chan<- OUT, process func(IN) OUT, numWorkers int) {
	// close output channel when all workers are done
	defer close(output)
	// Start the worker pool
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range input {
				output <- process(data)
			}
		}()
	}
	// Wait for all workers to finish
	wg.Wait()
}

func asynchronousPipeline2Workers(input *csv.Reader) {
	fmt.Println("--Asynchronous pipeline with worker pool----")
	parseInputCh := make(chan []string)
	convertInputCh := make(chan Record)
	encodeInputCh := make(chan Record)
	// We read the output of the pipeline from this channel
	outputCh := make(chan []byte)
	// We need this channel to wait for the printing of
	// the final result
	done := make(chan struct{})

	numWorkers := 2
	// Start pipeline stages and connect them
	go workerPoolPipelineStage(parseInputCh, convertInputCh, parse, numWorkers)
	go workerPoolPipelineStage(convertInputCh, encodeInputCh, convert, numWorkers)
	go workerPoolPipelineStage(encodeInputCh, outputCh, encode, numWorkers)

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
