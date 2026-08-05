package search

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerTripsOn429(t *testing.T) {
	b := newCircuitBreaker(2, time.Minute)
	name := "snitch"
	if !b.Allow(name) {
		t.Fatal("expected allow initially")
	}
	err429 := &httpStatusError{Code: 429}
	b.Failure(name, err429)
	if !b.Allow(name) {
		t.Fatal("expected allow after 1 failure")
	}
	b.Failure(name, err429)
	if b.Allow(name) {
		t.Fatal("expected open after threshold")
	}
	if !errors.Is(ErrCircuitOpen, ErrCircuitOpen) {
		t.Fatal("sentinel")
	}
}

func TestCircuitBreakerIgnoresEmptyCatalog(t *testing.T) {
	b := newCircuitBreaker(1, time.Minute)
	b.Failure("snitch", errors.New("snitch: no products for \"Shirt\""))
	if !b.Allow("snitch") {
		t.Fatal("empty catalog should not trip breaker")
	}
}

func TestIsRateLimited(t *testing.T) {
	if !IsRateLimited(&httpStatusError{Code: 429}) {
		t.Fatal("expected 429")
	}
	if IsRateLimited(&httpStatusError{Code: 500}) {
		t.Fatal("500 is not rate limit")
	}
}
