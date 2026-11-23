package services

import (
	"sync"
	"time"
)

type EventThrottle struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSeen map[string]time.Time
}

func NewEventThrottle(cooldown time.Duration) *EventThrottle {
	return &EventThrottle{
		cooldown: cooldown,
		lastSeen: make(map[string]time.Time),
	}
}

// Allow returns true if enough time has passed since the last accepted event
// for this (deviceID, eventType) pair, and updates the timestamp.
func (t *EventThrottle) Allow(deviceID, eventType string, now time.Time) bool {
	if t.cooldown <= 0 {
		return true
	}

	key := deviceID + ":" + eventType

	t.mu.Lock()
	defer t.mu.Unlock()

	if prev, exists := t.lastSeen[key]; exists && now.Sub(prev) < t.cooldown {
		return false
	}

	t.lastSeen[key] = now
	return true
}
