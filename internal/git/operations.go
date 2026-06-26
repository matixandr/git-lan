package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Operations run git as a child process, streaming its output so the user sees
// familiar git progress. They take a fully-formed remote URL - for git-lan the
// transport rewrites that URL to point at a local proxy socket that tunnels the
// encrypted connection (see internal/transport).

// Clone clones remoteURL into destDir.
func Clone(remoteURL, destDir string, out io.Writer) error {
	return stream(out, "", "clone", remoteURL, destDir)
}

// Fetch fetches refs from remoteURL into the repo at root.
func (r *Repo) Fetch(remoteURL string, out io.Writer) error {
	return stream(out, r.Root, "fetch", remoteURL)
}

// Pull fetches and merges branch from remoteURL into the repo at root.
func (r *Repo) Pull(remoteURL, branch string, out io.Writer) error {
	args := []string{"pull", remoteURL}
	if branch != "" {
		args = append(args, branch)
	}
	return stream(out, r.Root, args...)
}

// Push pushes branch to remoteURL. An empty branch pushes the current branch.
func (r *Repo) Push(remoteURL, branch string, out io.Writer) error {
	if branch == "" {
		b, err := r.Branch()
		if err != nil {
			return err
		}
		branch = b
	}
	return stream(out, r.Root, "push", remoteURL, branch)
}

// stream runs git in dir with its stdout/stderr connected to out so progress is
// visible live.
func stream(out io.Writer, dir string, args ...string) error {
	if out == nil {
		out = os.Stderr
	}
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", args[0], err)
	}
	return nil
}
