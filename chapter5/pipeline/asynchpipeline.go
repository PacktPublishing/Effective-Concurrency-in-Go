package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

func pipelineStage[IN any, OUT any](input <-chan IN, output chan<- OUT, process func(IN) OUT) {
	defer close(output)
	for data := range input {
		output <- process(data)
	}
}

func asynchronousPipeline(input *csv.Reader) {
	fmt.Println("--Asynchronous pipeline----")
	parseInputCh := make(chan []string)
	convertInputCh := make(chan Record)
	encodeInputCh := make(chan Record)
	// We read the output of the pipeline from this channel
	outputCh := make(chan []byte)
	// We need this channel to wait for the printing of
	// the final result
	done := make(chan struct{})

	// Start pipeline stages and connect them
	go pipelineStage(parseInputCh, convertInputCh, parse)
	go pipelineStage(convertInputCh, encodeInputCh, convert)
	go pipelineStage(encodeInputCh, outputCh, encode)

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
