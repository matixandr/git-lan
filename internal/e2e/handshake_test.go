package e2e

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// loopback returns a connected pair of TCP sockets on the loopback interface.
// Unlike net.Pipe these are buffered, matching the real transport: the
// handshake writes both public keys before either side reads.
func loopback(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	type ar struct {
		c   net.Conn
		err error
	}
	acceptCh := make(chan ar, 1)
	go func() {
		c, err := ln.Accept()
		acceptCh <- ar{c, err}
	}()
	dialed, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	got := <-acceptCh
	if got.err != nil {
		t.Fatalf("accept: %v", got.err)
	}
	return dialed, got.c
}

// runHandshake drives both ends of a handshake over a loopback socket and
// returns the initiator's and responder's negotiated keys.
func runHandshake(t *testing.T) (a, b *sessionKeys) {
	t.Helper()
	c1, c2 := loopback(t)
	defer c1.Close()
	defer c2.Close()

	type res struct {
		keys *sessionKeys
		err  error
	}
	ach := make(chan res, 1)
	bch := make(chan res, 1)
	go func() { k, err := handshake(c1, true); ach <- res{k, err} }()
	go func() { k, err := handshake(c2, false); bch <- res{k, err} }()

	var ar, br res
	for i := 0; i < 2; i++ {
		select {
		case ar = <-ach:
		case br = <-bch:
		case <-time.After(2 * time.Second):
			t.Fatal("handshake timed out")
		}
	}
	if ar.err != nil || br.err != nil {
		t.Fatalf("handshake errors: a=%v b=%v", ar.err, br.err)
	}
	return ar.keys, br.keys
}

func TestHandshakeDirectionalKeysMatch(t *testing.T) {
	a, b := runHandshake(t)

	// The initiator's send key must equal the responder's receive key, and
	// vice versa - that is what makes each direction independently keyed.
	if !bytes.Equal(a.send, b.recv) {
		t.Error("A.send != B.recv")
	}
	if !bytes.Equal(a.recv, b.send) {
		t.Error("A.recv != B.send")
	}
	// The two directions must use distinct keys.
	if bytes.Equal(a.send, a.recv) {
		t.Error("directional keys are identical; HKDF labels not applied")
	}
	if len(a.send) != KeySize {
		t.Errorf("key size = %d, want %d", len(a.send), KeySize)
	}
}
