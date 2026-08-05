package search

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// httpStatusError is returned for non-2xx retailer responses so callers can
// distinguish rate-limits (429) from hard blocks (403) and open the breaker.
type httpStatusError struct {
	Code int
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("status %d", e.Code)
}

func IsRateLimited(err error) bool {
	var se *httpStatusError
	if errors.As(err, &se) && se.Code == 429 {
		return true
	}
	return err != nil && strings.Contains(err.Error(), "status 429")
}

func IsBlocked(err error) bool {
	var se *httpStatusError
	if errors.As(err, &se) && (se.Code == 403 || se.Code == 401) {
		return true
	}
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status 403") || strings.Contains(msg, "access denied")
}

type circuitBreaker struct {
	mu        sync.Mutex
	openUntil map[string]time.Time
	fails     map[string]int
	threshold int
	cooldown  time.Duration
}

func newCircuitBreaker(threshold int, cooldown time.Duration) *circuitBreaker {
	if threshold <= 0 {
		threshold = 2
	}
	if cooldown <= 0 {
		cooldown = 2 * time.Minute
	}
	return &circuitBreaker{
		openUntil: make(map[string]time.Time),
		fails:     make(map[string]int),
		threshold: threshold,
		cooldown:  cooldown,
	}
}

func (b *circuitBreaker) Allow(name string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	until, ok := b.openUntil[name]
	if !ok {
		return true
	}
	if time.Now().After(until) {
		delete(b.openUntil, name)
		b.fails[name] = 0
		return true
	}
	return false
}

func (b *circuitBreaker) Success(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.fails[name] = 0
	delete(b.openUntil, name)
}

func (b *circuitBreaker) Failure(name string, err error) {
	if err == nil {
		return
	}
	// Only trip on rate-limit / bot-block — empty catalogs are not outages.
	if !IsRateLimited(err) && !IsBlocked(err) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.fails[name]++
	if b.fails[name] >= b.threshold {
		b.openUntil[name] = time.Now().Add(b.cooldown)
		b.fails[name] = 0
	}
}

var ErrCircuitOpen = errors.New("circuit open (recently rate-limited)")
