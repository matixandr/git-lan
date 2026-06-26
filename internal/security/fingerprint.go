package security

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/matixandr/git-lan/pkg/config"
)

// Identity is this host's long-term X25519 keypair. Its public half identifies
// the host to peers; its fingerprint is what users compare and pin. The private
// half never leaves the machine and is stored 0600.
type Identity struct {
	priv *ecdh.PrivateKey
}

// PublicKey returns the 32-byte X25519 public key.
func (id *Identity) PublicKey() []byte { return id.priv.PublicKey().Bytes() }

// Private returns the underlying private key for use in the authenticated
// handshake (static-ephemeral Diffie-Hellman).
func (id *Identity) Private() *ecdh.PrivateKey { return id.priv }

// Fingerprint returns this identity's fingerprint string, e.g.
// "SHA256:Hk9s...". See FingerprintOf for the format.
func (id *Identity) Fingerprint() string { return FingerprintOf(id.PublicKey()) }

// FingerprintOf renders a public key as a stable, human-comparable fingerprint:
// "SHA256:" followed by the unpadded base64url of SHA-256(pubkey).
func FingerprintOf(pub []byte) string {
	sum := sha256.Sum256(pub)
	return "SHA256:" + base64.RawURLEncoding.EncodeToString(sum[:])
}

// LoadOrCreateIdentity loads the persistent identity key, generating and
// persisting a new one (0600) on first run.
func LoadOrCreateIdentity() (*Identity, error) {
	path, err := config.IdentityPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		return parseIdentity(data)
	case os.IsNotExist(err):
		return generateIdentity(path)
	default:
		return nil, fmt.Errorf("read identity: %w", err)
	}
}

func parseIdentity(raw []byte) (*Identity, error) {
	if len(raw) != 32 {
		return nil, fmt.Errorf("identity key is %d bytes, want 32", len(raw))
	}
	priv, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		return nil, fmt.Errorf("parse identity key: %w", err)
	}
	return &Identity{priv: priv}, nil
}

func generateIdentity(path string) (*Identity, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}
	if err := writeSecretFile(path, priv.Bytes()); err != nil {
		return nil, fmt.Errorf("persist identity: %w", err)
	}
	return &Identity{priv: priv}, nil
}
