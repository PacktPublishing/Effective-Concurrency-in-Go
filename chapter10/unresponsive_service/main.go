package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

var ErrBusy = errors.New("Busy")
var ErrTimeout = errors.New("Timeout")

type Monitor[Req, Rsp any] struct {
	CallTimeout  time.Duration
	AlertTimeout time.Duration
	Alert        chan struct{}
	SlowFunc     func(Req) (Rsp, error)
	Done         chan struct{}

	active    chan struct{}
	full      chan struct{}
	heartBeat chan struct{}
}

type Request struct{}
type Response struct{}

func NewMonitor[Req, Rsp any](callTimeout time.Duration,
	alertTimeout time.Duration,
	maxActive int,
	call func(Req) (Rsp, error)) *Monitor[Req, Rsp] {
	mon := &Monitor[Req, Rsp]{
		CallTimeout:  callTimeout,
		AlertTimeout: alertTimeout,
		SlowFunc:     call,
		Alert:        make(chan struct{}, 1),
		active:       make(chan struct{}, maxActive),
		Done:         make(chan struct{}),
		full:         make(chan struct{}),
		heartBeat:    make(chan struct{}),
	}

	go func() {
		var timer *time.Timer
		for {
			if timer == nil {
				select {
				case <-mon.full:
					timer = time.NewTimer(mon.AlertTimeout)
				case <-mon.Done:
					return
				}
			} else {
				select {
				case <-timer.C:
					mon.Alert <- struct{}{}
				case <-mon.heartBeat:
					if !timer.Stop() {
						<-timer.C
					}
				case <-mon.Done:
					return
				}
				timer = nil
			}
		}
	}()

	return mon
}

func (mon *Monitor[Req, Rsp]) Close() {
	close(mon.Done)
}

func SlowFunc(req *Request) (*Response, error) {
	k := rand.Intn(100)
	if k == 98 {
		fmt.Println("This call will hang!")
		select {}
	}
	if k > 85 {
		time.Sleep(10 * time.Second)
	}
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	return &Response{}, nil
}

func (mon *Monitor[Req, Rsp]) Call(ctx context.Context, req Req) (Rsp, error) {
	var (
		rsp Rsp
		err error
	)
	// If the monitor cannot accept a new call, return ErrBusy immediately, but also start the alert timer
	select {
	case mon.active <- struct{}{}:
	default:
		// Start the alert timer
		select {
		case mon.active <- struct{}{}:
		case mon.full <- struct{}{}:
			return rsp, ErrBusy
		default:
			return rsp, ErrBusy
		}
	}
	// Call the function in a separate goroutine
	complete := make(chan struct{})
	go func() {
		defer func() {
			// Notify the monitor that the function returned
			<-mon.active
			select {
			case mon.heartBeat <- struct{}{}:
			default:
			}
			// Notify the caller that the call is complete
			close(complete)
		}()
		rsp, err = mon.SlowFunc(req)
	}()
	select {
	case <-time.After(mon.CallTimeout):
		return rsp, ErrTimeout
	case <-complete:
		return rsp, err
	}
}

func main() {
	mon := NewMonitor[*Request, *Response](50*time.Millisecond, 5*time.Second, 10, SlowFunc)
	go func() {
		select {
		case <-mon.Alert:
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		case <-mon.Done:
			return
		}
	}()
	for i := 0; i < 5; i++ {
		go func() {
			for {
				_, err := mon.Call(context.Background(), &Request{})
				if err == nil {
					fmt.Println(len(mon.active), err)
				}
			}
		}()
	}
	select {}
}
