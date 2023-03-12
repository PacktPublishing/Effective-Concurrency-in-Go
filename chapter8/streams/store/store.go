package store

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	DB *sql.DB
}

type Request struct {
	// This would include the query parameters
}

type Entry struct {
	At    time.Time
	Value float64
	Error error
}

func (svc Store) Stream(ctx context.Context, req Request) (<-chan Entry, error) {
	rows, err := svc.DB.Query(`select at,value from measurements`)
	if err != nil {
		return nil, err
	}
	ret := make(chan Entry)
	go func() {
		// Close the channel to notify the receiver that data stream is
		// finished
		defer close(ret)
		defer rows.Close()
		for {
			var at int64
			var entry Entry
			select {
			case <-ctx.Done():
				return
			default:
			}
			if !rows.Next() {
				break
			}
			if err := rows.Scan(&at, &entry.Value); err != nil {
				ret <- Entry{Error: err}
				continue
			}
			entry.At = time.UnixMilli(at)
			ret <- entry
		}
		if err := rows.Err(); err != nil {
			ret <- Entry{Error: err}
		}
	}()
	return ret, nil
}
