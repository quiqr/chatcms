// Package secrets keeps a secret so that a copy of the database is worth
// nothing. Everything stored is sealed with AES-256-GCM under one master key,
// which is read from a file rather than from configuration so that it is
// never an environment variable and never in the Nix store.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"
)

// KeyLength is how many bytes the master key needs. Anything shorter is
// refused at start-up rather than silently stretched.
const KeyLength = 32

var (
	// ErrNoKey says no master key was configured.
	ErrNoKey = errors.New("no master key is configured, so secrets cannot be kept")

	// ErrKeyTooShort says the key file holds too little.
	ErrKeyTooShort = fmt.Errorf("the master key needs at least %d bytes", KeyLength)

	// ErrCannotOpen says a stored value will not open: it was altered, or it
	// was sealed under a different key.
	ErrCannotOpen = errors.New("this stored secret cannot be read")
)

// Keeper seals and opens secrets.
type Keeper struct {
	gcm cipher.AEAD
}

// NewKeeper reads the master key from a file and prepares to seal with it.
func NewKeeper(keyPath string) (*Keeper, error) {
	if strings.TrimSpace(keyPath) == "" {
		return nil, ErrNoKey
	}

	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("the master key could not be read: %w", err)
	}

	key := normalise(raw)
	if len(key) < KeyLength {
		return nil, ErrKeyTooShort
	}

	block, err := aes.NewCipher(key[:KeyLength])
	if err != nil {
		return nil, fmt.Errorf("the master key is unusable: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("the master key is unusable: %w", err)
	}

	return &Keeper{gcm: gcm}, nil
}

// NewKeeperFromBytes prepares to seal with a key already in hand, which the
// tests use.
func NewKeeperFromBytes(key []byte) (*Keeper, error) {
	if len(key) < KeyLength {
		return nil, ErrKeyTooShort
	}

	block, err := aes.NewCipher(key[:KeyLength])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Keeper{gcm: gcm}, nil
}

// normalise lets a key file hold raw bytes or text with a trailing newline,
// because both are what people produce.
func normalise(raw []byte) []byte {
	return []byte(strings.TrimRight(string(raw), "\r\n"))
}

// Seal encrypts a secret. The nonce is fresh each time and is carried with
// the result, so the same secret sealed twice gives different values.
func (k *Keeper) Seal(secret []byte) ([]byte, error) {
	nonce := make([]byte, k.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("no randomness available: %w", err)
	}
	return k.gcm.Seal(nonce, nonce, secret, nil), nil
}

// Open decrypts a sealed value, failing rather than returning something when
// it has been altered.
func (k *Keeper) Open(sealed []byte) ([]byte, error) {
	size := k.gcm.NonceSize()
	if len(sealed) < size+1 {
		return nil, ErrCannotOpen
	}

	secret, err := k.gcm.Open(nil, sealed[:size], sealed[size:], nil)
	if err != nil {
		return nil, ErrCannotOpen
	}
	return secret, nil
}

// SealString and OpenString are the common case.
func (k *Keeper) SealString(secret string) ([]byte, error) {
	return k.Seal([]byte(secret))
}

func (k *Keeper) OpenString(sealed []byte) (string, error) {
	secret, err := k.Open(sealed)
	return string(secret), err
}
