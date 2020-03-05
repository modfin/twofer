package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

func New(tpm uint) *Ratelimiter {
	r := &Ratelimiter{
		tpm:    tpm,
		ledger: map[string]*item{},
	}

	go r.start()

	return r
}

type item struct {
	start time.Time
	count uint
}

type Ratelimiter struct {
	tpm    uint // transaction per minute
	ledger map[string]*item
	mu     sync.Mutex
}

func (r *Ratelimiter) start() {
	for {
		<-time.After(time.Minute)
		fmt.Println("Cleaning")
		c := r.clean()
		if c > 0 {
			fmt.Println("Cleaner: cleared", c, "from rate limit")
		}
	}
}

func (r *Ratelimiter) clean() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for k, v := range r.ledger {
		if v == nil {
			delete(r.ledger, k)
			continue
		}
		if time.Since(v.start) > time.Minute {
			delete(r.ledger, k)
			count++
			continue
		}
	}
	return count
}

func (r *Ratelimiter) Hit(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.ledger[key]
	if i == nil || time.Since(i.start) > time.Minute {
		i = &item{
			start: time.Now(),
			count: 0,
		}
		r.ledger[key] = i
	}
	i.count += 1

	if r.tpm < i.count {
		return fmt.Errorf("rate limit for uri is reached, max %d transactions/min", r.tpm)
	}
	return nil
}
