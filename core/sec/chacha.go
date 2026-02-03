package sec

//---grok generated code
import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptedConn wraps a net.Conn with ChaCha20-Poly1305 encryption
type EncryptedConn struct {
	net.Conn
	aead         cipher.AEAD
	readMutex    sync.Mutex
	writeMutex   sync.Mutex
	nonceCounter uint64 // simple monotonic nonce (64-bit)
}

// NewEncryptedConn creates encrypted connection using a 32-byte key
func NewEncryptedConn(base net.Conn, key [32]byte) (*EncryptedConn, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, err
	}

	return &EncryptedConn{
		Conn:         base,
		aead:         aead,
		nonceCounter: 0,
	}, nil
}

// Read decrypts and authenticates data
func (c *EncryptedConn) Read(p []byte) (n int, err error) {
	c.readMutex.Lock()
	defer c.readMutex.Unlock()

	// Read length + nonce + tag + ciphertext
	header := make([]byte, 4+chacha20poly1305.NonceSize+chacha20poly1305.Overhead)
	_, err = io.ReadFull(c.Conn, header)
	if err != nil {
		return 0, err
	}

	length := binary.BigEndian.Uint32(header[:4])
	if length > 65535 {
		return 0, errors.New("frame too large")
	}

	nonce := header[4 : 4+chacha20poly1305.NonceSize]
	ciphertext := make([]byte, length+chacha20poly1305.Overhead)
	copy(ciphertext, header[4+chacha20poly1305.NonceSize:])

	_, err = io.ReadFull(c.Conn, ciphertext[chacha20poly1305.Overhead:])
	if err != nil {
		return 0, err
	}

	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err // authentication failed → MITM / corruption
	}

	n = copy(p, plaintext)
	if n < len(plaintext) {
		// TODO: you can implement buffering if needed
		return n, io.ErrShortBuffer
	}

	return n, nil
}

// Write encrypts and sends data (splits into chunks if needed)
func (c *EncryptedConn) Write(p []byte) (int, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	const maxChunk = 16384 // reasonable size (like TLS record)

	totalWritten := 0
	for len(p) > 0 {
		chunk := p
		if len(chunk) > maxChunk {
			chunk = chunk[:maxChunk]
		}

		nonce := make([]byte, chacha20poly1305.NonceSize)
		binary.BigEndian.PutUint64(nonce[:8], c.nonceCounter)
		c.nonceCounter++

		ciphertext := c.aead.Seal(nil, nonce, chunk, nil)

		// length (4) + nonce (12) + tag+ciphertext (16+len)
		header := make([]byte, 4)
		binary.BigEndian.PutUint32(header, uint32(len(chunk)))

		frame := append(header, nonce...)
		frame = append(frame, ciphertext...)

		_, err := c.Conn.Write(frame)
		if err != nil {
			return totalWritten, err
		}

		totalWritten += len(chunk)
		p = p[len(chunk):]
	}

	return totalWritten, nil
}

func (c *EncryptedConn) Close() error {
	return c.Conn.Close()
}

/////////////

// 32bit key
type ChaCha struct {
	Key [32]byte
}

func (ch ChaCha) WrapConn(c net.Conn) net.Conn {
	ec, err := NewEncryptedConn(c, ch.Key)
	if err != nil {
		return nil
	}
	return ec
}
