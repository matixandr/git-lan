package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
	"github.com/matixandr/git-lan/internal/git"
	"github.com/matixandr/git-lan/pkg/config"
)

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

// localAdvertisement builds the TXT metadata for the repo in the current
// directory, used when a command announces this host.
func localAdvertisement(sessionName string, locked bool, presence discovery.Presence) (discovery.Advertisement, *git.Repo, error) {
	repo, err := git.Open("")
	if err != nil {
		return discovery.Advertisement{}, nil, err
	}
	branch, _ := repo.Branch()
	ad := discovery.Advertisement{
		Repo:     repo.Name(),
		Branch:   branch,
		Head:     repo.Head(),
		Modified: repo.ModifiedCount(),
		Session:  sessionName,
		Locked:   locked,
		Presence: presence,
	}
	return ad, repo, nil
}

// displayName returns the configured display name, falling back to the hostname.
func displayName() string {
	cfg, _ := config.Load()
	if cfg.DisplayName != "" {
		return cfg.DisplayName
	}
	if h, err := os.Hostname(); err == nil {
		return h
	}
	return "git-lan-peer"
}
