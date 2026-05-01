package discovery

import (
	"sort"
	"sync"
	"time"
)

// Default presence/TTL timing. A peer is considered offline after it misses
// two heartbeat intervals.
const (
	HeartbeatInterval = 60 * time.Second
	PeerTTL           = 2 * HeartbeatInterval
)

// Registry is a concurrency-safe, in-memory view of currently known peers.
// Entries expire once they have not been seen within PeerTTL.
type Registry struct {
	mu    sync.RWMutex
	peers map[string]Peer
	now   func() time.Time // injectable clock for tests
}

// NewRegistry returns an empty peer registry.
func NewRegistry() *Registry {
	return &Registry{
		peers: make(map[string]Peer),
		now:   time.Now,
	}
}

// Upsert inserts or refreshes a peer, stamping its LastSeen.
func (r *Registry) Upsert(p Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.LastSeen = r.now()
	r.peers[p.Instance] = p
}

// Remove deletes a peer by instance name (e.g. on an mDNS "goodbye").
func (r *Registry) Remove(instance string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, instance)
}

// Get returns a peer by instance name.
func (r *Registry) Get(instance string) (Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.peers[instance]
	return p, ok
}

// List returns all live peers (not yet expired), sorted by name. Expired peers
// are dropped lazily on read.
func (r *Registry) List() []Peer {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	out := make([]Peer, 0, len(r.peers))
	for name, p := range r.peers {
		if now.Sub(p.LastSeen) > PeerTTL {
			delete(r.peers, name)
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Instance < out[j].Instance })
	return out
}

// PresenceOf derives the live presence of a peer using the registry's clock.
func (r *Registry) PresenceOf(p Peer) Presence { return presenceAt(p, r.now()) }

// presenceAt is the shared presence rule. Reachability is decided locally from
// LastSeen (a peer past PeerTTL is offline regardless of what it last claimed).
// Everything else - coding vs idle vs online - is self-reported by the peer,
// which alone knows its own activity. We fall back to deriving "coding" from the
// modified-file count if the peer advertised nothing.
func presenceAt(p Peer, now time.Time) Presence {
	if now.Sub(p.LastSeen) > PeerTTL {
		return PresenceOffline
	}
	switch p.Advertised {
	case PresenceIdle, PresenceCoding, PresenceOnline:
		return p.Advertised
	default:
		if p.Modified > 0 {
			return PresenceCoding
		}
		return PresenceOnline
	}
}
