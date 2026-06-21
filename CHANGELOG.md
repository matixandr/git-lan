# Changelog

All notable changes to git-lan are documented here. This project adheres to
[Semantic Versioning](https://semver.org).

## [0.1.0]

First release. Zero-config peer-to-peer git collaboration over the LAN.

### Added

- **Discovery** - automatic peer discovery over mDNS/DNS-SD (`_gitlan._tcp`),
  with a TTL'd registry and minimal plaintext TXT metadata.
- **End-to-end encryption** - ephemeral X25519 handshake with long-term identity
  binding (MITM-resistant), HKDF directional keys, ChaCha20-Poly1305 framing,
  and nonce-based anti-replay (`internal/e2e`).
- **Encrypted transport** - git served via `git daemon --inetd` over the
  decrypted stream, with a client-side loopback bridge. Dynamic port fallback
  and graceful shutdown.
- **Commands** - `list`, `status`, `clone`, `push`, `pull`, `session
  create/join/invite/leave`, `trust add/remove/list`, `config`, `completion`.
- **Sessions** - Argon2id password protection enforced by an in-channel
  challenge/response gate, plus one-time HMAC-signed invite tokens (base58).
- **Trust ring** - pinned peer fingerprints in `trusted_peers.json`, auto-accept
  for known peers, loud abort on fingerprint mismatch.
- **Presence** - online / coding / idle derived from working-tree state and live
  fsnotify activity; broadcast over mDNS.
- **Conflict early-warning** - flags overlapping uncommitted work before a push.
- **Nerd Fonts auto-detection** - ANSI cursor-width probe, cached per terminal
  profile in `terminal_profiles.toml`, with a clean ASCII fallback.
- **Cross-platform** - Linux, macOS (amd64 + arm64), Windows; install scripts and
  a cross-compiling Makefile.

[0.1.0]: https://github.com/matixandr/git-lan/releases/tag/v0.1.0
