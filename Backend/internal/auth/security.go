package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordConfig struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

var (
	// NIST/OWASP recommended parameters for Argon2id (Sensitive Environment)
	// Memory: 64MB, Iterations: 1, Parallelism: 4
	argon2Config = &PasswordConfig{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
	}
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// HashPassword generates an Argon2id hash of the password using a random salt and a static pepper.
// The pepper is a secret application-level key that must be kept secure.
func HashPassword(password, pepper string) (string, error) {
	// Generate random salt (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Combine password and pepper
	// P + Secret
	pepperedPassword := []byte(password + pepper)

	// Keep sensitive data in memory for as short as possible (though Go GC makes this hard to guarantee)
	hash := argon2.IDKey(pepperedPassword, salt, argon2Config.time, argon2Config.memory, argon2Config.threads, argon2Config.keyLen)

	// Base64 encode the salt and hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Config.memory, argon2Config.time, argon2Config.threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword compares a password against an Argon2id hash using a static pepper.
func VerifyPassword(password, encodedHash, pepper string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	var memory uint32
	var time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLen := uint32(len(decodedHash))

	pepperedPassword := []byte(password + pepper)
	comparisonHash := argon2.IDKey(pepperedPassword, salt, time, memory, threads, keyLen)

	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return true, nil
	}
	return false, nil
}
