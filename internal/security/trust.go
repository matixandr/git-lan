package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/matixandr/git-lan/pkg/config"
)

// ErrFingerprintMismatch indicates a peer presented an identity that does not
// match its pinned fingerprint - a possible man-in-the-middle.
var ErrFingerprintMismatch = errors.New("peer fingerprint does not match pinned trust entry")

// TrustEntry pins a hostname to the fingerprint we expect it to present.
type TrustEntry struct {
	Hostname    string    `json:"hostname"`
	Fingerprint string    `json:"fingerprint"`
	AddedAt     time.Time `json:"added_at"`
}

// TrustRing is the persisted set of trusted peers, keyed by hostname.
type TrustRing struct {
	Entries map[string]TrustEntry `json:"entries"`
}

// LoadTrust reads trusted_peers.json, returning an empty ring if absent.
func LoadTrust() (*TrustRing, error) {
	tr := &TrustRing{Entries: map[string]TrustEntry{}}
	path, err := config.TrustPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return tr, nil
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, tr); err != nil {
		return nil, err
	}
	if tr.Entries == nil {
		tr.Entries = map[string]TrustEntry{}
	}
	return tr, nil
}

// Save persists the ring.
func (t *TrustRing) Save() error {
	path, err := config.TrustPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Add pins hostname to fingerprint. It overwrites an existing pin (use with
// care - that is how legitimate key rotation is recorded).
func (t *TrustRing) Add(hostname, fingerprint string) {
	t.Entries[hostname] = TrustEntry{
		Hostname:    hostname,
		Fingerprint: fingerprint,
		AddedAt:     time.Now(),
	}
}

// Remove deletes a pin.
func (t *TrustRing) Remove(hostname string) bool {
	if _, ok := t.Entries[hostname]; !ok {
		return false
	}
	delete(t.Entries, hostname)
	return true
}

// Get returns the pin for a hostname.
func (t *TrustRing) Get(hostname string) (TrustEntry, bool) {
	e, ok := t.Entries[hostname]
	return e, ok
}

// VerifyHost checks a presented fingerprint against any pin for hostname.
// It distinguishes three cases:
//
//	known + match    → trusted, nil error
//	known + mismatch → ErrFingerprintMismatch (possible MITM)
//	unknown          → trusted=false, nil error (caller decides: TOFU prompt)
func (t *TrustRing) VerifyHost(hostname, fingerprint string) (trusted bool, err error) {
	e, ok := t.Entries[hostname]
	if !ok {
		return false, nil
	}
	if !ConstantTimeEqual([]byte(e.Fingerprint), []byte(fingerprint)) {
		return false, fmt.Errorf("%w: %s is pinned to %s but presented %s",
			ErrFingerprintMismatch, hostname, e.Fingerprint, fingerprint)
	}
	return true, nil
}

// List returns all pins sorted by hostname.
func (t *TrustRing) List() []TrustEntry {
	out := make([]TrustEntry, 0, len(t.Entries))
	for _, e := range t.Entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Hostname < out[j].Hostname })
	return out
}
