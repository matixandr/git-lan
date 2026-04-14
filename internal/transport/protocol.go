// Package transport carries git's wire protocol between peers over an encrypted
// channel. The design keeps git itself oblivious to encryption:
//
//	server:  encrypted listener  ──decrypt──▶  git daemon on 127.0.0.1
//	client:  git  ──▶  127.0.0.1 bridge  ──encrypt──▶  peer's encrypted listener
//
// git speaks plaintext to a loopback socket; this package is the only thing
// that touches the network, and everything it sends is ChaCha20-Poly1305.
package transport

import (
	"fmt"
	"strings"
)

// DefaultPort is the preferred TCP port for the encrypted transport. git's own
// daemon traditionally uses 9418; we reuse the number for familiarity but the
// bytes on it are encrypted, and we fall back to a dynamic port if it is taken.
const DefaultPort = 9418

// Target identifies a repository on a peer, written as "peer" or "peer/repo".
type Target struct {
	Peer string
	Repo string
}

// ParseTarget splits a "peer/repo" reference. A bare "peer" leaves Repo empty,
// meaning "the peer's default shared repo".
func ParseTarget(s string) (Target, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Target{}, fmt.Errorf("empty peer reference")
	}
	peer, repo, found := strings.Cut(s, "/")
	if peer == "" {
		return Target{}, fmt.Errorf("invalid reference %q: missing peer", s)
	}
	t := Target{Peer: peer}
	if found {
		t.Repo = repo
	}
	return t, nil
}

// String renders a target back to "peer/repo" form.
func (t Target) String() string {
	if t.Repo == "" {
		return t.Peer
	}
	return t.Peer + "/" + t.Repo
}

// gitDaemonURL builds the git:// URL git uses to talk to a local loopback
// bridge or daemon.
func gitDaemonURL(host string, port int, repo string) string {
	repo = strings.TrimPrefix(repo, "/")
	return fmt.Sprintf("git://%s:%d/%s", host, port, repo)
}
