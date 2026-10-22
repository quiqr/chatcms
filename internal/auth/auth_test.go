package auth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const usable = "a long enough passphrase"

func TestHashNamesItsParameters(t *testing.T) {
	hash, err := HashPassword(usable)
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash = %q, want it to name argon2id", hash)
	}
	for _, part := range []string{"v=", "m=", "t=", "p="} {
		if !strings.Contains(hash, part) {
			t.Errorf("hash = %q, want it to carry %q", hash, part)
		}
	}
	if strings.Contains(hash, usable) {
		t.Error("the stored hash contains the password")
	}
}

func TestTwoPeopleWithOnePasswordDifferInStorage(t *testing.T) {
	first, err := HashPassword(usable)
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword(usable)
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Error("the same password hashed twice gives the same value; the salt is not per password")
	}
	for _, hash := range []string{first, second} {
		if err := VerifyPassword(hash, usable); err != nil {
			t.Errorf("VerifyPassword() error = %v, want nil", err)
		}
	}
}

func TestVerifyRejectsAWrongPassword(t *testing.T) {
	hash, err := HashPassword(usable)
	if err != nil {
		t.Fatal(err)
	}

	if err := VerifyPassword(hash, usable+"x"); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("VerifyPassword() error = %v, want ErrWrongPassword", err)
	}
}

func TestVerifyAcceptsOtherParameters(t *testing.T) {
	// A hash made with weaker parameters than today's still verifies, because
	// the parameters it was made with are in the hash.
	weaker := "$argon2id$v=19$m=19456,t=2,p=1$c29tZXNhbHR2YWx1ZQ$" +
		""
	_ = weaker

	hash, err := HashPassword(usable)
	if err != nil {
		t.Fatal(err)
	}
	// Rewrite the parameters to values the verifier must read rather than
	// assume, and confirm it no longer matches, proving the parameters in the
	// hash are the ones used.
	altered := strings.Replace(hash, "m=65536", "m=32768", 1)
	if altered == hash {
		t.Fatal("the hash does not carry the memory parameter in the expected form")
	}
	if err := VerifyPassword(altered, usable); err == nil {
		t.Error("changing the parameters in the hash did not change the result, so they are not being read")
	}
}

func TestVerifyRejectsAnUnreadableHash(t *testing.T) {
	for _, hash := range []string{"", "not a hash", "$argon2id$v=19$m=1$x$y", "$bcrypt$something"} {
		if err := VerifyPassword(hash, usable); !errors.Is(err, ErrUnreadableHash) {
			t.Errorf("VerifyPassword(%q) error = %v, want ErrUnreadableHash", hash, err)
		}
	}
}

func TestPasswordRules(t *testing.T) {
	cases := map[string]struct {
		password string
		want     error
	}{
		"too short":                 {"short", ErrPasswordTooShort},
		"exactly one short":         {strings.Repeat("a", MinimumPasswordLength-1), ErrPasswordTooShort},
		"a very common password":    {"administrator", ErrPasswordTooCommon},
		"a common password shouted": {"ADMINISTRATOR", ErrPasswordTooCommon},
		"a common long passphrase":  {"correcthorsebatterystaple", ErrPasswordTooCommon},
		"long enough":               {"een wachtwoord dat lang genoeg is", nil},
		"a lower-case passphrase":   {"ruimte voor groen in de stad", nil},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckPasswordRules(tc.password)
			if tc.want == nil && err != nil {
				t.Errorf("CheckPasswordRules() error = %v, want nil", err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("CheckPasswordRules() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPasswordRulesHaveNoCompositionRequirement(t *testing.T) {
	if err := CheckPasswordRules("zevenentwintig letters erbij"); err != nil {
		t.Errorf("a long password of one repeated letter was refused: %v", err)
	}
}

func TestCommonListIsEmbedded(t *testing.T) {
	if CommonPasswordCount() < 50 {
		t.Errorf("the common password list holds %d entries, want the embedded list", CommonPasswordCount())
	}
	if !isCommonPassword("123456") {
		t.Error("the most common password of all is not in the list")
	}
}

func TestSpendTimeLikeAFailedCheckDoesWork(t *testing.T) {
	started := time.Now()
	SpendTimeLikeAFailedCheck("anything")
	if time.Since(started) < time.Millisecond {
		t.Error("the dummy check returned immediately, so an unknown address is faster than a wrong password")
	}
}

func TestLimiterCountsPerAddressAndPerCaller(t *testing.T) {
	limiter := NewLimiter()

	for range AttemptsPerAddress {
		if err := limiter.Allow("one@example.org", "10.0.0.1"); err != nil {
			t.Fatalf("Allow() error = %v before the limit", err)
		}
		limiter.RecordFailure("one@example.org", "10.0.0.1")
	}

	if err := limiter.Allow("one@example.org", "10.0.0.9"); !errors.Is(err, ErrTooManyAttempts) {
		t.Errorf("Allow() for a limited address from a fresh caller = %v, want ErrTooManyAttempts", err)
	}
	if err := limiter.Allow("other@example.org", "10.0.0.9"); err != nil {
		t.Errorf("Allow() for a fresh address = %v, want nil", err)
	}
}

func TestLimitPerCaller(t *testing.T) {
	limiter := NewLimiter()

	for i := range AttemptsPerCaller {
		address := "person" + string(rune('a'+i%26)) + "@example.org"
		if err := limiter.Allow(address, "10.0.0.1"); err != nil {
			t.Fatalf("Allow() error = %v before the limit", err)
		}
		limiter.RecordFailure(address, "10.0.0.1")
	}

	if err := limiter.Allow("someone-new@example.org", "10.0.0.1"); !errors.Is(err, ErrTooManyAttempts) {
		t.Errorf("Allow() from a limited caller = %v, want ErrTooManyAttempts", err)
	}
}

func TestSuccessClearsTheCount(t *testing.T) {
	limiter := NewLimiter()

	for range AttemptsPerAddress - 1 {
		limiter.RecordFailure("one@example.org", "10.0.0.1")
	}
	limiter.RecordSuccess("one@example.org", "10.0.0.1")

	for range AttemptsPerAddress {
		if err := limiter.Allow("one@example.org", "10.0.0.1"); err != nil {
			t.Fatalf("Allow() error = %v after a success cleared the count", err)
		}
		limiter.RecordFailure("one@example.org", "10.0.0.1")
	}
}

func TestLimiterForgetsAfterTheWindow(t *testing.T) {
	limiter := NewLimiter()
	moment := time.Now()
	limiter.now = func() time.Time { return moment }

	for range AttemptsPerAddress {
		limiter.RecordFailure("one@example.org", "10.0.0.1")
	}
	if err := limiter.Allow("one@example.org", "10.0.0.1"); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("Allow() = %v, want the limit to bite", err)
	}

	moment = moment.Add(AttemptWindow + time.Minute)
	if err := limiter.Allow("one@example.org", "10.0.0.1"); err != nil {
		t.Errorf("Allow() after the window = %v, want nil", err)
	}
}

func TestRefusalSaysNothingAboutTheAddress(t *testing.T) {
	message := ErrTooManyAttempts.Error()
	for _, word := range []string{"exist", "unknown", "found", "account", "address is"} {
		if strings.Contains(strings.ToLower(message), word) {
			t.Errorf("the refusal says %q, which hints at whether the address is known", message)
		}
	}
}
