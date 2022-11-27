package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
)

type Record struct {
	Row    int     `json:"row"`
	Height float64 `json:"height"`
	Weight float64 `json:"weight"`
}

func newRecord(in []string) (rec Record, err error) {
	rec.Row, err = strconv.Atoi(in[0])
	if err != nil {
		return
	}
	rec.Height, err = strconv.ParseFloat(in[1], 64)
	if err != nil {
		return
	}
	rec.Weight, err = strconv.ParseFloat(in[2], 64)
	return
}

func parse(input []string) Record {
	rec, err := newRecord(input)
	if err != nil {
		panic(err)
	}
	return rec
}

func convert(input Record) Record {
	input.Height = 2.54 * input.Height
	input.Weight = 0.454 * input.Weight
	return input
}

func encode(input Record) []byte {
	data, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return data
}

func (r Record) getIndex() int { return r.Row }

func main() {
	for _, pipeline := range []func(*csv.Reader){
		synchronousPipeline,
		asynchronousPipeline,
		asynchronousPipeline2Workers,
		fanOutFanIn,
		orderedFanOutFanIn,
	} {
		input, err := os.Open("sample.csv")
		if err != nil {
			panic(err)
		}
		reader := csv.NewReader(input)
		pipeline(reader)
		input.Close()
	}
}
