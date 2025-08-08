package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
	"io"
)

// deriveKey generates a 32-byte AES key from a password and salt using PBKDF2.
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha3.New256)
}

// aesEncrypt encrypts data using AES-256-GCM.
func aesEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// aesDecrypt decrypts data using AES-256-GCM.
func aesDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, encryptedMessage, nil)
}

// --- AES Coder & Crypto Utilities ---

// AesCoder implements the arpc.Coder interface for encryption/decryption,
// following the buffer manipulation pattern.
type AesCoder struct {
	key []byte
	log logger.ILogger
}

func NewAesCoder(key []byte, log logger.ILogger) *AesCoder {
	return &AesCoder{
		key: key,
		log: log,
	}
}

// Encode is called before a message is sent. It encrypts the message's body.
func (c *AesCoder) Encode(client *arpc.Client, msg *arpc.Message) *arpc.Message {
	plaintext := msg.Data()
	if len(plaintext) == 0 {
		return msg // Nothing to encrypt.
	}

	ciphertext, err := aesEncrypt(c.key, plaintext)
	if err != nil {
		c.log.Error("AesCoder: Failed to encrypt message", "error", err)
		// Following the reference, we don't alter the message on error.
		// In a real-world scenario, you might want to return an error message instead.
		return msg
	}

	// Rebuild the buffer: keep the header + method, append the new encrypted body.
	methodLen := msg.MethodLen()
	msg.Buffer = append(msg.Buffer[:arpc.HeadLen+methodLen], ciphertext...)
	msg.SetBodyLen(methodLen + len(ciphertext)) // Update the total body length.

	return msg
}

// Decode is called after a message is received. It decrypts the message's body.
func (c *AesCoder) Decode(client *arpc.Client, msg *arpc.Message) *arpc.Message {

	ciphertext := msg.Data()
	if len(ciphertext) == 0 {
		return msg // Nothing to decrypt.
	}

	plaintext, err := aesDecrypt(c.key, ciphertext)
	if err != nil {
		c.log.Error("AesCoder: Failed to decrypt message", "error", err)
		// On failure, we return the message as-is. The next handler will likely
		// fail to unmarshal the corrupted (still encrypted) data.
		return msg
	}

	// Rebuild the buffer: keep the header + method, append the new decrypted body.
	methodLen := msg.MethodLen()
	msg.Buffer = append(msg.Buffer[:arpc.HeadLen+methodLen], plaintext...)
	msg.SetBodyLen(methodLen + len(plaintext)) // Update the total body length.

	return msg
}
