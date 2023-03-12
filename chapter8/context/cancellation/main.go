package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	go func() {
		<-ctx.Done()
		fmt.Println("Empty context canceled")
	}()
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()
	go func() {
		<-ctx1.Done()
		fmt.Println("ctx1 canceled")
	}()
	ctx2, cancel2 := context.WithCancel(ctx1)
	defer cancel2()
	go func() {
		<-ctx2.Done()
		fmt.Println("ctx2 canceled")
	}()
	ctx3, cancel3 := context.WithCancel(ctx1)
	defer cancel3()
	go func() {
		<-ctx3.Done()
		fmt.Println("ctx3 canceled")
	}()

	ctx4, cancel4 := context.WithCancel(ctx2)
	defer cancel4()
	go func() {
		<-ctx4.Done()
		fmt.Println("ctx4 canceled")
	}()
	time.Sleep(time.Second)
	cancel1()
	time.Sleep(time.Second)
}
