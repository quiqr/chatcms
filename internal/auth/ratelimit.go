package auth

import (
	"errors"
	"sync"
	"time"
)

// ErrTooManyAttempts says guessing has been going on and further attempts
// are refused for a while. It never says whether the address exists.
var ErrTooManyAttempts = errors.New("too many attempts; wait a few minutes and try again")

// How much guessing is allowed before it is slowed down. Two buckets, because
// limiting only by address lets one caller work through a list of people, and
// limiting only by caller lets a botnet work on one person.
const (
	AttemptsPerAddress = 10
	AttemptsPerCaller  = 30
	AttemptWindow      = 15 * time.Minute
)

type bucket struct {
	count int
	until time.Time
}

// Limiter counts attempts in this process. The deployment model is one
// server, so this is correct today; the design records what changes when
// there are two.
type Limiter struct {
	mu      sync.Mutex
	byKey   map[string]*bucket
	now     func() time.Time
	window  time.Duration
	allowed map[string]int
}

func NewLimiter() *Limiter {
	return &Limiter{
		byKey:  map[string]*bucket{},
		now:    time.Now,
		window: AttemptWindow,
		allowed: map[string]int{
			"address": AttemptsPerAddress,
			"caller":  AttemptsPerCaller,
		},
	}
}

// Allow reports whether another attempt may be made for this address from
// this caller.
func (l *Limiter) Allow(address, caller string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, pair := range []struct{ kind, value string }{
		{"address", address},
		{"caller", caller},
	} {
		if pair.value == "" {
			continue
		}
		if l.exhausted(pair.kind, pair.value) {
			return ErrTooManyAttempts
		}
	}
	return nil
}

func (l *Limiter) exhausted(kind, value string) bool {
	key := kind + ":" + value
	entry, ok := l.byKey[key]
	if !ok {
		return false
	}
	if l.now().After(entry.until) {
		delete(l.byKey, key)
		return false
	}
	return entry.count >= l.allowed[kind]
}

// RecordFailure counts an attempt that did not work.
func (l *Limiter) RecordFailure(address, caller string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, pair := range []struct{ kind, value string }{
		{"address", address},
		{"caller", caller},
	} {
		if pair.value == "" {
			continue
		}
		key := pair.kind + ":" + pair.value
		entry, ok := l.byKey[key]
		if !ok || l.now().After(entry.until) {
			entry = &bucket{until: l.now().Add(l.window)}
			l.byKey[key] = entry
		}
		entry.count++
	}
}

// RecordSuccess clears what was counted against an address, so that someone
// who mistyped twice and then got it right is not punished.
func (l *Limiter) RecordSuccess(address, caller string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.byKey, "address:"+address)
	delete(l.byKey, "caller:"+caller)
}
