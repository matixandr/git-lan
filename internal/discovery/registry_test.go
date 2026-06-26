package discovery

import (
	"testing"
	"time"
)

func TestRegistryTTLExpiry(t *testing.T) {
	clock := time.Now()
	r := NewRegistry()
	r.now = func() time.Time { return clock }

	r.Upsert(Peer{Instance: "maciek-laptop"})
	if got := r.List(); len(got) != 1 {
		t.Fatalf("expected 1 live peer, got %d", len(got))
	}

	clock = clock.Add(PeerTTL + time.Second)
	if got := r.List(); len(got) != 0 {
		t.Fatalf("expected peer to expire, got %d", len(got))
	}
}

func TestRegistryPresence(t *testing.T) {
	clock := time.Now()
	r := NewRegistry()
	r.now = func() time.Time { return clock }

	r.Upsert(Peer{Instance: "a", Modified: 0})
	r.Upsert(Peer{Instance: "b", Modified: 4})

	a, _ := r.Get("a")
	b, _ := r.Get("b")
	if got := r.PresenceOf(a); got != PresenceOnline {
		t.Errorf("clean peer presence = %q, want online", got)
	}
	if got := r.PresenceOf(b); got != PresenceCoding {
		t.Errorf("dirty peer presence = %q, want coding", got)
	}

	clock = clock.Add(IdleAfter + time.Minute)
	if got := r.PresenceOf(a); got != PresenceIdle {
		t.Errorf("stale peer presence = %q, want idle", got)
	}
}
