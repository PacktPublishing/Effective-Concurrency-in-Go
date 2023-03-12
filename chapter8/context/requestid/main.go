package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type requestIDKeyType int

var requestIDKey requestIDKeyType

func WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIDKey, uuid.New())
}

func GetRequestID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(requestIDKey).(uuid.UUID)
	return id
}

func main() {
	ctx := context.Background()
	ctx1 := WithRequestID(ctx)
	ctx2 := WithRequestID(ctx1)
	fmt.Println(GetRequestID(ctx), GetRequestID(ctx1), GetRequestID(ctx2))
}
