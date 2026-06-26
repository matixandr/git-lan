package display

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/matixandr/git-lan/pkg/config"
)

// profileCache is the on-disk record of Nerd Fonts detection results, keyed by
// terminal profile so each terminal is probed at most once.
type profileCache struct {
	Profiles map[string]profileEntry `toml:"profiles"`
}

type profileEntry struct {
	NerdFonts  bool   `toml:"nerd_fonts"`
	DetectedAt string `toml:"detected_at"`
}

// loadCache reads terminal_profiles.toml, returning an empty cache if absent.
func loadCache() *profileCache {
	c := &profileCache{Profiles: map[string]profileEntry{}}
	path, err := config.TerminalProfilesPath()
	if err != nil {
		return c
	}
	if _, err := toml.DecodeFile(path, c); err != nil {
		// A corrupt or missing cache is non-fatal: behave as a clean slate.
		return &profileCache{Profiles: map[string]profileEntry{}}
	}
	if c.Profiles == nil {
		c.Profiles = map[string]profileEntry{}
	}
	return c
}

// lookup returns the cached result for key and whether it was present.
func (c *profileCache) lookup(key string) (nerd bool, found bool) {
	e, ok := c.Profiles[key]
	return e.NerdFonts, ok
}

// set records a detection result for key, stamped with the current time.
func (c *profileCache) set(key string, nerd bool) {
	c.Profiles[key] = profileEntry{
		NerdFonts:  nerd,
		DetectedAt: time.Now().Format("2006-01-02T15:04:05"),
	}
}

// remove deletes a profile entry so the next detection re-probes it.
func (c *profileCache) remove(key string) {
	delete(c.Profiles, key)
}

// save persists the cache to disk, creating the file if needed.
func (c *profileCache) save() error {
	path, err := config.TerminalProfilesPath()
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
