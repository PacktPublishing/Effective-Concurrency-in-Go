package chat

import (
	"errors"
	"io"
	"time"
)

type Message struct {
	Timestamp time.Time
	Message   string
	From      string
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
