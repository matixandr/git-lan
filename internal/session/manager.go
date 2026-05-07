package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/matixandr/git-lan/pkg/config"
)

// Session is a hosted collaboration session: the repo a host shares plus its
// access policy. Secrets here never leave the machine and never touch mDNS.
type Session struct {
	Name         string    `json:"name"`
	RepoRoot     string    `json:"repo_root"`
	PasswordHash string    `json:"password_hash,omitempty"`
	Secret       []byte    `json:"secret"` // HMAC key for invite tokens
	Salt         []byte    `json:"salt"`   // password key-seed salt
	AllowPush    bool      `json:"allow_push"`
	Port         int       `json:"port"`
	CreatedAt    time.Time `json:"created_at"`
	Burned       []string  `json:"burned_invites,omitempty"`
}

// Store is the persisted session state. At most one session is hosted locally
// at a time; Active is nil when none is running.
type Store struct {
	Active *Session `json:"active"`
}

// ErrNoSession indicates no session is currently active.
var ErrNoSession = errors.New("no active session")

// New creates a session for repoRoot. An empty password leaves the session
// open (no lock). Secret and salt are freshly random.
func New(name, password, repoRoot string, allowPush bool) (*Session, error) {
	s := &Session{
		Name:      name,
		RepoRoot:  repoRoot,
		AllowPush: allowPush,
		CreatedAt: time.Now(),
	}
	s.Secret = make([]byte, 32)
	s.Salt = make([]byte, 16)
	if _, err := rand.Read(s.Secret); err != nil {
		return nil, err
	}
	if _, err := rand.Read(s.Salt); err != nil {
		return nil, err
	}
	if password != "" {
		h, err := HashPassword(password)
		if err != nil {
			return nil, err
		}
		s.PasswordHash = h
	}
	return s, nil
}

// HasPassword reports whether the session is password-protected.
func (s *Session) HasPassword() bool { return s.PasswordHash != "" }

// CheckPassword verifies a candidate password against the session.
func (s *Session) CheckPassword(password string) bool {
	if !s.HasPassword() {
		return true
	}
	return VerifyPassword(password, s.PasswordHash)
}

// Burn records an invite ID as used. Returns false if it was already burned.
func (s *Session) Burn(id InviteID) bool {
	key := base64.RawStdEncoding.EncodeToString(id[:])
	for _, b := range s.Burned {
		if b == key {
			return false
		}
	}
	s.Burned = append(s.Burned, key)
	return true
}

// IsBurned reports whether an invite ID has already been used.
func (s *Session) IsBurned(id InviteID) bool {
	key := base64.RawStdEncoding.EncodeToString(id[:])
	for _, b := range s.Burned {
		if b == key {
			return true
		}
	}
	return false
}

// Load reads sessions.json, returning an empty store if absent.
func Load() (*Store, error) {
	path, err := config.SessionsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{}, nil
	} else if err != nil {
		return nil, err
	}
	var st Store
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// Save persists the store to sessions.json with owner-only permissions (it
// contains the invite-signing secret).
func (st *Store) Save() error {
	path, err := config.SessionsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
