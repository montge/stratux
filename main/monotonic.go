/*
	Copyright (c) 2015-2016 Christopher Young
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	Modifications (c) 2016 AvSquirrel (https://github.com/AvSquirrel)
	monotonic.go: Create monotonic clock using time.Timer - necessary because of real time clock changes on RPi.
*/

package main

import (
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
)

// Timer (since start).

type monotonic struct {
	Milliseconds uint64
	Time         time.Time
	ticker       *time.Ticker
	realTimeSet  bool
	RealTime     time.Time
	mu           sync.RWMutex // Protects all fields from concurrent access
}

func (m *monotonic) Watcher() {
	for {
		<-m.ticker.C
		m.mu.Lock()
		m.Milliseconds += 10
		m.Time = m.Time.Add(10 * time.Millisecond)
		if m.realTimeSet {
			m.RealTime = m.RealTime.Add(10 * time.Millisecond)
		}
		m.mu.Unlock()
	}
}

func (m *monotonic) Since(t time.Time) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Time.Sub(t)
}

func (m *monotonic) HumanizeTime(t time.Time) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return humanize.RelTime(t, m.Time, "ago", "from now")
}

func (m *monotonic) Unix() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(m.Time.Sub(time.Time{}).Seconds())
}

func (m *monotonic) HasRealTimeReference() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.realTimeSet
}

func (m *monotonic) SetRealTimeReference(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.realTimeSet { // Only allow the real clock to be set once.
		m.RealTime = t
		m.realTimeSet = true
	}
}

// GetMilliseconds returns the current milliseconds count in a thread-safe manner.
func (m *monotonic) GetMilliseconds() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Milliseconds
}

// GetTime returns the current monotonic time in a thread-safe manner.
func (m *monotonic) GetTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Time
}

// GetRealTime returns the current real time reference in a thread-safe manner.
func (m *monotonic) GetRealTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.RealTime
}

func NewMonotonic() *monotonic {
	t := &monotonic{Milliseconds: 0, Time: time.Now(), ticker: time.NewTicker(10 * time.Millisecond)}
	go t.Watcher()
	return t
}
