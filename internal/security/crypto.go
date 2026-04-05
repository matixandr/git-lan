// Package security holds git-lan's long-term identity keys, peer fingerprints,
// and the trusted-peers ring. The per-connection symmetric encryption lives in
// internal/e2e; this package is about *who* a peer is, not *how* bytes are
// encrypted in flight.
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"os"
	"path/filepath"
)

// RandomBytes returns n cryptographically random bytes.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// ConstantTimeEqual compares two byte slices without leaking length-position
// timing. Length mismatch returns false.
func ConstantTimeEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// writeSecretFile writes data to path atomically with 0600 permissions. The
// file is created via a temp file in the same directory and renamed, so a
// reader never observes a partially written key.
func writeSecretFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return secureFilePerms(path)
}
