package session

import (
	"errors"
	"testing"
	"time"
)

func TestBase58RoundTrip(t *testing.T) {
	inputs := [][]byte{
		{0, 0, 1, 2, 3},
		{255, 254, 253},
		[]byte("git-lan invite payload bytes!!"),
	}
	for _, in := range inputs {
		enc := base58Encode(in)
		out, ok := base58Decode(enc)
		if !ok || string(out) != string(in) {
			t.Errorf("round trip failed for %v: got %v ok=%v", in, out, ok)
		}
	}
}

func TestInviteGenerateParse(t *testing.T) {
	secret := []byte("session-secret-key")
	token, id, err := GenerateInvite(secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseInvite(secret, token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got != id {
		t.Error("parsed invite ID does not match generated ID")
	}
}

func TestInviteWrongSecretRejected(t *testing.T) {
	token, _, _ := GenerateInvite([]byte("real"), time.Hour)
	if _, err := ParseInvite([]byte("forged"), token); !errors.Is(err, ErrInviteInvalid) {
		t.Errorf("expected ErrInviteInvalid, got %v", err)
	}
}

func TestInviteExpiry(t *testing.T) {
	token, _, _ := GenerateInvite([]byte("k"), -time.Second)
	if _, err := ParseInvite([]byte("k"), token); !errors.Is(err, ErrInviteExpired) {
		t.Errorf("expected ErrInviteExpired, got %v", err)
	}
}

func TestInviteTamperRejected(t *testing.T) {
	token, _, _ := GenerateInvite([]byte("k"), time.Hour)
	// Flip a character in the token body.
	b := []byte(token)
	b[0] = flip58(b[0])
	if _, err := ParseInvite([]byte("k"), string(b)); err == nil {
		t.Error("tampered token accepted")
	}
}

func flip58(c byte) byte {
	if c == '2' {
		return '3'
	}
	return '2'
}
