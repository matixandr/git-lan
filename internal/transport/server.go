package transport

import (
	"context"
	"crypto/ecdh"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/matixandr/git-lan/internal/e2e"
)

// Server accepts encrypted peer connections and serves a single git repository
// over them. Each accepted connection is authenticated, optionally checked
// against the trust ring, then handed to a one-shot `git daemon --inetd` whose
// stdin/stdout are the decrypted stream.
type Server struct {
	identity  *ecdh.PrivateKey
	repoRoot  string
	repoName  string
	allowPush bool

	// Verify checks the peer identity after the handshake; nil = trust-on-
	// first-use. Set by the trust layer.
	Verify VerifyFunc
	// Log receives human-readable diagnostics; may be nil.
	Log func(format string, args ...any)

	// Password gate. When RequireAuth is true, peers must prove knowledge of
	// the session password (via Seed/Salt) before any git bytes are served.
	RequireAuth bool
	Seed        []byte
	Salt        []byte

	ln   net.Listener
	port int

	mu   sync.Mutex
	subs map[*exec.Cmd]struct{}
}

// NewServer creates a server for the repo rooted at repoRoot.
func NewServer(identity *ecdh.PrivateKey, repoRoot string, allowPush bool) *Server {
	return &Server{
		identity:  identity,
		repoRoot:  repoRoot,
		repoName:  filepath.Base(repoRoot),
		allowPush: allowPush,
		subs:      make(map[*exec.Cmd]struct{}),
	}
}

// Port returns the actual TCP port the server is listening on (meaningful after
// Listen succeeds). This is the value advertised over mDNS.
func (s *Server) Port() int { return s.port }

// RepoName returns the share name clients use to reach this repo.
func (s *Server) RepoName() string { return s.repoName }

// Listen binds the encrypted transport. It tries preferredPort first and falls
// back to an OS-assigned port if that is taken, so two peers on one host (or a
// lingering socket) never block startup.
func (s *Server) Listen(preferredPort int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", preferredPort))
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("transport listen: %w", err)
		}
	}
	s.ln = ln
	s.port = ln.Addr().(*net.TCPAddr).Port
	return nil
}

// Serve accepts connections until ctx is canceled or the listener closes.
func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.ln.Close()
	}()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // clean shutdown
			default:
				return err
			}
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	ec, peerID, err := e2e.ServerAuth(conn, s.identity)
	if err != nil {
		s.logf("handshake from %s failed: %v", conn.RemoteAddr(), err)
		return
	}
	if s.Verify != nil {
		if err := s.Verify(peerID); err != nil {
			s.logf("rejecting %s: %v", conn.RemoteAddr(), err)
			return
		}
	}
	if err := ServerGate(ec, s.RequireAuth, s.Salt, s.Seed); err != nil {
		s.logf("auth from %s: %v", conn.RemoteAddr(), err)
		return
	}
	if err := s.serveGit(ec); err != nil {
		s.logf("serving %s: %v", conn.RemoteAddr(), err)
	}
}

// serveGit runs a one-shot git daemon whose I/O is the decrypted connection.
func (s *Server) serveGit(ec io.ReadWriter) error {
	// git daemon wants forward slashes for base-path/whitelist even on Windows.
	parent := filepath.ToSlash(filepath.Dir(s.repoRoot))
	args := []string{
		"daemon", "--inetd", "--export-all",
		"--base-path=" + parent,
	}
	if s.allowPush {
		args = append(args, "--enable=receive-pack")
	}
	// Whitelist exactly this repo directory so siblings under the parent are
	// never exposed, even with --export-all.
	args = append(args, filepath.ToSlash(s.repoRoot))

	cmd := exec.Command("git", args...)
	cmd.Stdin = ec
	cmd.Stdout = ec

	s.track(cmd)
	defer s.untrack(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git daemon: %w", err)
	}
	return nil
}

// Shutdown closes the listener and kills any in-flight git daemon subprocesses.
// Safe to call once after Serve; pairs with cancelling the Serve context.
func (s *Server) Shutdown() {
	if s.ln != nil {
		_ = s.ln.Close()
	}
	s.mu.Lock()
	for cmd := range s.subs {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
	s.subs = make(map[*exec.Cmd]struct{})
	s.mu.Unlock()
}

func (s *Server) track(cmd *exec.Cmd) {
	s.mu.Lock()
	s.subs[cmd] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) untrack(cmd *exec.Cmd) {
	s.mu.Lock()
	delete(s.subs, cmd)
	s.mu.Unlock()
}

func (s *Server) logf(format string, args ...any) {
	if s.Log != nil {
		s.Log(format, args...)
	}
}
