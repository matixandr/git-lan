package transport

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func mustID(t *testing.T) *ecdh.PrivateKey {
	t.Helper()
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestEndToEndClone clones a repo from an encrypted Server through the Client
// bridge and verifies the file content survives the round trip. This exercises
// the whole stack: handshake, EncryptedConn, git daemon --inetd, and the bridge.
func TestEndToEndClone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}

	// Build a source repo with one commit.
	srcParent := t.TempDir()
	src := filepath.Join(srcParent, "demo")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, src, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(src, "hello.txt"), []byte("ahoj svete"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, src, "add", ".")
	git(t, src, "commit", "-q", "-m", "first")

	// Stand up the encrypted server.
	srv := NewServer(mustID(t), src, false)
	srv.Log = t.Logf
	if err := srv.Listen(0); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()

	// Client bridge to the server.
	cli := &Client{
		Identity: mustID(t),
		PeerAddr: srv.ln.Addr().String(),
		Log:      t.Logf,
	}
	url, stop, err := cli.Bridge(srv.RepoName())
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	// Clone through the tunnel.
	dest := filepath.Join(t.TempDir(), "clone")
	clone := exec.Command("git", "clone", "-q", url, dest)
	clone.Env = os.Environ()
	done := make(chan error, 1)
	go func() {
		out, err := clone.CombinedOutput()
		if err != nil {
			t.Logf("clone output: %s", out)
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("clone failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("clone timed out")
	}

	got, err := os.ReadFile(filepath.Join(dest, "hello.txt"))
	if err != nil {
		t.Fatalf("cloned file missing: %v", err)
	}
	if string(got) != "ahoj svete" {
		t.Fatalf("content mismatch: %q", got)
	}
}

// makeRepo builds a one-commit repo and returns its path.
func makeRepo(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	src := filepath.Join(parent, "demo")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, src, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, src, "add", ".")
	git(t, src, "commit", "-q", "-m", "c")
	return src
}

// TestPasswordGateClone confirms the password gate admits a correct seed and
// rejects a wrong one through the full clone path.
func TestPasswordGateClone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
	src := makeRepo(t)

	salt := []byte("0123456789abcdef")
	seed := []byte("shared-32-byte-session-seed-value")

	srv := NewServer(mustID(t), src, false)
	srv.RequireAuth = true
	srv.Salt = salt
	srv.Seed = seed
	if err := srv.Listen(0); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Serve(ctx) }()

	tryClone := func(derive func([]byte) []byte) error {
		cli := &Client{Identity: mustID(t), PeerAddr: srv.ln.Addr().String(), DeriveSeed: derive}
		url, stop, err := cli.Bridge(srv.RepoName())
		if err != nil {
			return err
		}
		defer stop()
		dest := filepath.Join(t.TempDir(), "c")
		cmd := exec.Command("git", "clone", "-q", url, dest)
		cmd.Env = os.Environ()
		done := make(chan error, 1)
		go func() { _, e := cmd.CombinedOutput(); done <- e }()
		select {
		case e := <-done:
			return e
		case <-time.After(30 * time.Second):
			return context.DeadlineExceeded
		}
	}

	if err := tryClone(func([]byte) []byte { return seed }); err != nil {
		t.Errorf("correct seed should clone, got %v", err)
	}
	if err := tryClone(func([]byte) []byte { return []byte("the-wrong-seed-entirely-nope-xx!") }); err == nil {
		t.Error("wrong seed should fail the clone")
	}
}
