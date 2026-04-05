//go:build windows

package security

// secureFilePerms is a best-effort placeholder on Windows. NTFS ACL hardening
// (stripping inherited access so only the current user can read the key) is
// wired up in a later pass; the file is still created with restrictive Go
// permissions and lives under the user's per-profile config directory.
func secureFilePerms(path string) error {
	return nil
}
