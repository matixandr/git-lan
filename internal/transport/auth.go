package transport

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"io"
)

// Session access gate. After the E2E handshake establishes a confidential,
// identity-authenticated channel, a protected session runs a short
// challenge/response *inside* that channel before any git protocol bytes flow.
//
// A client may satisfy the gate two ways:
//
//   - password: proves knowledge of the Argon2id seed via HMAC(seed, challenge)
//   - invite:   presents a one-time, HMAC-signed token the host validates+burns
//
// The server never sees the password, and a wrong credential fails closed.
const (
	gateOpen      byte = 0 // session is open to anyone on the LAN
	gateProtected byte = 1 // session requires a credential

	methodPassword byte = 'P'
	methodInvite   byte = 'I'

	gateSaltLen      = 16
	gateChallengeLen = 32
	gateMACLen       = 32
	maxInviteLen     = 256
)

// ErrAuthFailed means the peer did not present a valid credential.
var ErrAuthFailed = errors.New("session authentication failed")

// InviteValidator validates and burns a one-time invite token. It returns nil
// to admit the peer. Supplied by the session layer (it holds the secret and the
// burned-token set).
type InviteValidator func(token string) error

// ServerGateConfig configures the server side of the gate.
type ServerGateConfig struct {
	RequireAuth bool
	Salt        []byte
	Seed        []byte           // password seed; may be nil if only invites
	Invite      InviteValidator  // may be nil if invites are not accepted
}

// ServerGate runs the server side of the access gate over the encrypted
// connection.
func ServerGate(rw io.ReadWriter, cfg ServerGateConfig) error {
	if !cfg.RequireAuth {
		_, err := rw.Write([]byte{gateOpen})
		return err
	}

	challenge := make([]byte, gateChallengeLen)
	if _, err := rand.Read(challenge); err != nil {
		return err
	}
	msg := make([]byte, 0, 1+gateSaltLen+gateChallengeLen)
	msg = append(msg, gateProtected)
	msg = append(msg, cfg.Salt...)
	msg = append(msg, challenge...)
	if _, err := rw.Write(msg); err != nil {
		return err
	}

	method := make([]byte, 1)
	if _, err := io.ReadFull(rw, method); err != nil {
		return err
	}
	switch method[0] {
	case methodPassword:
		resp := make([]byte, gateMACLen)
		if _, err := io.ReadFull(rw, resp); err != nil {
			return err
		}
		if cfg.Seed == nil || subtle.ConstantTimeCompare(resp, gateMAC(cfg.Seed, challenge)) != 1 {
			return ErrAuthFailed
		}
		return nil
	case methodInvite:
		token, err := readInvite(rw)
		if err != nil {
			return err
		}
		if cfg.Invite == nil {
			return ErrAuthFailed
		}
		if err := cfg.Invite(token); err != nil {
			return ErrAuthFailed
		}
		return nil
	default:
		return ErrAuthFailed
	}
}

// ClientGateConfig configures the client side. Exactly one credential is used:
// an invite token if non-empty, otherwise the password seed via Derive.
type ClientGateConfig struct {
	Token  string
	Derive func(salt []byte) []byte
}

// ClientGate runs the client side of the gate.
func ClientGate(rw io.ReadWriter, cfg ClientGateConfig) error {
	mode := make([]byte, 1)
	if _, err := io.ReadFull(rw, mode); err != nil {
		return err
	}
	if mode[0] == gateOpen {
		return nil
	}

	salt := make([]byte, gateSaltLen)
	if _, err := io.ReadFull(rw, salt); err != nil {
		return err
	}
	challenge := make([]byte, gateChallengeLen)
	if _, err := io.ReadFull(rw, challenge); err != nil {
		return err
	}

	switch {
	case cfg.Token != "":
		return writeInvite(rw, cfg.Token)
	case cfg.Derive != nil:
		seed := cfg.Derive(salt)
		out := append([]byte{methodPassword}, gateMAC(seed, challenge)...)
		_, err := rw.Write(out)
		return err
	default:
		return errors.New("session is protected; supply --password or --token")
	}
}

func writeInvite(rw io.ReadWriter, token string) error {
	if len(token) > maxInviteLen {
		return errors.New("invite token too long")
	}
	buf := make([]byte, 0, 3+len(token))
	buf = append(buf, methodInvite)
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(len(token)))
	buf = append(buf, l[:]...)
	buf = append(buf, token...)
	_, err := rw.Write(buf)
	return err
}

func readInvite(rw io.ReadWriter) (string, error) {
	var l [2]byte
	if _, err := io.ReadFull(rw, l[:]); err != nil {
		return "", err
	}
	n := int(binary.BigEndian.Uint16(l[:]))
	if n == 0 || n > maxInviteLen {
		return "", ErrAuthFailed
	}
	tok := make([]byte, n)
	if _, err := io.ReadFull(rw, tok); err != nil {
		return "", err
	}
	return string(tok), nil
}

func gateMAC(seed, challenge []byte) []byte {
	h := hmac.New(sha256.New, seed)
	h.Write(challenge)
	return h.Sum(nil)[:gateMACLen]
}
