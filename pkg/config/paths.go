package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the git-lan configuration directory for the current
// platform, creating it if necessary. Everything git-lan persists - identity
// key, sessions, trust ring, terminal profile cache - lives here.
//
//	linux/macOS: $XDG_CONFIG_HOME/gitlan, falling back to ~/.gitlan
//	windows:     %APPDATA%\gitlan, falling back to ~/.gitlan
func ConfigDir() (string, error) {
	dir := resolveConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func resolveConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "gitlan")
		}
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "gitlan")
		}
	}
	return filepath.Join(home, ".gitlan")
}

// configFile joins name onto the config directory.
func configFile(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// Standard config file locations.

// ConfigPath is the main TOML config file.
func ConfigPath() (string, error) { return configFile("config.toml") }

// IdentityPath is the long-term X25519 identity key.
func IdentityPath() (string, error) { return configFile("identity.key") }

// SessionsPath stores active session state.
func SessionsPath() (string, error) { return configFile("sessions.json") }

// TrustPath stores the trusted peers ring.
func TrustPath() (string, error) { return configFile("trusted_peers.json") }

// TerminalProfilesPath caches per-terminal Nerd Fonts detection.
func TerminalProfilesPath() (string, error) { return configFile("terminal_profiles.toml") }
