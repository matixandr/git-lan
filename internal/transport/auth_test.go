package transport

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"
)

// gateExchange runs ServerGate and ClientGate against each other over a pipe
// and returns the server-side error (the one that decides admission).
func gateExchange(t *testing.T, requireAuth bool, salt, seed []byte, derive func([]byte) []byte) error {
	t.Helper()
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	srvErr := make(chan error, 1)
	go func() { srvErr <- ServerGate(c1, requireAuth, salt, seed) }()
	cliErr := make(chan error, 1)
	go func() {
		err := ClientGate(c2, derive)
		// Mirror production teardown: a client that cannot authenticate closes
		// the connection, which unblocks the server's pending read.
		if err != nil {
			_ = c2.Close()
		}
		cliErr <- err
	}()

	var got error
	select {
	case got = <-srvErr:
	case <-time.After(2 * time.Second):
		t.Fatal("gate timed out")
	}
	<-cliErr
	return got
}

func TestGateOpenSession(t *testing.T) {
	if err := gateExchange(t, false, nil, nil, nil); err != nil {
		t.Fatalf("open session gate errored: %v", err)
	}
}

func TestGateCorrectPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	seed := []byte("the-shared-argon2id-seed-32bytes")
	derive := func(s []byte) []byte {
		if !bytes.Equal(s, salt) {
			t.Errorf("client received wrong salt")
		}
		return seed
	}
	if err := gateExchange(t, true, salt, seed, derive); err != nil {
		t.Fatalf("correct password rejected: %v", err)
	}
}

func TestGateWrongPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	seed := []byte("the-shared-argon2id-seed-32bytes")
	derive := func([]byte) []byte { return []byte("a-different-wrong-seed-value-xx!") }
	if err := gateExchange(t, true, salt, seed, derive); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("wrong password should fail with ErrAuthFailed, got %v", err)
	}
}

func TestGateMissingPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	seed := []byte("the-shared-argon2id-seed-32bytes")
	// Client has no derive function: locked session must not admit it.
	err := gateExchange(t, true, salt, seed, nil)
	if err == nil {
		t.Fatal("locked session admitted a client with no password")
	}
}
