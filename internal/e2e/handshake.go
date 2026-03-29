package e2e

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// sessionKeys holds the two directional ChaCha20-Poly1305 keys negotiated for
// a single connection.
type sessionKeys struct {
	send []byte // key this side encrypts with
	recv []byte // key this side decrypts with
}

// handshake performs an ephemeral X25519 ECDH over rw and derives directional
// session keys. The initiator (dialer) takes the "A" role; the responder takes
// "B". Each side sends its ephemeral public key in the clear - that is safe for
// ECDH - then both compute the same shared secret and expand it with HKDF.
func handshake(rw io.ReadWriter, initiator bool) (*sessionKeys, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("%w: generate ephemeral key: %v", ErrHandshake, err)
	}
	myPub := priv.PublicKey().Bytes()

	peerPub, err := exchange(rw, myPub)
	if err != nil {
		return nil, err
	}

	peerKey, err := ecdh.X25519().NewPublicKey(peerPub)
	if err != nil {
		return nil, fmt.Errorf("%w: bad peer key: %v", ErrHandshake, err)
	}
	shared, err := priv.ECDH(peerKey)
	if err != nil {
		return nil, fmt.Errorf("%w: ecdh: %v", ErrHandshake, err)
	}

	// Salt binds the derivation to this exact pair of ephemeral keys in a
	// canonical (order-independent) way, so both peers compute identical salt.
	salt := transcriptSalt(myPub, peerPub)
	keyAtoB := deriveKey(shared, salt, infoAtoB)
	keyBtoA := deriveKey(shared, salt, infoBtoA)

	if initiator {
		return &sessionKeys{send: keyAtoB, recv: keyBtoA}, nil
	}
	return &sessionKeys{send: keyBtoA, recv: keyAtoB}, nil
}

// exchange writes our public key and reads the peer's. Both are fixed 32 bytes.
func exchange(rw io.ReadWriter, myPub []byte) ([]byte, error) {
	if _, err := rw.Write(myPub); err != nil {
		return nil, fmt.Errorf("%w: send pubkey: %v", ErrHandshake, err)
	}
	peerPub := make([]byte, len(myPub))
	if _, err := io.ReadFull(rw, peerPub); err != nil {
		return nil, fmt.Errorf("%w: read peer pubkey: %v", ErrHandshake, err)
	}
	return peerPub, nil
}

// transcriptSalt returns SHA-256 over the two public keys in sorted order so
// both ends derive the same salt regardless of who is A or B.
func transcriptSalt(a, b []byte) []byte {
	lo, hi := a, b
	if bytes.Compare(a, b) > 0 {
		lo, hi = b, a
	}
	h := sha256.New()
	h.Write(lo)
	h.Write(hi)
	return h.Sum(nil)
}

// deriveKey expands the shared secret into one KeySize key bound to info.
func deriveKey(shared, salt []byte, info string) []byte {
	r := hkdf.New(sha256.New, shared, salt, []byte(info))
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(r, key); err != nil {
		panic("e2e: hkdf read failed: " + err.Error()) // never happens for SHA-256
	}
	return key
}
