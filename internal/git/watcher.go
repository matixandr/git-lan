package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher reports when the working tree was last touched, so a session can show
// "coding" while a peer is actively editing and "idle" once they stop. It does
// not interpret changes - it only timestamps activity.
type Watcher struct {
	root string
	fsw  *fsnotify.Watcher

	mu           sync.Mutex
	lastActivity time.Time
}

// NewWatcher creates a recursive watcher over the repository working tree,
// skipping the .git directory. Activity starts at creation time.
func NewWatcher(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{root: root, fsw: fsw, lastActivity: time.Now()}
	w.addTree(root)
	return w, nil
}

// addTree adds watches for root and every subdirectory except .git.
func (w *Watcher) addTree(root string) {
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			_ = w.fsw.Add(path)
		}
		return nil
	})
}

// Start consumes filesystem events until ctx is canceled, updating the activity
// timestamp and watching newly created directories.
func (w *Watcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-w.fsw.Events:
				if !ok {
					return
				}
				if strings.Contains(filepath.ToSlash(ev.Name), "/.git/") {
					continue // ignore git's own bookkeeping
				}
				w.touch()
				// Track directories created during the session.
				if ev.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
						_ = w.fsw.Add(ev.Name)
					}
				}
			case _, ok := <-w.fsw.Errors:
				if !ok {
					return
				}
			}
		}
	}()
}

func (w *Watcher) touch() {
	w.mu.Lock()
	w.lastActivity = time.Now()
	w.mu.Unlock()
}

// LastActivity returns when the tree was last modified.
func (w *Watcher) LastActivity() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastActivity
}

// Close stops watching and releases resources.
func (w *Watcher) Close() error { return w.fsw.Close() }
