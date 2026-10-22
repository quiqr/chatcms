package secrets

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// KeyPair is a deploy key: the half that goes to the git host, and the half
// that never leaves here unsealed.
type KeyPair struct {
	// Public is the single line an implementer pastes into the git host.
	Public string
	// Private is the OpenSSH private key. It is sealed before storage and is
	// never returned by any endpoint.
	Private string
}

// GenerateKeyPair makes a new Ed25519 deploy key, without running any other
// program, so no private key is ever written to a temporary file on the way.
func GenerateKeyPair(comment string) (KeyPair, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("a key could not be made: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(private, comment)
	if err != nil {
		return KeyPair{}, fmt.Errorf("the key could not be written: %w", err)
	}

	sshPublic, err := ssh.NewPublicKey(public)
	if err != nil {
		return KeyPair{}, fmt.Errorf("the key could not be written: %w", err)
	}

	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublic)))
	if comment != "" {
		line += " " + comment
	}

	return KeyPair{
		Public:  line,
		Private: string(pem.EncodeToMemory(block)),
	}, nil
}

// PublicKeyFrom reads the public half out of a private key, which the
// connection test uses to show the key again.
func PublicKeyFrom(private, comment string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(private))
	if err != nil {
		return "", fmt.Errorf("the stored key could not be read: %w", err)
	}

	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
	if comment != "" {
		line += " " + comment
	}
	return line, nil
}
