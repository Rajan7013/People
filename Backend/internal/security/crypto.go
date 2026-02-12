package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// Encrypt encrypts plain text string into base64 encoded string using AES-GCM
func Encrypt(plaintext string, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Key must be 32 bytes (256 bits)
	// We assume the key passed is the raw string from env, so we might need to pad/truncate or hash it to ensure it's 32 bytes
	// For better security, usually KDF is used, but for simplicity here we'll SHA256 the key to get 32 bytes
	keyBytes := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded string into plain text string using AES-GCM
func Decrypt(encryptedText string, key string) (string, error) {
	if encryptedText == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	keyBytes := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBytes[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HashBlindIndex creates a deterministic hash for searching equivalent values
// without exposing the plaintext or using the same key as encryption
func HashBlindIndex(input string, key string) string {
	if input == "" {
		return ""
	}
	// Use a different derivation for the blind index key
	// In production, this should ideally be a completely separate env var
	blindKey := sha256.Sum256([]byte("blind_index_" + key))

	h := hmac.New(sha256.New, blindKey[:])
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}
