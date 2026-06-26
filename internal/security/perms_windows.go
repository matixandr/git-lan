//go:build windows

package security

import (
	"os"
	"os/exec"
)

// secureFilePerms hardens a secret file on Windows using icacls: it disables
// inheritance and grants full access only to the current user, approximating
// Unix 0600. Best effort - failures are non-fatal because the file already
// lives under the per-user profile directory and was created with restrictive
// Go permissions.
func secureFilePerms(path string) error {
	user := os.Getenv("USERNAME")
	if user == "" {
		return nil
	}
	// /inheritance:r removes inherited ACEs; /grant:r replaces this user's ACE.
	cmd := exec.Command("icacls", path, "/inheritance:r", "/grant:r", user+":F")
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run() // best effort
	return nil
}
