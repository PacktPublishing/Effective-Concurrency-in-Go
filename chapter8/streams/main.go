package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"

	_ "modernc.org/sqlite"

	"streams/filters"
	"streams/store"
)

func initDB(dbName string) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		panic(err)
	}
	// Create the timeseries table
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(`create table if not exists measurements(at integer, value double)`)
	if err != nil {
		panic(err)
	}
	// Fill it with some data
	result, err := tx.Query(`select count(*) from measurements`)
	if err != nil {
		panic(err)
	}
	result.Next()
	var nItems int
	if err := result.Scan(&nItems); err != nil {
		panic(err)
	}
	tx.Commit()
	if nItems < 10000 {
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		fmt.Printf("nRows: %d, inserting data...\n", nItems)
		tm := time.Now().UnixMilli()
		stmt, err := tx.Prepare(`insert into measurements(at,value) values(?,?)`)
		if err != nil {
			panic(err)
		}
		for i := 0; i < 10000; i++ {
			if _, err := stmt.Exec(tm, rand.Float64()); err != nil {
				panic(err)
			}
			tm -= 100
		}
		tx.Commit()
	}
}

func simpleMain() {
	initDB("test.db")
	// Start database
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}

	// Create the store
	st := store.Store{DB: db}

	// Stream results
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entries, err := st.Stream(ctx, store.Request{})
	if err != nil {
		panic(err)
	}
	filteredEntries := filters.MinFilter(0.001, entries)
	entryCh, errCh := filters.ErrFilter(filteredEntries)
	resultCh := filters.MovingAvg(0.5, 5, entryCh)
	var streamErr error
	go func() {
		for err := range errCh {
			// Capture first error
			if streamErr == nil {
				streamErr = err
				cancel()
			}
		}
	}()
	for entry := range resultCh {
		fmt.Printf("%+v\n", entry)
	}
	if streamErr != nil {
		fmt.Println(streamErr)
	}
}

type Message struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
	Error string    `json:"err"`
}

func DecodeToChan[T any](decode func(*T) error) (<-chan T, <-chan error) {
	ret := make(chan T)
	errch := make(chan error, 1)
	go func() {
		defer close(ret)
		defer close(errch)
		var entry T
		for {
			if err := decode(&entry); err != nil {
				if !errors.Is(err, io.EOF) {
					errch <- err
				}
				return
			}
			ret <- entry
		}
	}()
	return ret, errch
}

func EncodeFromChan[T any](input <-chan T, encode func(T) ([]byte, error), out io.Writer) <-chan error {
	ret := make(chan error, 1)
	go func() {
		defer close(ret)
		for entry := range input {
			data, err := encode(entry)
			if err != nil {
				ret <- err
				return
			}
			if _, err := out.Write(data); err != nil {
				if !errors.Is(err, io.EOF) {
					ret <- err
				}
				return
			}
		}
	}()
	return ret
}

func wsMain() {
	initDB("test.db")
	// Start database
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}

	// Create the store
	st := store.Store{DB: db}

	// Create a websocket server
	http.Handle("/db", websocket.Handler(func(conn *websocket.Conn) {
		data, err := st.Stream(conn.Request().Context(), store.Request{})
		if err != nil {
			fmt.Println("Store error", err)
			if err != nil {
				return
			}
		}

		errCh := EncodeFromChan(data, func(entry store.Entry) ([]byte, error) {
			msg := Message{
				At:    entry.At,
				Value: entry.Value,
			}
			if entry.Error != nil {
				msg.Error = entry.Error.Error()
			}
			return json.Marshal(msg)
		}, conn)
		err = <-errCh
		if err != nil {
			fmt.Println("Encode error", err)
		}
	}))
	go func() {
		fmt.Println("Server started at :10001")
		if err := http.ListenAndServe(":10001", nil); err != nil {
			panic(err)
		}
	}()

	// Create a client
	cli, err := websocket.Dial("ws://localhost:10001/db", "", "http://localhost:10001")
	if err != nil {
		panic(err)
	}
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	decoder := json.NewDecoder(cli)
	entries, rcvErr := DecodeToChan[store.Entry](func(entry *store.Entry) error {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return err
		}
		entry.At = msg.At
		entry.Value = msg.Value
		if msg.Error != "" {
			entry.Error = fmt.Errorf(msg.Error)
		}

		return nil
	})

	filteredEntries := filters.MinFilter(0.001, entries)
	entryCh, errCh := filters.ErrFilter(filteredEntries)
	resultCh := filters.MovingAvg(0.5, 5, entryCh)

	go func() {
		for err := range errCh {
			fmt.Println("Stream error", err)
		}
	}()
	for entry := range resultCh {
		fmt.Printf("%+v\n", entry)
	}
	err = <-rcvErr
	if err != nil {
		fmt.Println("Receive error", err)
	}
}

func httpMain() {
	initDB("test.db")
	// Start database
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}

	// Create the store
	st := store.Store{DB: db}

	// Create an HTTP server
	http.HandleFunc("/db", func(w http.ResponseWriter, req *http.Request) {
		data, err := st.Stream(req.Context(), store.Request{})
		if err != nil {
			fmt.Println("Store error", err)
			if err != nil {
				return
			}
		}

		errCh := EncodeFromChan(data, func(entry store.Entry) ([]byte, error) {
			msg := Message{
				At:    entry.At,
				Value: entry.Value,
			}
			if entry.Error != nil {
				msg.Error = entry.Error.Error()
			}
			return json.Marshal(msg)
		}, w)
		err = <-errCh
		if err != nil {
			fmt.Println("Encode error", err)
		}
	})
	go func() {
		fmt.Println("Server started at :10001")
		if err := http.ListenAndServe(":10001", nil); err != nil {
			panic(err)
		}
	}()

	// Create a client
	resp, err := http.Get("http://localhost:10001/db")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	entries, rcvErr := DecodeToChan[store.Entry](func(entry *store.Entry) error {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return err
		}
		entry.At = msg.At
		entry.Value = msg.Value
		if msg.Error != "" {
			entry.Error = fmt.Errorf(msg.Error)
		}

		return nil
	})

	filteredEntries := filters.MinFilter(0.001, entries)
	entryCh, errCh := filters.ErrFilter(filteredEntries)
	resultCh := filters.MovingAvg(0.5, 5, entryCh)

	go func() {
		for err := range errCh {
			fmt.Println("Stream error", err)
		}
	}()
	for entry := range resultCh {
		fmt.Printf("%+v\n", entry)
	}
	err = <-rcvErr
	if err != nil {
		fmt.Println("Receive error", err)
	}
}

func main() {
	// Uncomment one of the implementations below, and comment out the others
	//simpleMain()
	//wsMain()
	httpMain()
}
