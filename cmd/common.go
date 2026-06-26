package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
	"github.com/matixandr/git-lan/internal/security"
	"github.com/matixandr/git-lan/internal/transport"
)

// resolvePeer parses a "peer/repo" reference, browses for the peer, and returns
// it along with the repo name to request (defaulting to the peer's advertised
// repo when omitted).
func resolvePeer(ref string, d time.Duration) (discovery.Peer, string, error) {
	target, err := transport.ParseTarget(ref)
	if err != nil {
		return discovery.Peer{}, "", err
	}
	peer, err := findPeer(target.Peer, d)
	if err != nil {
		return discovery.Peer{}, "", err
	}
	repo := target.Repo
	if repo == "" {
		repo = peer.Repo
	}
	if repo == "" {
		return discovery.Peer{}, "", fmt.Errorf("peer %q is not advertising a repo; specify peer/repo", target.Peer)
	}
	return peer, repo, nil
}

// clientForPeer builds an encrypted transport client aimed at a peer, using
// this host's long-term identity.
func clientForPeer(peer discovery.Peer) (*transport.Client, error) {
	id, err := security.LoadOrCreateIdentity()
	if err != nil {
		return nil, err
	}
	addr, err := peer.Addr()
	if err != nil {
		return nil, err
	}
	c := &transport.Client{
		Identity: id.Private(),
		PeerAddr: addr,
		Verify:   verifyPinnedPeer(peer.Instance),
	}
	if flagVerbose {
		c.Log = func(format string, args ...any) { fmt.Fprintf(os.Stderr, "[transport] "+format+"\n", args...) }
	}
	return c, nil
}

// verifyPinnedPeer returns a VerifyFunc that checks the peer's presented
// identity against any pin for hostname. A mismatch aborts loudly (possible
// MITM); an unknown peer is allowed (trust-on-first-use), and a matching pin
// passes silently.
func verifyPinnedPeer(hostname string) transport.VerifyFunc {
	return func(peerIdentity []byte) error {
		fp := security.FingerprintOf(peerIdentity)
		ring, err := security.LoadTrust()
		if err != nil {
			return nil // do not block on a trust-store read error
		}
		_, err = ring.VerifyHost(hostname, fp)
		if errors.Is(err, security.ErrFingerprintMismatch) {
			return fmt.Errorf("%w\n  peer %q presented %s\n  this could be a man-in-the-middle - aborting",
				err, hostname, fp)
		}
		return nil
	}
}

// browseFor runs mDNS discovery for d and returns the peers seen. It announces
// nothing - it is a pure listen, used by `list` and one-shot lookups.
func browseFor(d time.Duration) ([]discovery.Peer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	svc, err := discovery.Start(ctx, 0, nil)
	if err != nil {
		return nil, err
	}
	defer svc.Stop()

	<-ctx.Done()
	return svc.Peers(), nil
}

// findPeer browses briefly and returns the peer whose instance name matches
// name (case-insensitive prefix), or an error if none is found.
func findPeer(name string, d time.Duration) (discovery.Peer, error) {
	peers, err := browseFor(d)
	if err != nil {
		return discovery.Peer{}, err
	}
	for _, p := range peers {
		if equalFoldName(p.Instance, name) {
			return p, nil
		}
	}
	return discovery.Peer{}, fmt.Errorf("peer %q not found on the network", name)
}

func equalFoldName(a, b string) bool {
	return len(a) >= len(b) && toLower(a[:len(b)]) == toLower(b)
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

