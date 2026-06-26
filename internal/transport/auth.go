package transport

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"io"
)

// Session password gate. After the E2E handshake establishes a confidential,
// identity-authenticated channel, a password-protected session runs a short
// challenge/response *inside* that channel before any git protocol bytes flow.
// The shared secret is an Argon2id seed derived from the password; the server
// never sees the password and a wrong password simply fails the HMAC check.
const (
	authNone     byte = 0 // session is open
	authPassword byte = 1 // session requires the password seed

	authSaltLen      = 16
	authChallengeLen = 32
	authResponseLen  = 32
)

// ErrAuthFailed means the peer did not prove knowledge of the session password.
var ErrAuthFailed = errors.New("session password authentication failed")

// ServerGate runs the server side of the password gate over the encrypted
// connection. When requireAuth is false it advertises an open session and
// returns immediately. seed is the stored Argon2id seed; salt is the salt used
// to derive it, sent to the client so it can derive the same seed.
func ServerGate(rw io.ReadWriter, requireAuth bool, salt, seed []byte) error {
	if !requireAuth {
		_, err := rw.Write([]byte{authNone})
		return err
	}
	challenge := make([]byte, authChallengeLen)
	if _, err := rand.Read(challenge); err != nil {
		return err
	}
	msg := make([]byte, 0, 1+authSaltLen+authChallengeLen)
	msg = append(msg, authPassword)
	msg = append(msg, salt...)
	msg = append(msg, challenge...)
	if _, err := rw.Write(msg); err != nil {
		return err
	}

	resp := make([]byte, authResponseLen)
	if _, err := io.ReadFull(rw, resp); err != nil {
		return err
	}
	want := authMAC(seed, challenge)
	if subtle.ConstantTimeCompare(resp, want) != 1 {
		return ErrAuthFailed
	}
	return nil
}

// ClientGate runs the client side of the gate. derive turns the session salt
// into the Argon2id seed (it closes over the user's password); it may be nil
// when the user supplied no password, in which case a locked session fails.
func ClientGate(rw io.ReadWriter, derive func(salt []byte) []byte) error {
	mode := make([]byte, 1)
	if _, err := io.ReadFull(rw, mode); err != nil {
		return err
	}
	if mode[0] == authNone {
		return nil
	}
	salt := make([]byte, authSaltLen)
	if _, err := io.ReadFull(rw, salt); err != nil {
		return err
	}
	challenge := make([]byte, authChallengeLen)
	if _, err := io.ReadFull(rw, challenge); err != nil {
		return err
	}
	if derive == nil {
		return errors.New("session is password-protected; use --password")
	}
	seed := derive(salt)
	if _, err := rw.Write(authMAC(seed, challenge)); err != nil {
		return err
	}
	return nil
}

func authMAC(seed, challenge []byte) []byte {
	h := hmac.New(sha256.New, seed)
	h.Write(challenge)
	return h.Sum(nil)[:authResponseLen]
}
