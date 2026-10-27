//go:build integration

package auth_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/png"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/testsupport"
)

const password = "a long enough passphrase"

type harness struct {
	service *auth.Service
	logs    *bytes.Buffer
	db      *store.DB
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}

	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))

	return &harness{service: auth.NewService(db, logger, "https://chatcms.example"), logs: logs, db: db}
}

func (h *harness) person(t *testing.T, email string) auth.Person {
	t.Helper()

	person, err := h.service.People().Create(t.Context(), email, "A Person", password)
	if err != nil {
		t.Fatalf("creating a person failed: %v", err)
	}
	return person
}

func TestSignInIssuesASession(t *testing.T) {
	h := newHarness(t)
	h.person(t, "joris@example.org")

	actor, token, expires, err := h.service.SignIn(t.Context(), "joris@example.org", password, "10.0.0.1", false)
	if err != nil {
		t.Fatalf("SignIn() error = %v, want nil", err)
	}
	if actor.Email != "joris@example.org" || actor.Anonymous {
		t.Errorf("actor = %+v", actor)
	}
	if token == "" {
		t.Fatal("no session value was returned")
	}
	if !expires.After(time.Now()) {
		t.Errorf("expires = %v, want a time in the future", expires)
	}

	found, err := h.service.Lookup(t.Context(), token)
	if err != nil || found.Anonymous {
		t.Fatalf("Lookup() = %+v, %v, want the person", found, err)
	}
}

func TestAddressesAreCaseInsensitive(t *testing.T) {
	h := newHarness(t)
	h.person(t, "Joris@Example.org")

	if _, _, _, err := h.service.SignIn(t.Context(), "joris@example.ORG", password, "10.0.0.1", false); err != nil {
		t.Errorf("SignIn() with different capitalisation error = %v, want nil", err)
	}

	if _, err := h.service.People().Create(t.Context(), "JORIS@example.org", "Another", password); err == nil {
		t.Error("a second person was created with the same address in different capitalisation")
	}
}

func TestSignInSaysTheSameThingEitherWay(t *testing.T) {
	h := newHarness(t)
	h.person(t, "joris@example.org")

	_, _, _, wrongPassword := h.service.SignIn(t.Context(), "joris@example.org", "not the password at all", "10.0.0.1", false)
	_, _, _, unknownAddress := h.service.SignIn(t.Context(), "nobody@example.org", password, "10.0.0.2", false)

	if !errors.Is(wrongPassword, httpapi.ErrSignInRefused) {
		t.Errorf("a wrong password gave %v", wrongPassword)
	}
	if !errors.Is(unknownAddress, httpapi.ErrSignInRefused) {
		t.Errorf("an unknown address gave %v", unknownAddress)
	}
	if wrongPassword.Error() != unknownAddress.Error() {
		t.Errorf("the two answers differ: %q and %q", wrongPassword, unknownAddress)
	}
}

func TestUnknownAddressStillCostsTime(t *testing.T) {
	h := newHarness(t)
	h.person(t, "joris@example.org")

	started := time.Now()
	h.service.SignIn(t.Context(), "nobody@example.org", password, "10.0.0.1", false)
	unknown := time.Since(started)

	started = time.Now()
	h.service.SignIn(t.Context(), "joris@example.org", "not the password at all", "10.0.0.2", false)
	wrong := time.Since(started)

	// Not a timing-attack proof, only that the unknown case is not obviously
	// free: it should be within an order of magnitude of a real check.
	if unknown < wrong/4 {
		t.Errorf("an unknown address took %v and a wrong password %v, which tells an attacker which is which", unknown, wrong)
	}
}

func TestRememberedSessionsLastLonger(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	_, ordinary, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	_, remembered, err := h.service.Sessions().Issue(t.Context(), person.ID, true)
	if err != nil {
		t.Fatal(err)
	}

	if !remembered.ExpiresAt.After(ordinary.ExpiresAt) {
		t.Errorf("a remembered session expires at %v and an ordinary one at %v",
			remembered.ExpiresAt, ordinary.ExpiresAt)
	}
	if !remembered.Remembered || ordinary.Remembered {
		t.Error("the stored rows do not record which was remembered")
	}
}

func TestStoredSessionIsNotTheCookieValue(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	token, _, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := h.db.Pool().Query(t.Context(), "SELECT encode(token_hash, 'base64') FROM sessions")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var stored string
		if err := rows.Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(stored, token) {
			t.Error("the stored session contains the value the cookie carries")
		}
	}
}

