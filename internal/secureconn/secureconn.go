package secureconn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"net"

	"golang.org/x/crypto/curve25519"
)

type Conn struct {
	net.Conn
	reader cipher.Stream
	writer cipher.Stream
}

// Wrap turns a net.Conn into an encrypted connection.
// isServer determines who starts the key exchange handshake.
func Wrap(conn net.Conn, isServer bool) (*Conn, error) {
	// Generate ephemeral key pair
	var privKey [32]byte
	_, err := rand.Read(privKey[:])
	if err != nil {
		return nil, err
	}
	pubKey, err := curve25519.X25519(privKey[:], curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	// Exchange pubkeys
	var peerPubKey [32]byte
	if isServer {
		// Server reads first
		if _, err := io.ReadFull(conn, peerPubKey[:]); err != nil {
			return nil, err
		}
		if _, err := conn.Write(pubKey); err != nil {
			return nil, err
		}
	} else {
		// Client writes first
		if _, err := conn.Write(pubKey); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(conn, peerPubKey[:]); err != nil {
			return nil, err
		}
	}

	// Derive shared secret
	sharedSecret, err := curve25519.X25519(privKey[:], peerPubKey[:])
	if err != nil {
		return nil, err
	}

	// Derive key and nonce
	hash := sha256.Sum256(sharedSecret)
	key := hash[:16]   // AES-128 for speed (use [:32] for AES-256 if preferred)
	nonce := hash[16:] // 16 bytes

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create XOR streams (CTR mode)
	reader := cipher.NewCTR(block, nonce)
	writer := cipher.NewCTR(block, nonce)

	return &Conn{
		Conn:   conn,
		reader: reader,
		writer: writer,
	}, nil
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.reader.XORKeyStream(p[:n], p[:n])
	}
	return n, err
}

func (c *Conn) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	c.writer.XORKeyStream(buf, p)
	return c.Conn.Write(buf)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}
