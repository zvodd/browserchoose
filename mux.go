package browserchoose

import (
	"sync"
	"time"
)

// LockTimeout is how long a locked browser stays selected without a URL being
// routed through it. Each routed URL restarts the countdown. It is a variable
// so tests can shorten it.
var LockTimeout = 4 * time.Minute

// ChooseMode means no browser is locked: ask the user for every URL.
const ChooseMode = -1

// Mux holds the session-only browser lock. It lives in memory only and is
// never written to the config. It is safe for concurrent use.
type Mux struct {
	mu       sync.Mutex
	locked   int
	gen      uint64 // bumped on every change so stale timers are ignored
	timer    *time.Timer
	onChange func()
}

// NewMux returns a Mux in ChooseMode. onChange is called, without the lock
// held, whenever the mode changes, including when a lock times out.
func NewMux(onChange func()) *Mux {
	return &Mux{locked: ChooseMode, onChange: onChange}
}

// Pick returns the browser index that should handle the next URL, or
// ChooseMode if the user should be asked. A routed URL restarts the idle timer.
func (m *Mux) Pick() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked != ChooseMode {
		m.armLocked()
	}
	return m.locked
}

// Lock pins the mux to browser i and starts the idle timer.
func (m *Mux) Lock(i int) {
	m.mu.Lock()
	m.locked = i
	m.armLocked()
	m.mu.Unlock()
	m.notify()
}

// Unlock returns the mux to ChooseMode.
func (m *Mux) Unlock() {
	m.mu.Lock()
	m.clearLocked()
	m.mu.Unlock()
	m.notify()
}

// Mode returns the locked browser index, or ChooseMode.
func (m *Mux) Mode() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locked
}

// armLocked (re)starts the idle countdown. The caller must hold m.mu.
func (m *Mux) armLocked() {
	if m.timer != nil {
		m.timer.Stop()
	}
	m.gen++
	gen := m.gen
	m.timer = time.AfterFunc(LockTimeout, func() { m.expire(gen) })
}

// clearLocked drops the lock and cancels its timer. The caller must hold m.mu.
func (m *Mux) clearLocked() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.gen++
	m.locked = ChooseMode
}

func (m *Mux) expire(gen uint64) {
	m.mu.Lock()
	if gen != m.gen || m.locked == ChooseMode {
		m.mu.Unlock()
		return
	}
	m.clearLocked()
	m.mu.Unlock()
	m.notify()
}

func (m *Mux) notify() {
	if m.onChange != nil {
		m.onChange()
	}
}
