package display

import (
	"fmt"
	"os"

	"github.com/matixandr/git-lan/pkg/config"
	"golang.org/x/term"
)

// Options controls Nerd Fonts detection, supplied from CLI flags and config.
type Options struct {
	// ForceOff corresponds to --no-nerd-fonts: highest priority, never probes.
	ForceOff bool
	// ConfigValue is the explicit nerd_fonts setting from config.toml, or nil
	// when unset (auto-detect).
	ConfigValue *bool
}

// DetectNerdFonts resolves whether to use Nerd Fonts glyphs, following a strict
// override hierarchy:
//
//  1. --no-nerd-fonts            → false, no probe, no cache
//  2. nerd_fonts in config.toml  → explicit global override
//  3. terminal profile cache hit → return cached value
//  4. TTY probe                  → probe, cache, return
//  5. non-TTY fallback           → false, do not cache
func DetectNerdFonts(opts Options) bool {
	if opts.ForceOff {
		return false
	}
	if opts.ConfigValue != nil {
		return *opts.ConfigValue
	}

	key := terminalProfileKey()
	cache := loadCache()
	if val, found := cache.lookup(key); found {
		return val
	}

	// Cache miss. Only probe a real terminal; never cache a non-TTY result,
	// since piped/CI output says nothing about the user's actual terminal.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	result := probeNerdFonts()
	cache.set(key, result)
	_ = cache.save() // best effort; a failed cache write just means re-probe
	return result
}

// RedetectCurrent removes the current terminal's cached entry and re-runs the
// probe, returning the new result and the profile key it was stored under. It
// backs `git lan config --detect-fonts`.
func RedetectCurrent() (nerd bool, profile string, err error) {
	key := terminalProfileKey()
	cache := loadCache()
	cache.remove(key)

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false, key, fmt.Errorf("stdout is not a terminal; cannot probe")
	}
	nerd = probeNerdFonts()
	cache.set(key, nerd)
	if err := cache.save(); err != nil {
		return nerd, key, err
	}
	return nerd, key, nil
}

// ConfigBoolPtr adapts a config value (which may be unset) into the *bool the
// detector expects.
func ConfigBoolPtr(cfg config.Config) *bool { return cfg.NerdFonts }
