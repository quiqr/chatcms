package secrets

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGeneratedKeyIsAUsableSSHKey(t *testing.T) {
	pair, err := GenerateKeyPair("chatcms project korenaar")
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v, want nil", err)
	}

	if !strings.HasPrefix(pair.Public, "ssh-ed25519 ") {
		t.Errorf("public half = %q, want an ed25519 line", pair.Public)
	}
	if !strings.HasSuffix(pair.Public, "chatcms project korenaar") {
		t.Errorf("public half = %q, want it to carry the comment", pair.Public)
	}
	if strings.Contains(pair.Public, "\n") {
		t.Error("the public half is more than one line, so it cannot be pasted")
	}

	if !strings.HasPrefix(pair.Private, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		t.Errorf("private half does not look like an OpenSSH key:\n%s", pair.Private)
	}

	// The halves have to belong together, which is what git will rely on.
	signer, err := ssh.ParsePrivateKey([]byte(pair.Private))
	if err != nil {
		t.Fatalf("the private half will not parse: %v", err)
	}

	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pair.Public))
	if err != nil {
		t.Fatalf("the public half will not parse: %v", err)
	}
	if string(parsed.Marshal()) != string(signer.PublicKey().Marshal()) {
		t.Error("the two halves do not belong together")
	}
}

func TestEveryKeyIsDifferent(t *testing.T) {
	seen := map[string]bool{}

	for range 5 {
		pair, err := GenerateKeyPair("chatcms")
		if err != nil {
			t.Fatal(err)
		}
		if seen[pair.Public] {
			t.Fatal("a key was generated twice")
		}
		seen[pair.Public] = true
	}
}

func TestPublicKeyFromThePrivateHalf(t *testing.T) {
	pair, err := GenerateKeyPair("chatcms project korenaar")
	if err != nil {
		t.Fatal(err)
	}

	recovered, err := PublicKeyFrom(pair.Private, "chatcms project korenaar")
	if err != nil {
		t.Fatalf("PublicKeyFrom() error = %v, want nil", err)
	}
	if recovered != pair.Public {
		t.Errorf("recovered %q, want %q", recovered, pair.Public)
	}
}

func TestPublicKeyFromSomethingThatIsNotAKey(t *testing.T) {
	if _, err := PublicKeyFrom("not a key at all", ""); err == nil {
		t.Error("PublicKeyFrom() error = nil, want a failure")
	}
}

func TestAKeyPairSealsAndOpens(t *testing.T) {
	pair, err := GenerateKeyPair("chatcms")
	if err != nil {
		t.Fatal(err)
	}

	k := keeper(t)
	sealed, err := k.SealString(pair.Private)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(sealed), "OPENSSH") {
		t.Error("the sealed key still looks like a key")
	}

	opened, err := k.OpenString(sealed)
	if err != nil || opened != pair.Private {
		t.Errorf("the key did not come back: %v", err)
	}
}
