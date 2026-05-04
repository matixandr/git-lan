// Package session manages git-lan collaboration sessions: their lifecycle,
// password protection, one-time invite tokens, and on-disk persistence.
package session

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters. Chosen to be interactive-fast yet memory-hard; tuned for
// a developer laptop, not a password-cracking rig.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// HashPassword derives an Argon2id hash and returns it in PHC string format,
// e.g. "$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>". Safe to store on disk.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		b64(salt), b64(hash)), nil
}

// VerifyPassword reports whether password matches a PHC-encoded Argon2id hash.
// Comparison is constant-time.
func VerifyPassword(password, encoded string) bool {
	salt, want, ok := decodePHC(encoded)
	if !ok {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// DeriveSeed derives a stable 32-byte key seed from a password and salt. The
// session uses it as a pre-shared component when present, layered on top of the
// per-connection ephemeral handshake.
func DeriveSeed(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

func b64(b []byte) string { return base64.RawStdEncoding.EncodeToString(b) }

func decodePHC(encoded string) (salt, hash []byte, ok bool) {
	parts := strings.Split(encoded, "$")
	// ["", "argon2id", "v=19", "m=..,t=..,p=..", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, false
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, false
	}
	return salt, hash, true
}
