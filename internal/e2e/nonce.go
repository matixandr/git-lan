package e2e

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// nonceSequencer produces the 12-byte ChaCha20-Poly1305 nonces for one
// direction of a connection: a 4-byte random prefix fixed at construction plus
// a 64-bit big-endian counter that increments per frame. The random prefix
// makes nonces unpredictable; the counter makes them unique.
type nonceSequencer struct {
	prefix  [4]byte
	counter uint64
}

func newNonceSequencer() (*nonceSequencer, error) {
	s := &nonceSequencer{}
	if _, err := rand.Read(s.prefix[:]); err != nil {
		return nil, fmt.Errorf("%w: nonce prefix: %v", ErrHandshake, err)
	}
	return s, nil
}

// next returns the nonce for the next frame and advances the counter. It
// returns an error if the counter would wrap, which would catastrophically
// reuse a nonce - the caller must tear the connection down instead.
func (s *nonceSequencer) next() ([]byte, error) {
	if s.counter == ^uint64(0) {
		return nil, fmt.Errorf("e2e: nonce counter exhausted")
	}
	nonce := make([]byte, NonceSize)
	copy(nonce[:4], s.prefix[:])
	binary.BigEndian.PutUint64(nonce[4:], s.counter)
	s.counter++
	return nonce, nil
}

// replayGuard enforces that received nonces strictly advance. Because a given
// peer's prefix is constant and its counter increments, treating the whole
// 12-byte nonce as a big-endian integer gives a strict ordering; any frame
// whose nonce does not exceed the last accepted one is a replay or reorder.
type replayGuard struct {
	have bool
	last [NonceSize]byte
}

func (g *replayGuard) accept(nonce []byte) error {
	if len(nonce) != NonceSize {
		return ErrReplay
	}
	var n [NonceSize]byte
	copy(n[:], nonce)
	if g.have && !greater(n, g.last) {
		return ErrReplay
	}
	g.last = n
	g.have = true
	return nil
}

// greater reports whether a > b as 96-bit big-endian integers.
func greater(a, b [NonceSize]byte) bool {
	for i := 0; i < NonceSize; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false // equal is not greater
}
