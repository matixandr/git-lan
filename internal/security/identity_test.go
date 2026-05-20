package security

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// isolateConfig points the config directory at a temp dir for the duration of a
// test, so identity creation does not touch the real user profile.
func isolateConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
		t.Setenv("HOME", dir)
	}
}

func TestLoadOrCreateIdentityRoundTrip(t *testing.T) {
	isolateConfig(t)

	first, err := LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("create identity: %v", err)
	}
	// The key file must now exist.
	path, _ := identityPathForTest(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("identity key not written: %v", err)
	}

	// Loading again must return the same identity, not generate a new one.
	second, err := LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("reload identity: %v", err)
	}
	if first.Fingerprint() != second.Fingerprint() {
		t.Fatal("identity changed across loads - key not persisted")
	}
}

func identityPathForTest(t *testing.T) (string, error) {
	t.Helper()
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir = filepath.Join(os.Getenv("XDG_CONFIG_HOME"))
	}
	return filepath.Join(dir, "gitlan", "identity.key"), nil
}
