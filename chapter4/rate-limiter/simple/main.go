package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Limiter struct {
	mu sync.Mutex
	// Bucket is filled with rate tokens per second
	rate int
	// Bucket size
	bucketSize int
	// Number of tokens in bucket
	nTokens int
	// Time last token was generated
	lastToken time.Time
}

func (s *Limiter) Wait() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nTokens > 0 {
		s.nTokens--
		return
	}
	// Here, there is not enough tokens in the bucket
	tElapsed := time.Since(s.lastToken)
	period := time.Second / time.Duration(s.rate)
	nTokens := tElapsed.Nanoseconds() / period.Nanoseconds()
	s.nTokens = int(nTokens)
	if s.nTokens > s.bucketSize {
		s.nTokens = s.bucketSize
	}
	s.lastToken = s.lastToken.Add(time.Duration(nTokens) * period)
	// We filled the bucket. There may not be enough
	if s.nTokens > 0 {
		s.nTokens--
		return
	}
	// We have to wait until more tokens are available
	// A token should be available at:
	next := s.lastToken.Add(period)
	wait := next.Sub(time.Now())
	if wait >= 0 {
		time.Sleep(wait)
	}
	s.lastToken = next
}

func NewLimiter(rate int, limit int) *Limiter {
	return &Limiter{
		rate:       rate,
		bucketSize: limit,
		nTokens:    limit,
		lastToken:  time.Now(),
	}
}

func main() {
	limiter := NewLimiter(5, 10)

	for i := 0; i < 100; i++ {
		limiter.Wait()
		fmt.Printf("Request: %v %+v\n", time.Now(), limiter)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(400)))
	}
	time.Sleep(time.Second * 2)
	for i := 0; i < 100; i++ {
		limiter.Wait()
		fmt.Printf("Request: %v %+v\n", time.Now(), limiter)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(400)))
	}

}
