// Package presence derives a host's self-reported activity state - online,
// coding, or idle - from its working-tree cleanliness and recent edit activity.
// This is what a host broadcasts about itself; reachability (offline) is decided
// separately by peers from mDNS liveness.
package presence

import (
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
)

// IdleAfter is how long without an edit before a host reports itself idle.
const IdleAfter = 15 * time.Minute

// Compute returns the presence a host should advertise.
//
//	idle    - no edit activity for at least IdleAfter
//	coding  - recently active and the working tree is dirty
//	online  - recently active with a clean tree
func Compute(dirty bool, lastActivity time.Time, now time.Time) discovery.Presence {
	if now.Sub(lastActivity) >= IdleAfter {
		return discovery.PresenceIdle
	}
	if dirty {
		return discovery.PresenceCoding
	}
	return discovery.PresenceOnline
}
