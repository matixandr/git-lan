# Security Model

git-lan is built so that **no application data ever crosses the network in the
clear**. This document describes what it protects, how, and the limits of that
protection.

## Threat model

git-lan targets a shared local network (an office, a hackathon hall, a
classroom) where you trust *some* people but not the network itself. The
adversary we defend against is a passive or active attacker on the same LAN:

- **Eavesdropping** - reading your git data off the wire.
- **Tampering** - modifying bytes in flight.
- **Impersonation / MITM** - pretending to be a peer you meant to talk to.
- **Replay** - re-sending captured frames.

git-lan does **not** try to defend against a compromised endpoint, a malicious
peer you explicitly trusted, or an attacker who can run code on your machine.

## How connections are protected

Every peer-to-peer connection is wrapped by `internal/e2e` before a single byte
of git protocol flows.

1. **Ephemeral key exchange.** Both sides generate fresh X25519 keypairs per
   connection and exchange public keys. This gives *forward secrecy*: recording
   today's traffic and stealing a key tomorrow does not decrypt it.

2. **Identity binding.** Each side also presents its long-term X25519 identity
   key. The handshake mixes two static–ephemeral Diffie–Hellman results into the
   key schedule, so the session keys depend on both identities. An attacker who
   substitutes their own identity cannot derive the same keys.

3. **Directional keys.** HKDF-SHA256 expands the shared secret into two
   independent ChaCha20-Poly1305 keys, one per direction. Compromise of one
   direction's key does not affect the other.

4. **Authenticated framing.** Every frame is
   `[len][nonce][ciphertext+tag]`, sealed with ChaCha20-Poly1305. The length
   prefix is authenticated, so truncation and extension are detected.

5. **Anti-replay.** Nonces are a 32-bit random prefix plus a 64-bit counter.
   Receivers reject any nonce that does not strictly advance, defeating replay
   and reordering. A replayed frame is dropped, not crashed on.

## Identity and trust

- On first run, git-lan generates a long-term X25519 identity in
  `identity.key` (0600 on Unix; ACL-restricted to the current user on Windows).
- A peer's **fingerprint** is `SHA256:` + base64url(SHA-256(public key)).
- The first time you connect to an unknown peer you see its fingerprint and
  choose accept-once, reject, or trust-always. Trusted peers are pinned in
  `trusted_peers.json`.
- On every later connection, the peer's presented identity is checked against
  its pin. **A mismatch aborts the connection with a man-in-the-middle
  warning** - it is never silently accepted.

Verify fingerprints out of band (say them aloud, compare on screen) the first
time, exactly as you would SSH host keys.

## Session access control

- Sessions may be password-protected. The password is run through **Argon2id**
  to derive a seed; the seed never leaves the host and is never written to disk.
- Joining a locked session runs a challenge/response **inside** the already
  encrypted channel. A wrong password fails the HMAC check and the connection is
  closed before any git data is served.
- **Invite tokens** are `base58(random id ‖ expiry ‖ HMAC-SHA256)`. They are
  signed by the session secret, expire, and are **one-time**: the host burns the
  token id on first use (constant-time comparison) so it cannot be replayed.

## What is *not* encrypted

mDNS service discovery is a broadcast protocol and its TXT records are visible
to anyone on the LAN by design. git-lan keeps them to harmless metadata only:

    protocol version, repo name, branch, short HEAD, modified-file count,
    session name, lock flag, presence

No keys, passwords, tokens, or file contents ever appear in a TXT record.

## Reporting

This is a hobby project, not audited software. If you find a vulnerability,
please open an issue describing it (omit working exploits for anything severe
until it can be fixed).
