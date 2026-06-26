package discovery

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/grandcat/zeroconf"
)

// Announcer registers this host's repo as a _gitlan._tcp service on the LAN
// and keeps the registration alive until Close is called.
type Announcer struct {
	server   *zeroconf.Server
	instance string
}

// hostInstance returns a stable, human-friendly instance name for this host.
func hostInstance() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		// Strip any trailing .local and domain suffix for a clean label.
		if i := strings.IndexByte(h, '.'); i >= 0 {
			h = h[:i]
		}
		return h
	}
	return "git-lan-peer"
}

// Announce registers the service and starts broadcasting. The instance name
// defaults to the hostname. Call Update to refresh the TXT metadata and Close
// to send a goodbye and stop.
func Announce(instance string, port int, ad Advertisement) (*Announcer, error) {
	if instance == "" {
		instance = hostInstance()
	}
	server, err := zeroconf.Register(instance, ServiceType, Domain, port, ad.TXT(), nil)
	if err != nil {
		return nil, fmt.Errorf("mDNS register: %w", err)
	}
	return &Announcer{server: server, instance: instance}, nil
}

// Update replaces the advertised TXT metadata in place (e.g. when the branch,
// HEAD, or presence changes).
func (a *Announcer) Update(ad Advertisement) {
	if a.server != nil {
		a.server.SetText(ad.TXT())
	}
}

// Instance returns the announced instance name.
func (a *Announcer) Instance() string { return a.instance }

// Close sends an mDNS goodbye and tears down the registration.
func (a *Announcer) Close() {
	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
	}
}

// Browse discovers git-lan peers on the LAN, feeding the registry until ctx is
// canceled. Announcements matching selfInstance are ignored so a host never
// lists itself. It blocks; run it in a goroutine.
func Browse(ctx context.Context, reg *Registry, selfInstance string) error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 16)
	go func() {
		for entry := range entries {
			if entry == nil {
				continue
			}
			if strings.EqualFold(entry.Instance, selfInstance) {
				continue // never list ourselves
			}
			reg.Upsert(peerFromEntry(entry))
		}
	}()

	if err := resolver.Browse(ctx, ServiceType, Domain, entries); err != nil {
		return fmt.Errorf("mDNS browse: %w", err)
	}
	<-ctx.Done()
	return nil
}

// peerFromEntry converts a zeroconf service entry into a Peer.
func peerFromEntry(e *zeroconf.ServiceEntry) Peer {
	ad := ParseTXT(e.Text)
	p := Peer{
		Instance:   e.Instance,
		Host:       strings.TrimSuffix(e.HostName, "."),
		Port:       e.Port,
		Repo:       ad.Repo,
		Branch:     ad.Branch,
		Head:       ad.Head,
		Modified:   ad.Modified,
		Session:    ad.Session,
		Locked:     ad.Locked,
		Advertised: ad.Presence,
		Protocol:   protocolOf(e.Text),
	}
	p.Addrs = append(p.Addrs, e.AddrIPv4...)
	p.Addrs = append(p.Addrs, e.AddrIPv6...)
	return p
}