func TestLookupRefusesEachBadSession(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	t.Run("never issued", func(t *testing.T) {
		if _, err := h.service.Lookup(t.Context(), "nothing-like-a-session"); err != nil {
			t.Fatalf("Lookup() error = %v, want nil with an anonymous actor", err)
		}
		actor, _ := h.service.Lookup(t.Context(), "nothing-like-a-session")
		if !actor.Anonymous {
			t.Error("an unissued value identified somebody")
		}
	})

	t.Run("revoked", func(t *testing.T) {
		token, _, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.service.Sessions().Revoke(t.Context(), token); err != nil {
			t.Fatal(err)
		}

		actor, _ := h.service.Lookup(t.Context(), token)
		if !actor.Anonymous {
			t.Error("a revoked session still identified somebody")
		}
	})

	t.Run("expired", func(t *testing.T) {
		token, session, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := h.db.Pool().Exec(t.Context(),
			"UPDATE sessions SET expires_at = now() - interval '1 minute' WHERE id = $1", session.ID); err != nil {
			t.Fatal(err)
		}

		actor, _ := h.service.Lookup(t.Context(), token)
		if !actor.Anonymous {
			t.Error("an expired session still identified somebody")
		}
	})

	t.Run("empty", func(t *testing.T) {
		actor, _ := h.service.Lookup(t.Context(), "")
		if !actor.Anonymous {
			t.Error("an empty value identified somebody")
		}
	})
}

func TestUsingASessionDoesNotExtendIt(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	token, session, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := h.service.Lookup(t.Context(), token); err != nil {
		t.Fatal(err)
	}

	var expires time.Time
	var lastSeen time.Time
	if err := h.db.Pool().QueryRow(t.Context(),
		"SELECT expires_at, last_seen_at FROM sessions WHERE id = $1", session.ID).
		Scan(&expires, &lastSeen); err != nil {
		t.Fatal(err)
	}

	if !expires.Equal(session.ExpiresAt) {
		t.Errorf("expiry moved from %v to %v by using the session", session.ExpiresAt, expires)
	}
	if lastSeen.Before(session.ExpiresAt.Add(-auth.SessionLifetime)) {
		t.Error("the time the session was last used was not recorded")
	}
}

func TestRevokeAllForAPerson(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	for range 3 {
		if _, _, err := h.service.Sessions().Issue(t.Context(), person.ID, false); err != nil {
			t.Fatal(err)
		}
	}

	if count, _ := h.service.Sessions().LiveCount(t.Context(), person.ID); count != 3 {
		t.Fatalf("live sessions = %d, want 3", count)
	}

	if err := h.service.Sessions().RevokeAll(t.Context(), person.ID); err != nil {
		t.Fatal(err)
	}
	if count, _ := h.service.Sessions().LiveCount(t.Context(), person.ID); count != 0 {
		t.Errorf("live sessions after revoking all = %d, want 0", count)
	}
}

func TestPersonWithNoPassword(t *testing.T) {
	h := newHarness(t)

	person, err := h.service.People().Create(t.Context(), "invited@example.org", "Invited", "")
	if err != nil {
		t.Fatal(err)
	}
	if person.HasPassword {
		t.Error("a person created with no password has one")
	}

	if _, _, _, err := h.service.SignIn(t.Context(), "invited@example.org", password, "10.0.0.1", false); !errors.Is(err, httpapi.ErrSignInRefused) {
		t.Errorf("SignIn() for a person with no password = %v, want a refusal", err)
	}

	// Everything that reads a session still works for them.
	token, _, err := h.service.Sessions().Issue(t.Context(), person.ID, false)
	if err != nil {
		t.Fatalf("issuing a session for a person with no password failed: %v", err)
	}
	actor, err := h.service.Lookup(t.Context(), token)
	if err != nil || actor.Anonymous {
		t.Errorf("Lookup() = %+v, %v, want the person", actor, err)
	}
	if err := h.service.SignOut(t.Context(), token); err != nil {
		t.Errorf("SignOut() error = %v, want nil", err)
	}
}

func TestAcceptingAnInvitation(t *testing.T) {
	h := newHarness(t)

	person, err := h.service.People().Create(t.Context(), "invited@example.org", "Invited", "")
	if err != nil {
		t.Fatal(err)
	}

	token, _, err := h.service.Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := h.service.SetPasswordWithToken(t.Context(), token, password); err != nil {
		t.Fatalf("SetPasswordWithToken() error = %v, want nil", err)
	}

	if _, _, _, err := h.service.SignIn(t.Context(), "invited@example.org", password, "10.0.0.1", false); err != nil {
		t.Errorf("SignIn() after accepting an invitation error = %v, want nil", err)
	}
}

