package presence

import (
	"testing"
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
)

func TestComputePresence(t *testing.T) {
	now := time.Now()
	recent := now.Add(-time.Minute)
	stale := now.Add(-IdleAfter - time.Minute)

	cases := []struct {
		name     string
		dirty    bool
		activity time.Time
		want     discovery.Presence
	}{
		{"active dirty -> coding", true, recent, discovery.PresenceCoding},
		{"active clean -> online", false, recent, discovery.PresenceOnline},
		{"stale dirty -> idle", true, stale, discovery.PresenceIdle},
		{"stale clean -> idle", false, stale, discovery.PresenceIdle},
	}
	for _, c := range cases {
		if got := Compute(c.dirty, c.activity, now); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}
