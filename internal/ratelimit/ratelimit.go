// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package ratelimit

import (
	"strings"
	"time"
)

// Counter is Redis-backed in production, so a limit holds across instances.
type Counter interface {
	Incr(key string, ttl time.Duration) (int64, error)
	Delete(key string) error
}

type Limiter struct {
	counter  Counter
	scope    string
	attempts int64
	window   time.Duration
}

func New(counter Counter, scope string, attempts int64, window time.Duration) Limiter {
	return Limiter{
		counter:  counter,
		scope:    scope,
		attempts: attempts,
		window:   window,
	}
}

func (l Limiter) Allow(subject ...string) (bool, error) {
	count, err := l.counter.Incr(l.key(subject), l.window)
	if err != nil {
		return false, err
	}

	return count <= l.attempts, nil
}

func (l Limiter) Reset(subject ...string) error {
	return l.counter.Delete(l.key(subject))
}

func (l Limiter) key(subject []string) string {
	return l.scope + ":" + strings.Join(subject, ":")
}
