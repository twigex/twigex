// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package ratelimit

import (
	"errors"
	"testing"
	"time"
)

type fakeCounter struct {
	counts map[string]int64
	ttls   map[string]time.Duration
	err    error
}

func newFakeCounter() *fakeCounter {
	return &fakeCounter{
		counts: map[string]int64{},
		ttls:   map[string]time.Duration{},
	}
}

func (f *fakeCounter) Incr(key string, ttl time.Duration) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.counts[key]++
	f.ttls[key] = ttl
	return f.counts[key], nil
}

func (f *fakeCounter) Delete(key string) error {
	if f.err != nil {
		return f.err
	}
	delete(f.counts, key)
	return nil
}

func TestAllowUpToTheLimit(t *testing.T) {
	counter := newFakeCounter()
	limiter := New(counter, "login_attempt", 3, time.Minute)

	for i := 1; i <= 3; i++ {
		allowed, err := limiter.Allow("10.0.0.1", "someone")
		if err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
		if !allowed {
			t.Errorf("attempt %d should be allowed", i)
		}
	}

	allowed, err := limiter.Allow("10.0.0.1", "someone")
	if err != nil {
		t.Fatalf("attempt 4: %v", err)
	}
	if allowed {
		t.Error("attempt past the limit should be refused")
	}
}

func TestResetClearsTheCount(t *testing.T) {
	counter := newFakeCounter()
	limiter := New(counter, "login_attempt", 1, time.Minute)

	if _, err := limiter.Allow("10.0.0.1", "someone"); err != nil {
		t.Fatal(err)
	}
	if allowed, _ := limiter.Allow("10.0.0.1", "someone"); allowed {
		t.Fatal("second attempt should be refused before the reset")
	}

	if err := limiter.Reset("10.0.0.1", "someone"); err != nil {
		t.Fatal(err)
	}

	if allowed, _ := limiter.Allow("10.0.0.1", "someone"); !allowed {
		t.Error("the reset should give the subject its budget back")
	}
}

func TestSubjectsAndScopesAreCountedApart(t *testing.T) {
	counter := newFakeCounter()
	login := New(counter, "login_attempt", 1, time.Minute)
	share := New(counter, "share_auth", 1, time.Minute)

	if _, err := login.Allow("10.0.0.1", "someone"); err != nil {
		t.Fatal(err)
	}

	if allowed, _ := login.Allow("10.0.0.2", "someone"); !allowed {
		t.Error("another address must not spend the first one's budget")
	}
	if allowed, _ := login.Allow("10.0.0.1", "someone-else"); !allowed {
		t.Error("another subject must not spend the first one's budget")
	}
	if allowed, _ := share.Allow("10.0.0.1", "someone"); !allowed {
		t.Error("another scope must not spend the first one's budget")
	}
}

func TestKeyAndWindowReachTheCounter(t *testing.T) {
	counter := newFakeCounter()
	limiter := New(counter, "login_attempt", 5, 5*time.Minute)

	if _, err := limiter.Allow("10.0.0.1", "someone"); err != nil {
		t.Fatal(err)
	}

	const want = "login_attempt:10.0.0.1:someone"
	if _, ok := counter.counts[want]; !ok {
		t.Errorf("want the key %s, got %v", want, counter.counts)
	}
	if got := counter.ttls[want]; got != 5*time.Minute {
		t.Errorf("want a 5m window, got %s", got)
	}
}

func TestCounterFailureRefusesTheAttempt(t *testing.T) {
	counter := newFakeCounter()
	counter.err = errors.New("redis is down")
	limiter := New(counter, "login_attempt", 5, time.Minute)

	allowed, err := limiter.Allow("10.0.0.1", "someone")
	if err == nil {
		t.Fatal("the counter error must reach the caller")
	}
	if allowed {
		t.Error("a failed count must not report the attempt as allowed")
	}
}
