package e2e

import (
	"errors"
	"testing"
)

func TestNonceSequencerMonotonic(t *testing.T) {
	s, err := newNonceSequencer()
	if err != nil {
		t.Fatal(err)
	}
	first, _ := s.next()
	second, _ := s.next()

	// Prefix is stable, counter advances.
	for i := 0; i < 4; i++ {
		if first[i] != second[i] {
			t.Fatalf("prefix changed between frames at byte %d", i)
		}
	}
	var a, b [NonceSize]byte
	copy(a[:], first)
	copy(b[:], second)
	if !greater(b, a) {
		t.Fatal("counter did not advance")
	}
}

func TestReplayGuardRejectsReorderAndReplay(t *testing.T) {
	s, _ := newNonceSequencer()
	n0, _ := s.next()
	n1, _ := s.next()
	n2, _ := s.next()

	var g replayGuard
	if err := g.accept(n0); err != nil {
		t.Fatalf("first frame rejected: %v", err)
	}
	if err := g.accept(n2); err != nil {
		t.Fatalf("advancing frame rejected: %v", err)
	}
	// n1 arrives late - must be rejected as a reorder.
	if err := g.accept(n1); !errors.Is(err, ErrReplay) {
		t.Fatalf("reorder accepted, got %v", err)
	}
	// Replaying n2 must be rejected.
	if err := g.accept(n2); !errors.Is(err, ErrReplay) {
		t.Fatalf("replay accepted, got %v", err)
	}
}
