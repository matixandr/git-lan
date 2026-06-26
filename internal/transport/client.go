package transport

import (
	"crypto/ecdh"
	"fmt"
	"net"

	"github.com/matixandr/git-lan/internal/e2e"
)

// Client dials a peer's encrypted transport. It exposes a local loopback bridge
// that the git binary connects to in plaintext; every byte the bridge forwards
// to the peer is encrypted.
type Client struct {
	Identity *ecdh.PrivateKey
	PeerAddr string // host:port of the peer's encrypted listener

	// Verify checks the peer identity after each handshake; nil = TOFU.
	Verify VerifyFunc
	// DeriveSeed turns a session salt into the password seed for the gate. Nil
	// when the user supplied no password; a locked session then fails cleanly.
	DeriveSeed func(salt []byte) []byte
	// Log receives diagnostics; may be nil.
	Log func(format string, args ...any)
}

// Bridge starts a loopback endpoint tunneling to the peer and returns the
// git:// URL git should use for repo, plus a stop function to call once the git
// operation completes. Each git connection to the bridge triggers its own
// authenticated handshake with the peer.
func (c *Client) Bridge(repo string) (url string, stop func(), err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("bridge listen: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	done := make(chan struct{})
	go c.acceptLoop(ln, done)

	stop = func() {
		close(done)
		_ = ln.Close()
	}
	return gitDaemonURL("127.0.0.1", port, repo), stop, nil
}

func (c *Client) acceptLoop(ln net.Listener, done <-chan struct{}) {
	for {
		gitConn, err := ln.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
				c.logf("bridge accept: %v", err)
				return
			}
		}
		go c.tunnel(gitConn)
	}
}

// tunnel connects git's local plaintext socket to the peer over an encrypted,
// authenticated channel.
func (c *Client) tunnel(gitConn net.Conn) {
	raw, err := net.Dial("tcp", c.PeerAddr)
	if err != nil {
		c.logf("dial peer %s: %v", c.PeerAddr, err)
		_ = gitConn.Close()
		return
	}

	ec, peerID, err := e2e.ClientAuth(raw, c.Identity)
	if err != nil {
		c.logf("handshake with %s: %v", c.PeerAddr, err)
		_ = raw.Close()
		_ = gitConn.Close()
		return
	}
	if c.Verify != nil {
		if err := c.Verify(peerID); err != nil {
			c.logf("peer verification failed: %v", err)
			_ = ec.Close()
			_ = gitConn.Close()
			return
		}
	}
	if err := ClientGate(ec, c.DeriveSeed); err != nil {
		c.logf("session auth failed: %v", err)
		_ = ec.Close()
		_ = gitConn.Close()
		return
	}

	duplex(gitConn, ec)
}

func (c *Client) logf(format string, args ...any) {
	if c.Log != nil {
		c.Log(format, args...)
	}
}
