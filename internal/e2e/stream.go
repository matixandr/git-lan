package e2e

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptedConn wraps a net.Conn so that every byte read or written is
// transparently encrypted with ChaCha20-Poly1305 using the directional keys
// negotiated during the handshake. It implements net.Conn, so it is a drop-in
// replacement anywhere a plain connection is used - including as the transport
// under git's pkt-line protocol.
//
// Wire frame:
//
//	[4 bytes  big-endian payload length L]
//	[12 bytes nonce]
//	[L bytes  ciphertext || 16-byte Poly1305 tag]
//
// The 4-byte length is authenticated as additional data, so an attacker cannot
// silently truncate or extend a frame without the tag check failing.
type EncryptedConn struct {
	conn net.Conn

	sendAEAD cipher.AEAD
	recvAEAD cipher.AEAD

	writeMu sync.Mutex
	seq     *nonceSequencer

	readMu  sync.Mutex
	guard   replayGuard
	readBuf []byte // leftover plaintext from a partially consumed frame
}

// Client performs the handshake as the initiator and returns an encrypted
// connection over conn.
func Client(conn net.Conn) (*EncryptedConn, error) { return wrap(conn, true) }

// Server performs the handshake as the responder and returns an encrypted
// connection over conn.
func Server(conn net.Conn) (*EncryptedConn, error) { return wrap(conn, false) }

func wrap(conn net.Conn, initiator bool) (*EncryptedConn, error) {
	keys, err := handshake(conn, initiator)
	if err != nil {
		return nil, err
	}
	sendAEAD, err := chacha20poly1305.New(keys.send)
	if err != nil {
		return nil, fmt.Errorf("%w: send cipher: %v", ErrHandshake, err)
	}
	recvAEAD, err := chacha20poly1305.New(keys.recv)
	if err != nil {
		return nil, fmt.Errorf("%w: recv cipher: %v", ErrHandshake, err)
	}
	seq, err := newNonceSequencer()
	if err != nil {
		return nil, err
	}
	return &EncryptedConn{
		conn:     conn,
		sendAEAD: sendAEAD,
		recvAEAD: recvAEAD,
		seq:      seq,
	}, nil
}

// Write encrypts p and sends it as one or more frames. It satisfies io.Writer
// semantics: on success it returns len(p), nil.
func (c *EncryptedConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	total := 0
	for len(p) > 0 {
		chunk := p
		if len(chunk) > MaxFrameSize {
			chunk = chunk[:MaxFrameSize]
		}
		if err := c.writeFrame(chunk); err != nil {
			return total, err
		}
		total += len(chunk)
		p = p[len(chunk):]
	}
	return total, nil
}

func (c *EncryptedConn) writeFrame(plaintext []byte) error {
	nonce, err := c.seq.next()
	if err != nil {
		return err
	}
	ctLen := len(plaintext) + TagSize

	frame := make([]byte, LenPrefixSize+NonceSize, LenPrefixSize+NonceSize+ctLen)
	binary.BigEndian.PutUint32(frame[:LenPrefixSize], uint32(ctLen))
	copy(frame[LenPrefixSize:], nonce)

	// Seal appends ciphertext+tag onto frame; the length prefix is AAD.
	frame = c.sendAEAD.Seal(frame, nonce, plaintext, frame[:LenPrefixSize])
	if _, err := c.conn.Write(frame); err != nil {
		return err
	}
	return nil
}

// Read returns decrypted plaintext. It satisfies io.Reader: it may return fewer
// bytes than a full frame, buffering the remainder for the next call.
func (c *EncryptedConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if len(c.readBuf) == 0 {
		plaintext, err := c.readFrame()
		if err != nil {
			return 0, err
		}
		c.readBuf = plaintext
	}
	n := copy(p, c.readBuf)
	c.readBuf = c.readBuf[n:]
	return n, nil
}

func (c *EncryptedConn) readFrame() ([]byte, error) {
	var hdr [LenPrefixSize + NonceSize]byte
	if _, err := io.ReadFull(c.conn, hdr[:]); err != nil {
		return nil, err
	}
	ctLen := binary.BigEndian.Uint32(hdr[:LenPrefixSize])
	if ctLen < TagSize || int(ctLen)-TagSize > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	nonce := hdr[LenPrefixSize:]
	if err := c.guard.accept(nonce); err != nil {
		return nil, err
	}

	ct := make([]byte, ctLen)
	if _, err := io.ReadFull(c.conn, ct); err != nil {
		return nil, err
	}
	plaintext, err := c.recvAEAD.Open(ct[:0], nonce, ct, hdr[:LenPrefixSize])
	if err != nil {
		return nil, fmt.Errorf("e2e: frame authentication failed: %w", err)
	}
	return plaintext, nil
}

// net.Conn passthrough.

func (c *EncryptedConn) Close() error                  { return c.conn.Close() }
func (c *EncryptedConn) LocalAddr() net.Addr           { return c.conn.LocalAddr() }
func (c *EncryptedConn) RemoteAddr() net.Addr          { return c.conn.RemoteAddr() }
func (c *EncryptedConn) SetDeadline(t time.Time) error { return c.conn.SetDeadline(t) }
func (c *EncryptedConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
func (c *EncryptedConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
