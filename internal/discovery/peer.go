package discovery

import (
	"fmt"
	"net"
	"time"
)

// Presence describes what a peer is currently doing, derived from its mDNS
// advertisement and last-seen time.
type Presence string

const (
	PresenceOnline  Presence = "online"  // reachable, clean working tree
	PresenceCoding  Presence = "coding"  // reachable, uncommitted changes
	PresenceIdle    Presence = "idle"    // reachable, no update for a while
	PresenceOffline Presence = "offline" // not responding
)

// Peer is a host advertising a git-lan repository on the local network. The
// fields below the address come straight from the (unencrypted) mDNS TXT
// record - they are metadata only and never carry secrets.
type Peer struct {
	// Instance is the mDNS service instance name, typically the hostname.
	Instance string
	// Host is the resolvable hostname (e.g. "maciek-laptop.local").
	Host string
	// Addrs are the resolved IP addresses.
	Addrs []net.IP
	// Port is the encrypted transport port advertised by the peer.
	Port int

	// Repo is the short repository name being shared.
	Repo string
	// Branch is the current branch (truncated to 20 chars in the TXT record).
	Branch string
	// Head is the short HEAD hash (7 chars).
	Head string
	// Modified is the count of locally modified files reported by the peer.
	Modified int
	// Session is the active session name, empty if none.
	Session string
	// Locked indicates the session is password-protected.
	Locked bool
	// Advertised is the presence the peer reports about itself (it knows its
	// own activity). The registry trusts this unless the peer has gone stale.
	Advertised Presence
	// Protocol is the advertised protocol version, e.g. "v1".
	Protocol string

	// LastSeen is when we last received an advertisement from this peer.
	LastSeen time.Time
}

// Addr returns a dialable host:port for the peer, preferring an IPv4 address.
func (p Peer) Addr() (string, error) {
	ip := p.preferredIP()
	if ip == nil {
		if p.Host == "" {
			return "", fmt.Errorf("peer %s has no usable address", p.Instance)
		}
		return net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port)), nil
	}
	return net.JoinHostPort(ip.String(), fmt.Sprintf("%d", p.Port)), nil
}

func (p Peer) preferredIP() net.IP {
	var v6 net.IP
	for _, ip := range p.Addrs {
		if v4 := ip.To4(); v4 != nil {
			return v4
		}
		if v6 == nil {
			v6 = ip
		}
	}
	return v6
}

// Name returns the short display name for the peer (instance without the
// trailing service suffix).
func (p Peer) Name() string {
	return p.Instance
}
