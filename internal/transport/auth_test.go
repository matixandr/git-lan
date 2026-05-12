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
func gateExchange(t *testing.T, srv ServerGateConfig, cli ClientGateConfig) error {
	t.Helper()
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	srvErr := make(chan error, 1)
	go func() { srvErr <- ServerGate(c1, srv) }()
	cliErr := make(chan error, 1)
	go func() {
		err := ClientGate(c2, cli)
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
	if err := gateExchange(t, ServerGateConfig{RequireAuth: false}, ClientGateConfig{}); err != nil {
		t.Fatalf("open session gate errored: %v", err)
	}
}

func TestGateCorrectPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	seed := []byte("the-shared-argon2id-seed-32bytes")
	srv := ServerGateConfig{RequireAuth: true, Salt: salt, Seed: seed}
	cli := ClientGateConfig{Derive: func(s []byte) []byte {
		if !bytes.Equal(s, salt) {
			t.Errorf("client received wrong salt")
		}
		return seed
	}}
	if err := gateExchange(t, srv, cli); err != nil {
		t.Fatalf("correct password rejected: %v", err)
	}
}

func TestGateWrongPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	seed := []byte("the-shared-argon2id-seed-32bytes")
	srv := ServerGateConfig{RequireAuth: true, Salt: salt, Seed: seed}
	cli := ClientGateConfig{Derive: func([]byte) []byte { return []byte("a-different-wrong-seed-value-xx!") }}
	if err := gateExchange(t, srv, cli); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("wrong password should fail with ErrAuthFailed, got %v", err)
	}
}

func TestGateMissingCredential(t *testing.T) {
	srv := ServerGateConfig{RequireAuth: true, Salt: make([]byte, 16), Seed: []byte("seed")}
	if err := gateExchange(t, srv, ClientGateConfig{}); err == nil {
		t.Fatal("protected session admitted a client with no credential")
	}
}

func TestGateValidInvite(t *testing.T) {
	called := false
	srv := ServerGateConfig{
		RequireAuth: true,
		Salt:        make([]byte, 16),
		Invite: func(token string) error {
			called = true
			if token != "GOODTOKEN" {
				return ErrAuthFailed
			}
			return nil
		},
	}
	if err := gateExchange(t, srv, ClientGateConfig{Token: "GOODTOKEN"}); err != nil {
		t.Fatalf("valid invite rejected: %v", err)
	}
	if !called {
		t.Error("invite validator was not called")
	}
}

func TestGateInvalidInvite(t *testing.T) {
	srv := ServerGateConfig{
		RequireAuth: true,
		Salt:        make([]byte, 16),
		Invite:      func(string) error { return ErrAuthFailed },
	}
	if err := gateExchange(t, srv, ClientGateConfig{Token: "BADTOKEN"}); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("invalid invite should fail, got %v", err)
	}
}
