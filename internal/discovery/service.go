package discovery

import (
	"context"
	"sync"
)

// Service combines announcing this host and browsing for peers into a single
// lifecycle object. It owns a Registry that callers can query for the live
// peer list.
type Service struct {
	Registry *Registry

	announcer *Announcer
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// Start begins browsing for peers immediately. If announce is non-nil the host
// also advertises itself on the given port. Either side may be used alone:
// `git lan list` browses without announcing; a session announces and browses.
func Start(ctx context.Context, port int, announce *Advertisement) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	svc := &Service{
		Registry: NewRegistry(),
		cancel:   cancel,
	}

	self := hostInstance()
	if announce != nil {
		a, err := Announce(self, port, *announce)
		if err != nil {
			cancel()
			return nil, err
		}
		svc.announcer = a
		self = a.Instance()
	}

	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		// Browse returns on ctx cancellation or a fatal resolver error; an
		// interface flapping mid-session surfaces here as a transient error,
		// so we simply stop browsing rather than crash the process.
		_ = Browse(ctx, svc.Registry, self)
	}()

	return svc, nil
}

// UpdateAdvertisement refreshes the broadcast metadata, if announcing.
func (s *Service) UpdateAdvertisement(ad Advertisement) {
	if s.announcer != nil {
		s.announcer.Update(ad)
	}
}

// Peers returns the current live peer list.
func (s *Service) Peers() []Peer { return s.Registry.List() }

// Stop tears down announcing and browsing and waits for goroutines to exit.
func (s *Service) Stop() {
	if s.announcer != nil {
		s.announcer.Close()
	}
	s.cancel()
	s.wg.Wait()
}