func TestBadTokensAreRefusedAlike(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	spent, _, err := h.service.Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := h.service.SetPasswordWithToken(t.Context(), spent, password); err != nil {
		t.Fatal(err)
	}

	expired, _, err := h.service.Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.db.Pool().Exec(t.Context(),
		"UPDATE credential_tokens SET expires_at = now() - interval '1 minute' WHERE spent_at IS NULL"); err != nil {
		t.Fatal(err)
	}

	answers := map[string]error{
		"spent":        failureOf(h.service.SetPasswordWithToken(t.Context(), spent, password)),
		"expired":      failureOf(h.service.SetPasswordWithToken(t.Context(), expired, password)),
		"never issued": failureOf(h.service.SetPasswordWithToken(t.Context(), "nothing-like-a-token", password)),
	}

	for name, err := range answers {
		if !errors.Is(err, httpapi.ErrLinkUnusable) {
			t.Errorf("a %s link gave %v, want ErrLinkUnusable", name, err)
		}
	}
	if answers["spent"].Error() != answers["never issued"].Error() {
		t.Error("a spent link and one that never existed give different answers")
	}
}

func TestResetEndsOtherSessions(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	token, _, err := h.service.Sessions().Issue(t.Context(), person.ID, true)
	if err != nil {
		t.Fatal(err)
	}

	reset, _, err := h.service.Credentials().Issue(t.Context(), auth.KindPasswordReset, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := h.service.SetPasswordWithToken(t.Context(), reset, "a different long passphrase"); err != nil {
		t.Fatal(err)
	}

	actor, _ := h.service.Lookup(t.Context(), token)
	if !actor.Anonymous {
		t.Error("a session survived a password reset")
	}
}

func TestResetSaysNothingAboutTheAddress(t *testing.T) {
	h := newHarness(t)
	h.person(t, "joris@example.org")

	known := h.service.RequestPasswordReset(t.Context(), "joris@example.org", "10.0.0.1")
	unknown := h.service.RequestPasswordReset(t.Context(), "nobody@example.org", "10.0.0.2")

	if known != nil || unknown != nil {
		t.Errorf("the answers differ: known = %v, unknown = %v", known, unknown)
	}
}

func TestLinkIsLoggedNotReturned(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	if err := h.service.Invite(t.Context(), person.ID); err != nil {
		t.Fatal(err)
	}

	logged := h.logs.String()
	if !strings.Contains(logged, "set-password?token=") {
		t.Errorf("the log does not carry the link:\n%s", logged)
	}
	if !strings.Contains(logged, "https://chatcms.example") {
		t.Error("the logged link does not use the site's own address")
	}
}

func TestTokenIsStoredAsAHash(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	token, _, err := h.service.Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}

	var stored string
	if err := h.db.Pool().QueryRow(t.Context(),
		"SELECT encode(token_hash, 'base64') FROM credential_tokens LIMIT 1").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, token) {
		t.Error("the stored link contains the value from the link")
	}
}

func TestSignInIsRateLimited(t *testing.T) {
	h := newHarness(t)
	h.person(t, "joris@example.org")

	var lastErr error
	for range auth.AttemptsPerAddress + 2 {
		_, _, _, lastErr = h.service.SignIn(t.Context(), "joris@example.org", "wrong one entirely", "10.0.0.1", false)
	}

	if !errors.Is(lastErr, httpapi.ErrTooManyAttempts) {
		t.Fatalf("after many attempts the answer is %v, want ErrTooManyAttempts", lastErr)
	}

	// Even the right password is refused while the limit holds, and the
	// refusal says nothing about the address.
	_, _, _, err := h.service.SignIn(t.Context(), "joris@example.org", password, "10.0.0.1", false)
	if !errors.Is(err, httpapi.ErrTooManyAttempts) {
		t.Errorf("SignIn() while limited = %v, want ErrTooManyAttempts", err)
	}
}

// failureOf keeps only what a table of refusals is about, now that setting a
// password also hands back a session.
func failureOf(_ httpapi.Actor, _ string, _ time.Time, err error) error { return err }

