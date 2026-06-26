package transport

import (
	"io"
	"net"
	"sync"
)

// VerifyFunc inspects a peer's long-term identity public key immediately after
// the handshake and returns a non-nil error to abort the connection (e.g. a
// fingerprint mismatch against a pinned trust entry). It is the single hook the
// trust ring plugs into; a nil VerifyFunc means trust-on-first-use.
type VerifyFunc func(peerIdentity []byte) error

// duplex copies bytes in both directions between a and b until either side
// closes or errors, then tears both down. It returns when both copy directions
// have finished. This is the bridge between a decrypted EncryptedConn and the
// local plaintext git socket.
func duplex(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	cp := func(dst, src net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(dst, src)
		// Unblock the opposite direction by closing both ends.
		_ = dst.Close()
		_ = src.Close()
	}

	go cp(a, b)
	go cp(b, a)
	wg.Wait()
}
