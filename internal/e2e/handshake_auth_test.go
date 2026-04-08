package e2e

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"io"
	"testing"
)

func newID(t *testing.T) *ecdh.PrivateKey {
	t.Helper()
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestAuthenticatedHandshakeRoundTrip(t *testing.T) {
	c1, c2 := loopback(t)
	defer c1.Close()
	defer c2.Close()

	idA, idB := newID(t), newID(t)

	type res struct {
		ec     *EncryptedConn
		peerID []byte
		err    error
	}
	ch := make(chan res, 1)
	go func() {
		ec, pid, err := ServerAuth(c2, idB)
		ch <- res{ec, pid, err}
	}()
	clientEC, serverIDSeen, err := ClientAuth(c1, idA)
	if err != nil {
		t.Fatalf("client auth: %v", err)
	}
	srv := <-ch
	if srv.err != nil {
		t.Fatalf("server auth: %v", srv.err)
	}

	// Each side must observe the other's true identity public key.
	if !bytes.Equal(serverIDSeen, idB.PublicKey().Bytes()) {
		t.Error("client saw wrong server identity")
	}
	if !bytes.Equal(srv.peerID, idA.PublicKey().Bytes()) {
		t.Error("server saw wrong client identity")
	}

	// And the encrypted channel must actually work end to end.
	msg := []byte("authenticated channel up")
	go func() { _, _ = clientEC.Write(msg) }()
	got := make([]byte, len(msg))
	if _, err := io.ReadFull(srv.ec, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatal("authenticated round-trip mismatch")
	}
}
