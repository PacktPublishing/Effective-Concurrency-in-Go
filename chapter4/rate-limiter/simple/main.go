package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Limiter struct {
	sync.Mutex
	// Bucket is filled with rate tokens per second
	rate float64
	// Bucket size
	bucketSize int
	// Number of tokens in bucket
	nTokens int
	// Time last token was generated
	lastToken time.Time
}

func (s *Limiter) Wait() {
	s.Lock()
	if s.nTokens > 0 {
		s.nTokens--
		s.Unlock()
		return
	}
	// Here, there is not enough tokens in the bucket
	tElapsed := time.Since(s.lastToken)
	period := time.Duration(1.0 / float64(s.rate) * 1000000000)
	nTokens := int(tElapsed.Nanoseconds() / period.Nanoseconds())
	s.nTokens = nTokens
	if s.nTokens > s.bucketSize {
		s.nTokens = s.bucketSize
	}
	s.lastToken = s.lastToken.Add(time.Duration(nTokens) * period)
	// We filled the bucket. There may not be enough
	if s.nTokens > 0 {
		s.nTokens--
		s.Unlock()
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
	s.Unlock()
}

func NewLimiter(rate float64, limit int) *Limiter {
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
