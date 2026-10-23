package secrets

import (
	"bytes"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const aSecret = "a private key that nobody else should read"

func keeper(t *testing.T) *Keeper {
	t.Helper()

	key := make([]byte, KeyLength)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	made, err := NewKeeperFromBytes(key)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func TestSealedValueIsNotTheSecret(t *testing.T) {
	sealed, err := keeper(t).SealString(aSecret)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}

	if bytes.Contains(sealed, []byte(aSecret)) {
		t.Error("the sealed value contains the secret")
	}
	if bytes.Contains(sealed, []byte("private key")) {
		t.Error("the sealed value contains part of the secret")
	}
}

func TestOpeningReturnsWhatWasSealed(t *testing.T) {
	k := keeper(t)

	sealed, err := k.SealString(aSecret)
	if err != nil {
		t.Fatal(err)
	}

	opened, err := k.OpenString(sealed)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	if opened != aSecret {
		t.Errorf("opened %q, want the secret back", opened)
	}
}

func TestSealingTwiceDiffers(t *testing.T) {
	k := keeper(t)

	first, _ := k.SealString(aSecret)
	second, _ := k.SealString(aSecret)

	if bytes.Equal(first, second) {
		t.Error("the same secret sealed twice gave the same value, so the nonce is not fresh")
	}

	for _, sealed := range [][]byte{first, second} {
		opened, err := k.OpenString(sealed)
		if err != nil || opened != aSecret {
			t.Errorf("Open() = %q, %v", opened, err)
		}
	}
}

func TestTamperedValueWillNotOpen(t *testing.T) {
	k := keeper(t)
	sealed, _ := k.SealString(aSecret)

	cases := map[string]func([]byte) []byte{
		"a changed nonce": func(value []byte) []byte {
			altered := bytes.Clone(value)
			altered[0] ^= 0xff
			return altered
		},
		"a changed ciphertext": func(value []byte) []byte {
			altered := bytes.Clone(value)
			altered[len(altered)-1] ^= 0xff
			return altered
		},
		"truncated": func(value []byte) []byte {
			return value[:len(value)-1]
		},
		"empty": func([]byte) []byte { return nil },
	}

	for name, alter := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := k.Open(alter(sealed)); !errors.Is(err, ErrCannotOpen) {
				t.Errorf("Open() error = %v, want ErrCannotOpen", err)
			}
		})
	}
}

func TestAnotherKeyCannotOpenIt(t *testing.T) {
	sealed, _ := keeper(t).SealString(aSecret)

	if _, err := keeper(t).Open(sealed); !errors.Is(err, ErrCannotOpen) {
		t.Errorf("another key opened the value: %v", err)
	}
}

func TestMasterKeyIsReadFromItsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-key")
	if err := os.WriteFile(path, []byte(strings.Repeat("k", KeyLength)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	k, err := NewKeeper(path)
	if err != nil {
		t.Fatalf("NewKeeper() error = %v, want nil", err)
	}

	sealed, err := k.SealString(aSecret)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := k.OpenString(sealed)
	if err != nil || opened != aSecret {
		t.Errorf("a key read from a file does not work: %q, %v", opened, err)
	}
}

func TestMissingKeyRefusesToStart(t *testing.T) {
	if _, err := NewKeeper(""); !errors.Is(err, ErrNoKey) {
		t.Errorf("NewKeeper(\"\") error = %v, want ErrNoKey", err)
	}
	if _, err := NewKeeper(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("NewKeeper() with no such file error = nil, want a failure")
	}
}

func TestShortKeyRefusesToStart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "short-key")
	if err := os.WriteFile(path, []byte("too short"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewKeeper(path)
	if !errors.Is(err, ErrKeyTooShort) {
		t.Fatalf("NewKeeper() error = %v, want ErrKeyTooShort", err)
	}
	if !strings.Contains(err.Error(), "32") {
		t.Errorf("error = %q, want it to say how many bytes are needed", err)
	}
}

func TestNewKeeperFromBytesRefusesAShortKey(t *testing.T) {
	if _, err := NewKeeperFromBytes([]byte("short")); !errors.Is(err, ErrKeyTooShort) {
		t.Errorf("error = %v, want ErrKeyTooShort", err)
	}
}

func TestRawKeyEndingInANewlineStillWorks(t *testing.T) {
	// `head -c 32 /dev/urandom` writes raw bytes, and one time in a hundred
	// the last of them is a newline. Trimming it would leave the key a byte
	// short and the server would refuse to start, for no reason anybody
	// could see.
	raw := append([]byte(strings.Repeat("k", KeyLength-1)), '\n')

	path := filepath.Join(t.TempDir(), "raw-key")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewKeeper(path); err != nil {
		t.Fatalf("NewKeeper() error = %v, want nil", err)
	}
}

func TestATextKeyStillLosesItsNewline(t *testing.T) {
	longer := strings.Repeat("k", KeyLength+4) + "\n"

	path := filepath.Join(t.TempDir(), "text-key")
	if err := os.WriteFile(path, []byte(longer), 0o600); err != nil {
		t.Fatal(err)
	}

	k, err := NewKeeper(path)
	if err != nil {
		t.Fatal(err)
	}

	same, err := NewKeeperFromBytes([]byte(strings.Repeat("k", KeyLength+4)))
	if err != nil {
		t.Fatal(err)
	}

	sealed, err := k.SealString(aSecret)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := same.OpenString(sealed)
	if err != nil || opened != aSecret {
		t.Errorf("the newline was not dropped: %q, %v", opened, err)
	}
}
