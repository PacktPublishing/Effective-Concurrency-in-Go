package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type Work struct {
	file    string
	pattern *regexp.Regexp
	result  chan Result
}

type Result struct {
	file       string
	lineNumber int
	text       string
}

func worker(jobs chan Work) {
	for work := range jobs {
		f, err := os.Open(work.file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		scn := bufio.NewScanner(f)
		lineNumber := 1
		for scn.Scan() {
			result := work.pattern.Find(scn.Bytes())
			if len(result) > 0 {
				work.result <- Result{
					file:       work.file,
					lineNumber: lineNumber,
					text:       string(result),
				}
			}
			lineNumber++
		}
		close(work.result)
		f.Close()
	}
}

func main() {
	jobs := make(chan Work)
	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(jobs)
		}()
	}
	var err error
	rex, err := regexp.Compile(os.Args[2])
	if err != nil {
		panic(err)
	}

	allResults := make(chan chan Result)
	go func() {
		defer close(allResults)
		filepath.Walk(os.Args[1], func(path string, d fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				ch := make(chan Result)
				jobs <- Work{file: path, pattern: rex, result: ch}
				allResults <- ch
			}
			return nil
		})
	}()
	for resultCh := range allResults {
		for result := range resultCh {
			fmt.Printf("%s #%d: %s\n", result.file, result.lineNumber, result.text)
		}
	}

	close(jobs)
	wg.Wait()
}
