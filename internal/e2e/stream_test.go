package e2e

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptedConn must be a drop-in net.Conn.
var _ net.Conn = (*EncryptedConn)(nil)

// pair establishes an encrypted connection over loopback and returns both ends.
func pair(t *testing.T) (*EncryptedConn, *EncryptedConn) {
	t.Helper()
	c1, c2 := loopback(t)

	type res struct {
		ec  *EncryptedConn
		err error
	}
	ch := make(chan res, 1)
	go func() { ec, err := Server(c2); ch <- res{ec, err} }()
	client, err := Client(c1)
	if err != nil {
		t.Fatalf("client handshake: %v", err)
	}
	srv := <-ch
	if srv.err != nil {
		t.Fatalf("server handshake: %v", srv.err)
	}
	return client, srv.ec
}

func TestEncryptedRoundTrip(t *testing.T) {
	client, server := pair(t)
	defer client.Close()
	defer server.Close()

	msg := []byte("git upload-pack: the quick brown fox\x00\x01\x02")
	go func() { _, _ = client.Write(msg) }()

	got := make([]byte, len(msg))
	if _, err := io.ReadFull(server, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatalf("round-trip mismatch:\n got %q\nwant %q", got, msg)
	}
}

// A large payload must cross intact through multiple frames and partial reads.
func TestEncryptedLargePayload(t *testing.T) {
	client, server := pair(t)
	defer client.Close()
	defer server.Close()

	payload := make([]byte, 5*MaxFrameSize+1234)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := client.Write(payload); err != nil {
			t.Errorf("write: %v", err)
		}
	}()

	got := make([]byte, len(payload))
	if _, err := io.ReadFull(server, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	wg.Wait()
	if !bytes.Equal(got, payload) {
		t.Fatal("large payload corrupted in transit")
	}
}

// A flipped ciphertext byte must fail authentication, not silently pass.
func TestTamperedFrameRejected(t *testing.T) {
	// Seal a frame with a sender, then hand a corrupted copy to a receiver
	// keyed identically, and confirm the parser rejects it.
	send, recv := keyedPair(t)

	var buf bytes.Buffer
	send.conn = &bufConn{buf: &buf}
	if err := send.writeFrame([]byte("authentic data")); err != nil {
		t.Fatal(err)
	}
	frame := buf.Bytes()
	frame[len(frame)-1] ^= 0x01 // corrupt the auth tag

	recv.conn = &bufConn{buf: bytes.NewBuffer(frame)}
	if _, err := recv.readFrame(); err == nil {
		t.Fatal("tampered frame accepted")
	}
}

// keyedPair returns two EncryptedConns sharing one key in each direction, so a
// frame sealed by the first can be opened by the second without a live socket.
func keyedPair(t *testing.T) (send, recv *EncryptedConn) {
	t.Helper()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	mk := func() *EncryptedConn {
		aead, err := chacha20poly1305.New(key)
		if err != nil {
			t.Fatal(err)
		}
		seq, _ := newNonceSequencer()
		return &EncryptedConn{sendAEAD: aead, recvAEAD: aead, seq: seq}
	}
	return mk(), mk()
}

// bufConn is a minimal net.Conn backed by a bytes.Buffer for offline framing
// tests. Only Read and Write carry behaviour.
type bufConn struct{ buf *bytes.Buffer }

func (b *bufConn) Read(p []byte) (int, error)       { return b.buf.Read(p) }
func (b *bufConn) Write(p []byte) (int, error)      { return b.buf.Write(p) }
func (b *bufConn) Close() error                     { return nil }
func (b *bufConn) LocalAddr() net.Addr              { return nil }
func (b *bufConn) RemoteAddr() net.Addr             { return nil }
func (b *bufConn) SetDeadline(time.Time) error      { return nil }
func (b *bufConn) SetReadDeadline(time.Time) error  { return nil }
func (b *bufConn) SetWriteDeadline(time.Time) error { return nil }