func TestSigningOutEverywhereEndsEveryoneOfYours(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	// Three screens open, which is what somebody who has left a session
	// somewhere has.
	for range 3 {
		if _, _, _, err := h.service.SignIn(t.Context(), "joris@example.org", password, "a-caller", false); err != nil {
			t.Fatal(err)
		}
	}

	somebodyElse := h.person(t, "marieke@example.org")
	if _, _, _, err := h.service.SignIn(t.Context(), "marieke@example.org", password, "a-caller", false); err != nil {
		t.Fatal(err)
	}

	live, err := h.service.Sessions().LiveCount(t.Context(), person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if live != 3 {
		t.Fatalf("%d sessions before, want 3", live)
	}

	if err := h.service.SignOutEverywhere(t.Context(), httpapi.Actor{ID: person.ID.String()}); err != nil {
		t.Fatalf("SignOutEverywhere() error = %v, want nil", err)
	}

	left, err := h.service.Sessions().LiveCount(t.Context(), person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Errorf("%d sessions left, want none", left)
	}

	// Nobody else's doors were shut.
	theirs, err := h.service.Sessions().LiveCount(t.Context(), somebodyElse.ID)
	if err != nil {
		t.Fatal(err)
	}
	if theirs != 1 {
		t.Errorf("somebody else has %d sessions, want 1 untouched", theirs)
	}
}

// aTinyPicture is the smallest real PNG, so that what is stored is a picture
// rather than bytes named like one.
func aTinyPicture(t *testing.T) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestAPictureIsStoredSmallAndSquare(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	if err := h.service.People().SetPicture(t.Context(), person.ID, aTinyPicture(t)); err != nil {
		t.Fatalf("SetPicture() error = %v, want nil", err)
	}

	picture, has, err := h.service.People().ReadPicture(t.Context(), person.ID)
	if err != nil || !has {
		t.Fatalf("ReadPicture() = %v, %v, want a picture", has, err)
	}
	if picture.Kind != "image/png" {
		t.Errorf("kind = %q, want image/png", picture.Kind)
	}

	// It is kept at one size, so what somebody's phone produced is not what
	// the database holds.
	decoded, _, err := image.Decode(bytes.NewReader(picture.Content))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != auth.PictureSize || decoded.Bounds().Dy() != auth.PictureSize {
		t.Errorf("stored %dx%d, want %d square", decoded.Bounds().Dx(), decoded.Bounds().Dy(), auth.PictureSize)
	}

	if err := h.service.People().RemovePicture(t.Context(), person.ID); err != nil {
		t.Fatal(err)
	}
	if _, has, _ := h.service.People().ReadPicture(t.Context(), person.ID); has {
		t.Error("the picture is still there after being removed")
	}
}

func TestWhatIsNotAPictureIsRefused(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	for what, content := range map[string][]byte{
		"a text file": []byte("this is not a picture, whatever it was called"),
		"empty":       {},
		"a drawing":   []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
	} {
		if err := h.service.People().SetPicture(t.Context(), person.ID, content); !errors.Is(err, auth.ErrNotAPicture) {
			t.Errorf("SetPicture() with %s error = %v, want ErrNotAPicture", what, err)
		}
	}

	if _, has, _ := h.service.People().ReadPicture(t.Context(), person.ID); has {
		t.Error("something that is not a picture was stored")
	}
}

func TestAnAddressChangesOnlyOnceItIsProven(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")
	who := httpapi.Actor{ID: person.ID.String(), Email: person.Email}

	if err := h.service.AskForANewAddress(t.Context(), who, "joris@korenaar.example"); err != nil {
		t.Fatalf("AskForANewAddress() error = %v, want nil", err)
	}

	// Until the link is followed, the old address is the one that works.
	if _, _, _, err := h.service.SignIn(t.Context(), "joris@example.org", password, "a-caller", false); err != nil {
		t.Errorf("the old address stopped working before the new one was proven: %v", err)
	}
	if _, _, _, err := h.service.SignIn(t.Context(), "joris@korenaar.example", password, "a-caller", false); err == nil {
		t.Error("the new address worked before it was proven")
	}
}

func TestAnAddressSomebodyElseHoldsIsRefused(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")
	h.person(t, "taken@example.org")

	err := h.service.AskForANewAddress(t.Context(),
		httpapi.Actor{ID: person.ID.String(), Email: person.Email}, "taken@example.org")

	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("asking for an address somebody holds error = %v, want a refusal", err)
	}
}

