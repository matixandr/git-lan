package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"time"
)

// Invite token layout, before base58 encoding:
//
//	[16 bytes random token ID][8 bytes big-endian expiry unix seconds][16 bytes HMAC]
//
// The HMAC (truncated SHA-256 over ID+expiry, keyed by the session secret)
// makes tokens unforgeable; the random ID lets the session burn a token after
// first use so it cannot be replayed.
const (
	inviteIDLen   = 16
	inviteExpLen  = 8
	inviteMACLen  = 16
	invitePayload = inviteIDLen + inviteExpLen
)

var (
	// ErrInviteInvalid means the token is malformed or its HMAC failed.
	ErrInviteInvalid = errors.New("invite token is invalid")
	// ErrInviteExpired means the token's expiry is in the past.
	ErrInviteExpired = errors.New("invite token has expired")
)

// InviteID identifies a single token, used to enforce one-time use.
type InviteID [inviteIDLen]byte

// GenerateInvite mints a signed, expiring one-time token for a session secret.
func GenerateInvite(secret []byte, ttl time.Duration) (token string, id InviteID, err error) {
	if _, err = rand.Read(id[:]); err != nil {
		return "", id, err
	}
	exp := time.Now().Add(ttl).Unix()

	payload := make([]byte, invitePayload)
	copy(payload[:inviteIDLen], id[:])
	binary.BigEndian.PutUint64(payload[inviteIDLen:], uint64(exp))

	mac := inviteMAC(secret, payload)
	raw := append(payload, mac...)
	return base58Encode(raw), id, nil
}

// ParseInvite verifies a token against the session secret and checks expiry. It
// returns the token's ID so the caller can confirm it has not been burned. The
// HMAC check is constant-time.
func ParseInvite(secret []byte, token string) (InviteID, error) {
	var id InviteID
	raw, ok := base58Decode(token)
	if !ok || len(raw) != invitePayload+inviteMACLen {
		return id, ErrInviteInvalid
	}
	payload := raw[:invitePayload]
	mac := raw[invitePayload:]

	want := inviteMAC(secret, payload)
	if subtle.ConstantTimeCompare(mac, want) != 1 {
		return id, ErrInviteInvalid
	}

	exp := int64(binary.BigEndian.Uint64(payload[inviteIDLen:]))
	if time.Now().Unix() > exp {
		return id, ErrInviteExpired
	}

	copy(id[:], payload[:inviteIDLen])
	return id, nil
}

func inviteMAC(secret, payload []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return h.Sum(nil)[:inviteMACLen]
}
