package browserchoose

import (
	"sync/atomic"
	"testing"
	"time"
)

func withLockTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	old := LockTimeout
	LockTimeout = d
	t.Cleanup(func() { LockTimeout = old })
}

func TestMuxStartsInChooseMode(t *testing.T) {
	m := NewMux(nil)
	if got := m.Pick(); got != ChooseMode {
		t.Fatalf("Pick() = %d, want ChooseMode", got)
	}
}

func TestMuxLockAndUnlock(t *testing.T) {
	withLockTimeout(t, time.Hour)
	var changes atomic.Int32
	m := NewMux(func() { changes.Add(1) })

	m.Lock(2)
	if got := m.Pick(); got != 2 {
		t.Fatalf("Pick() after Lock(2) = %d, want 2", got)
	}

	m.Unlock()
	if got := m.Pick(); got != ChooseMode {
		t.Fatalf("Pick() after Unlock = %d, want ChooseMode", got)
	}
	if n := changes.Load(); n != 2 {
		t.Fatalf("onChange called %d times, want 2 (lock and unlock)", n)
	}
}

func TestMuxLockExpiresWhenIdle(t *testing.T) {
	withLockTimeout(t, 30*time.Millisecond)
	m := NewMux(nil)
	m.Lock(1)

	deadline := time.Now().Add(2 * time.Second)
	for m.Mode() != ChooseMode {
		if time.Now().After(deadline) {
			t.Fatal("lock did not expire after the idle timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestMuxRoutedURLKeepsLockAlive(t *testing.T) {
	withLockTimeout(t, 80*time.Millisecond)
	m := NewMux(nil)
	m.Lock(0)

	// Keep routing URLs well inside the timeout; the lock must survive
	// past the original deadline.
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		if got := m.Pick(); got != 0 {
			t.Fatalf("iteration %d: Pick() = %d, want 0 (lock should be kept alive)", i, got)
		}
	}
}

func TestMuxStaleTimerDoesNotDropNewLock(t *testing.T) {
	withLockTimeout(t, 100*time.Millisecond)
	m := NewMux(nil)
	m.Lock(0)
	time.Sleep(60 * time.Millisecond)
	m.Lock(1) // re-arms: deadline moves to ~160ms

	// The first lock's timer fires at ~100ms. It must not clear the second lock.
	time.Sleep(80 * time.Millisecond)
	if got := m.Mode(); got != 1 {
		t.Fatalf("Mode() = %d at ~140ms, want 1: stale timer dropped the new lock", got)
	}

	deadline := time.Now().Add(2 * time.Second)
	for m.Mode() != ChooseMode {
		if time.Now().After(deadline) {
			t.Fatal("second lock never expired")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
