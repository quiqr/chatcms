// Package auth answers who someone is. It holds passwords, sessions, and the
// single-use links behind invitations and resets. What it does not hold is
// what anyone is allowed to do, which is internal/policy's.
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

// The parameters a new hash is made with. They are written into the encoded
// form, so raising them later leaves older hashes verifiable.
const (
	argonMemory  = 64 * 1024
	argonTime    = 1
	argonThreads = 4
	argonSaltLen = 16
	argonKeyLen  = 32
)

// MinimumPasswordLength is a floor, and the only rule about shape. Rules
// about which characters must appear reliably produce worse passwords.
const MinimumPasswordLength = 12

var (
	// ErrPasswordTooShort says the password is below the floor.
	ErrPasswordTooShort = fmt.Errorf("a password needs at least %d characters", MinimumPasswordLength)

	// ErrPasswordTooCommon says the password is one of the ones attackers try
	// first.
	ErrPasswordTooCommon = errors.New("this password is one of the most common ones, so it is too easy to guess")

	// ErrWrongPassword says the password does not match.
	ErrWrongPassword = errors.New("that password is not right")

	// ErrUnreadableHash says a stored hash cannot be understood.
	ErrUnreadableHash = errors.New("the stored password cannot be read")
)

// CheckPasswordRules reports whether a password may be used.
func CheckPasswordRules(password string) error {
	if len([]rune(password)) < MinimumPasswordLength {
		return ErrPasswordTooShort
	}
	if isCommonPassword(password) {
		return ErrPasswordTooCommon
	}
	return nil
}

// HashPassword turns a password into the value that is stored. The result
// names its algorithm and parameters so it can be read back by a version of
// this program that uses different ones.
func HashPassword(password string) (string, error) {
	if err := CheckPasswordRules(password); err != nil {
		return "", err
	}

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("no randomness available: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword checks a password against a stored hash, taking the same
// time whether it matches or not.
func VerifyPassword(encoded, password string) error {
	parameters, salt, key, err := decodeHash(encoded)
	if err != nil {
		return err
	}

	candidate := argon2.IDKey([]byte(password), salt,
		parameters.time, parameters.memory, parameters.threads, uint32(len(key)))

	if subtle.ConstantTimeCompare(candidate, key) != 1 {
		return ErrWrongPassword
	}
	return nil
}

type argonParameters struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodeHash(encoded string) (argonParameters, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argonParameters{}, nil, nil, ErrUnreadableHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argonParameters{}, nil, nil, ErrUnreadableHash
	}

	var parameters argonParameters
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&parameters.memory, &parameters.time, &parameters.threads); err != nil {
		return argonParameters{}, nil, nil, ErrUnreadableHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonParameters{}, nil, nil, ErrUnreadableHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return argonParameters{}, nil, nil, ErrUnreadableHash
	}

	return parameters, salt, key, nil
}

// dummyHash is verified against when the address is unknown, so that a wrong
// address does not come back faster than a wrong password.
var dummyHash = func() string {
	hash, err := HashPassword("a password nobody has, long enough to pass the rules")
	if err != nil {
		panic("the dummy password does not pass the rules: " + err.Error())
	}
	return hash
}()

// SpendTimeLikeAFailedCheck does the work a real check would, so that an
// unknown address and a wrong password take the same time.
func SpendTimeLikeAFailedCheck(password string) {
	_ = VerifyPassword(dummyHash, password)
}
