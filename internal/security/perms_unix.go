//go:build !windows

package security

import "os"

// secureFilePerms enforces owner-only (0600) permissions on Unix.
func secureFilePerms(path string) error {
	return os.Chmod(path, 0o600)
}
