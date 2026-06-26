package e2e

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

// ClientAuth performs an authenticated handshake as the initiator and returns
// the encrypted connection together with the peer's identity public key.
func ClientAuth(conn net.Conn, identity *ecdh.PrivateKey) (*EncryptedConn, []byte, error) {
	return wrapAuth(conn, true, identity)
}

// ServerAuth performs an authenticated handshake as the responder.
func ServerAuth(conn net.Conn, identity *ecdh.PrivateKey) (*EncryptedConn, []byte, error) {
	return wrapAuth(conn, false, identity)
}

func wrapAuth(conn net.Conn, initiator bool, identity *ecdh.PrivateKey) (*EncryptedConn, []byte, error) {
	keys, peerID, err := handshakeAuth(conn, initiator, identity)
	if err != nil {
		return nil, nil, err
	}
	ec, err := fromKeys(conn, keys)
	if err != nil {
		return nil, nil, err
	}
	return ec, peerID, nil
}

// authMsg is the wire layout of each side's handshake message: its ephemeral
// public key followed by its long-term identity public key.
const authMsgLen = 32 + 32

// handshakeAuth performs a mutually-authenticated handshake. In addition to the
// ephemeral ECDH that gives forward secrecy, it mixes two static-ephemeral
// Diffie-Hellman results into the key schedule:
//
//	my_ephemeral × peer_identity   and   my_identity × peer_ephemeral
//
// By DH symmetry both peers compute the same pair, but only a party that holds
// the matching identity private key can. The session keys therefore depend on
// both identities - an active attacker who swaps in their own identity produces
// different keys (and a different, pin-detectable fingerprint).
//
// It returns the negotiated keys and the peer's identity public key, which the
// caller verifies against the trust ring before any application data flows.
func handshakeAuth(rw io.ReadWriter, initiator bool, identity *ecdh.PrivateKey) (*sessionKeys, []byte, error) {
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: ephemeral key: %v", ErrHandshake, err)
	}

	myMsg := make([]byte, 0, authMsgLen)
	myMsg = append(myMsg, eph.PublicKey().Bytes()...)
	myMsg = append(myMsg, identity.PublicKey().Bytes()...)

	peerMsg, err := exchangeFixed(rw, myMsg)
	if err != nil {
		return nil, nil, err
	}
	peerEphPub := peerMsg[:32]
	peerIDPub := peerMsg[32:]

	peerEph, err := ecdh.X25519().NewPublicKey(peerEphPub)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: peer ephemeral key: %v", ErrHandshake, err)
	}
	peerID, err := ecdh.X25519().NewPublicKey(peerIDPub)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: peer identity key: %v", ErrHandshake, err)
	}

	ssEE, err := eph.ECDH(peerEph) // forward secrecy
	if err != nil {
		return nil, nil, fmt.Errorf("%w: ee dh: %v", ErrHandshake, err)
	}
	ssES, err := eph.ECDH(peerID) // my ephemeral × peer identity
	if err != nil {
		return nil, nil, fmt.Errorf("%w: es dh: %v", ErrHandshake, err)
	}
	ssSE, err := identity.ECDH(peerEph) // my identity × peer ephemeral
	if err != nil {
		return nil, nil, fmt.Errorf("%w: se dh: %v", ErrHandshake, err)
	}

	// A's ssES equals B's ssSE and vice versa, so canonicalize by sorting the
	// two cross terms to reach a master secret both sides agree on.
	lo, hi := ssES, ssSE
	if bytes.Compare(lo, hi) > 0 {
		lo, hi = hi, lo
	}
	master := make([]byte, 0, len(ssEE)+len(lo)+len(hi))
	master = append(master, ssEE...)
	master = append(master, lo...)
	master = append(master, hi...)

	salt := transcriptSalt(eph.PublicKey().Bytes(), peerEphPub)
	keyAtoB := deriveKey(master, salt, infoAtoB)
	keyBtoA := deriveKey(master, salt, infoBtoA)

	keys := &sessionKeys{send: keyAtoB, recv: keyBtoA}
	if !initiator {
		keys = &sessionKeys{send: keyBtoA, recv: keyAtoB}
	}
	// Return a copy of the peer identity so callers cannot mutate our slice.
	return keys, append([]byte(nil), peerIDPub...), nil
}

// exchangeFixed writes msg and reads a peer message of the same fixed length.
func exchangeFixed(rw io.ReadWriter, msg []byte) ([]byte, error) {
	if _, err := rw.Write(msg); err != nil {
		return nil, fmt.Errorf("%w: send: %v", ErrHandshake, err)
	}
	peer := make([]byte, len(msg))
	if _, err := io.ReadFull(rw, peer); err != nil {
		return nil, fmt.Errorf("%w: read: %v", ErrHandshake, err)
	}
	return peer, nil
}
