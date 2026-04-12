// Package git wraps the system git binary. git-lan shells out to git rather
// than linking a git library: it is pragmatic, matches the user's real git
// behaviour exactly, and works identically on every platform.
package git

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotARepo is returned when the working directory is not inside a git repo.
var ErrNotARepo = errors.New("not a git repository")

// Repo is a handle to a git repository on disk, rooted at its top level.
type Repo struct {
	Root string
}

// Open locates the repository containing dir (or the current directory when dir
// is empty) by asking git for the work-tree root.
func Open(dir string) (*Repo, error) {
	out, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, ErrNotARepo
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return nil, ErrNotARepo
	}
	return &Repo{Root: root}, nil
}

// Name returns the repository's directory name, used as the share name.
func (r *Repo) Name() string { return filepath.Base(r.Root) }

// Branch returns the current branch, or "HEAD" when detached.
func (r *Repo) Branch() (string, error) {
	out, err := run(r.Root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Head returns the short (7-char) HEAD commit hash, or "" for an empty repo.
func (r *Repo) Head() string {
	out, err := run(r.Root, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ModifiedFiles returns the paths of files with uncommitted changes (staged,
// unstaged, or untracked), parsed from `git status --porcelain`.
func (r *Repo) ModifiedFiles() ([]string, error) {
	out, err := run(r.Root, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		// Porcelain v1: "XY <path>"; handle rename "orig -> new".
		path := strings.TrimSpace(line[3:])
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		files = append(files, path)
	}
	return files, nil
}

// ModifiedCount is the number of files with uncommitted changes.
func (r *Repo) ModifiedCount() int {
	files, err := r.ModifiedFiles()
	if err != nil {
		return 0
	}
	return len(files)
}

// IsClean reports whether the working tree has no uncommitted changes.
func (r *Repo) IsClean() bool { return r.ModifiedCount() == 0 }

// run executes git in dir and returns combined stdout, or an error carrying
// stderr for diagnostics.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New("git " + strings.Join(args, " ") + ": " + msg)
	}
	return stdout.String(), nil
}
