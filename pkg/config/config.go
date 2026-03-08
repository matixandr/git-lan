package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Config is the user-editable global configuration, loaded from config.toml.
// Every field has a sane zero-config default; the file need not exist.
type Config struct {
	// DisplayName is how this host advertises itself to peers. Defaults to
	// the OS hostname when empty.
	DisplayName string `toml:"display_name"`

	// NerdFonts is an explicit global override for icon rendering. When nil,
	// git-lan auto-detects per terminal profile. A non-nil value short-circuits
	// detection entirely.
	NerdFonts *bool `toml:"nerd_fonts"`

	// Port is the preferred TCP port for the encrypted transport. 0 means
	// "pick the default and fall back dynamically if taken".
	Port int `toml:"port"`

	// AllowPush, when true, lets trusted peers push without an interactive
	// approval prompt. Off by default - pushes are confirmed by a human.
	AllowPush bool `toml:"allow_push"`
}

// Default returns the built-in configuration used when no file is present.
func Default() Config {
	return Config{
		DisplayName: "",
		NerdFonts:   nil,
		Port:        0,
		AllowPush:   false,
	}
}

// Load reads config.toml, returning defaults if the file does not exist.
func Load() (Config, error) {
	cfg := Default()
	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	} else if err != nil {
		return cfg, err
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes the configuration back to config.toml.
func (c Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
