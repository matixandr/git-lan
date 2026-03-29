// Package e2e implements the end-to-end encryption layer that wraps every
// peer-to-peer connection in git-lan. No application bytes ever cross the wire
// unencrypted.
//
// The scheme is deliberately small and conventional:
//
//   - Ephemeral X25519 keys are exchanged per connection (forward secrecy).
//   - A shared secret is derived via X25519 ECDH.
//   - HKDF-SHA256 expands it into two *directional* ChaCha20-Poly1305 keys, so
//     compromise of one direction's key does not affect the other.
//   - Each frame carries a 12-byte nonce: a 32-bit random prefix fixed at
//     handshake time plus a 64-bit monotonically increasing counter.
//   - Receivers reject any nonce that does not strictly advance, defeating
//     replay and reordering.
package e2e

import "errors"

// HandshakeLabel components fed to HKDF. The directional labels guarantee the
// two keys differ even though they share one ECDH secret.
const (
	infoAtoB = "git-lan-v1-atob"
	infoBtoA = "git-lan-v1-btoa"
)

// Frame and key sizes.
const (
	KeySize       = 32 // ChaCha20-Poly1305 key
	NonceSize     = 12 // ChaCha20-Poly1305 nonce
	TagSize       = 16 // Poly1305 auth tag
	LenPrefixSize = 4  // big-endian payload length
	// MaxFrameSize bounds a single encrypted frame's plaintext to keep memory
	// bounded against a hostile or buggy peer.
	MaxFrameSize = 1 << 20 // 1 MiB
)

var (
	// ErrReplay is returned when a frame's nonce does not strictly advance.
	ErrReplay = errors.New("e2e: replayed or out-of-order nonce")
	// ErrFrameTooLarge is returned when a peer announces an oversized frame.
	ErrFrameTooLarge = errors.New("e2e: frame exceeds maximum size")
	// ErrHandshake is returned when the handshake fails or is malformed.
	ErrHandshake = errors.New("e2e: handshake failed")
	// ErrFingerprintMismatch indicates a trusted peer presented an identity
	// key that does not match its pinned fingerprint - a possible MITM.
	ErrFingerprintMismatch = errors.New("e2e: peer identity does not match pinned fingerprint")
)