func TestOneSessionIsEndedAndTheOthersAreNot(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")
	who := httpapi.Actor{ID: person.ID.String(), Email: person.Email}

	for range 3 {
		if _, _, _, err := h.service.SignIn(t.Context(), "joris@example.org", password, "a-caller", false); err != nil {
			t.Fatal(err)
		}
	}

	places, err := h.service.SessionsFor(t.Context(), who, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(places) != 3 {
		t.Fatalf("%d sessions, want 3", len(places))
	}

	if err := h.service.EndSession(t.Context(), who, places[0].ID.String()); err != nil {
		t.Fatalf("EndSession() error = %v, want nil", err)
	}

	left, err := h.service.SessionsFor(t.Context(), who, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 2 {
		t.Errorf("%d sessions left, want 2", len(left))
	}
}

func TestSomebodyCannotEndAnotherPersonsSession(t *testing.T) {
	h := newHarness(t)
	mine := h.person(t, "joris@example.org")
	theirs := h.person(t, "marieke@example.org")

	if _, _, _, err := h.service.SignIn(t.Context(), "marieke@example.org", password, "a-caller", false); err != nil {
		t.Fatal(err)
	}

	places, err := h.service.SessionsFor(t.Context(),
		httpapi.Actor{ID: theirs.ID.String(), Email: theirs.Email}, "")
	if err != nil || len(places) != 1 {
		t.Fatalf("reading their sessions = %v, %v", len(places), err)
	}

	// Asked for by somebody else, naming their session: the person is part
	// of the question, so this changes nothing.
	if err := h.service.EndSession(t.Context(),
		httpapi.Actor{ID: mine.ID.String(), Email: mine.Email}, places[0].ID.String()); err != nil {
		t.Fatalf("EndSession() error = %v", err)
	}

	still, err := h.service.SessionsFor(t.Context(),
		httpapi.Actor{ID: theirs.ID.String(), Email: theirs.Email}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(still) != 1 {
		t.Error("one account ended another's session")
	}
}

func TestAnAppearanceIsKeptWithTheAccount(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")
	who := httpapi.Actor{ID: person.ID.String(), Email: person.Email}

	// Nothing chosen is not a third appearance: it means the machine decides.
	chosen, err := h.service.AppearanceFor(t.Context(), person.ID)
	if err != nil || chosen != "" {
		t.Fatalf("a new account has %q, %v; want nothing chosen", chosen, err)
	}

	if err := h.service.SetAppearance(t.Context(), who, "dark"); err != nil {
		t.Fatalf("SetAppearance() error = %v, want nil", err)
	}

	chosen, err = h.service.AppearanceFor(t.Context(), person.ID)
	if err != nil || chosen != "dark" {
		t.Errorf("after choosing dark it is %q, %v", chosen, err)
	}

	// And it can be given back to the machine.
	if err := h.service.SetAppearance(t.Context(), who, ""); err != nil {
		t.Fatal(err)
	}
	if chosen, _ := h.service.AppearanceFor(t.Context(), person.ID); chosen != "" {
		t.Errorf("after following the machine again it is %q", chosen)
	}
}

func TestAnAppearanceThatIsNotOneIsRefused(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "joris@example.org")

	err := h.service.SetAppearance(t.Context(),
		httpapi.Actor{ID: person.ID.String(), Email: person.Email}, "purple")

	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("SetAppearance() with something that is not one error = %v, want a refusal", err)
	}
}

func TestALanguageIsKeptWithTheAccount(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "marieke@example.org")
	who := httpapi.Actor{ID: person.ID.String(), Email: person.Email}

	// Nothing chosen is not a third language: it means the browser decides,
	// which is what a Dutch customer gets before they have found the screen.
	chosen, err := h.service.LanguageFor(t.Context(), person.ID)
	if err != nil || chosen != "" {
		t.Fatalf("a new account has %q, %v; want nothing chosen", chosen, err)
	}

	if err := h.service.SetLanguage(t.Context(), who, "nl"); err != nil {
		t.Fatalf("SetLanguage() error = %v, want nil", err)
	}

	chosen, err = h.service.LanguageFor(t.Context(), person.ID)
	if err != nil || chosen != "nl" {
		t.Errorf("after choosing Dutch it is %q, %v", chosen, err)
	}

	// And it can be given back to the browser.
	if err := h.service.SetLanguage(t.Context(), who, ""); err != nil {
		t.Fatal(err)
	}
	if chosen, _ := h.service.LanguageFor(t.Context(), person.ID); chosen != "" {
		t.Errorf("after following the browser again it is %q", chosen)
	}
}

func TestALanguageTheInterfaceHasNotGotIsRefused(t *testing.T) {
	h := newHarness(t)
	person := h.person(t, "marieke@example.org")
	who := httpapi.Actor{ID: person.ID.String(), Email: person.Email}

	err := h.service.SetLanguage(t.Context(), who, "fr")
	var refused *httpapi.RefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("SetLanguage(\"fr\") error = %v, want a refusal", err)
	}
	// The screen says a refusal in whatever language it is in, which it can
	// only do if the refusal is named.
	if refused.Key == "" {
		t.Error("the refusal has no key, so no screen can say it in Dutch")
	}

	if chosen, _ := h.service.LanguageFor(t.Context(), person.ID); chosen != "" {
		t.Errorf("the refusal set %q anyway", chosen)
	}
}
