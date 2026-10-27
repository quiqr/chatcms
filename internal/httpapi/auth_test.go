package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// stubAuth stands in for the identity service, so the HTTP layer can be
// tested without a database. What it returns is what the handlers have to
// turn into status codes, cookies and wording.
type stubAuth struct {
	person       Actor
	signInErr    error
	signOutErr   error
	lookupErr    error
	resetErr     error
	setErr       error
	signedOut    []string
	resetAsked   []string
	setRequested [][2]string
	known        map[string]Actor
}

func (s *stubAuth) SignIn(_ context.Context, email, _, _ string, _ bool) (Actor, string, time.Time, error) {
	if s.signInErr != nil {
		return Actor{}, "", time.Time{}, s.signInErr
	}
	person := s.person
	person.Email = email
	return person, "a-session-value", time.Now().Add(time.Hour), nil
}

func (s *stubAuth) SignOut(_ context.Context, token string) error {
	s.signedOut = append(s.signedOut, token)
	return s.signOutErr
}

func (s *stubAuth) SignOutEverywhere(_ context.Context, who Actor) error {
	s.signedOut = append(s.signedOut, "everywhere for "+who.ID)
	return s.signOutErr
}

func (s *stubAuth) Lookup(_ context.Context, token string) (Actor, error) {
	if s.lookupErr != nil {
		return Anonymous(), s.lookupErr
	}
	if actor, ok := s.known[token]; ok {
		return actor, nil
	}
	return Anonymous(), nil
}

func (s *stubAuth) RequestPasswordReset(_ context.Context, email, _ string) error {
	s.resetAsked = append(s.resetAsked, email)
	return s.resetErr
}

func (s *stubAuth) SetPasswordWithToken(_ context.Context, token, password string) (Actor, string, time.Time, error) {
	s.setRequested = append(s.setRequested, [2]string{token, password})
	if s.setErr != nil {
		return Anonymous(), "", time.Time{}, s.setErr
	}
	return Actor{ID: "person-1", Email: "invited@example.org"},
		"a-session-token", time.Now().Add(time.Hour), nil
}

func withAuth(auth AuthService) http.Handler {
	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   auth,
	})
}

func call(t *testing.T, handler http.Handler, method, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = strings.NewReader(string(encoded))
	}

	request := httptest.NewRequest(method, path, reader)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	if !safeMethod(method) {
		request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
		request.Header.Set(RequestTokenHeader, "a-token")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func cookieNamed(response *http.Response, name string) *http.Cookie {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestSignInSetsTheSessionCookie(t *testing.T) {
	auth := &stubAuth{person: Actor{ID: "1", DisplayName: "Joris"}}
	recorder := call(t, withAuth(auth), http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": "a long enough passphrase"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}

	cookie := cookieNamed(recorder.Result(), SessionCookie)
	if cookie == nil || cookie.Value != "a-session-value" {
		t.Fatalf("session cookie = %+v, want the issued value", cookie)
	}

	var person personResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &person); err != nil {
		t.Fatal(err)
	}
	if person.Email != "joris@example.org" {
		t.Errorf("email = %q", person.Email)
	}
}

func TestSignInRevokesWhateverItArrivedWith(t *testing.T) {
	auth := &stubAuth{person: Actor{ID: "1"}}

	call(t, withAuth(auth), http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": "a long enough passphrase"},
		&http.Cookie{Name: SessionCookie, Value: "an-older-session"})

	if len(auth.signedOut) != 1 || auth.signedOut[0] != "an-older-session" {
		t.Errorf("signed out %v, want the session the request arrived with", auth.signedOut)
	}
}

func TestSignInTurnsEachFailureIntoItsOwnAnswer(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
		code   string
	}{
		"refused": {ErrSignInRefused, http.StatusUnauthorized, CodeUnauthorized},
		"limited": {ErrTooManyAttempts, http.StatusTooManyRequests, CodeTooManyAttempts},
		"broken":  {context.DeadlineExceeded, http.StatusInternalServerError, CodeInternal},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := call(t, withAuth(&stubAuth{signInErr: tc.err}), http.MethodPost, "/api/sign-in",
				map[string]any{"email": "joris@example.org", "password": "a long enough passphrase"})

			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.status)
			}
			if got := decodeError(t, recorder.Body.Bytes()).Code; got != tc.code {
				t.Errorf("code = %q, want %q", got, tc.code)
			}
			if cookieNamed(recorder.Result(), SessionCookie) != nil {
				t.Error("a session cookie was set although the sign-in failed")
			}
		})
	}
}

func TestSignInRefusalSaysNothingAboutTheAddress(t *testing.T) {
	recorder := call(t, withAuth(&stubAuth{signInErr: ErrSignInRefused}), http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": "wrong one entirely"})

	body := strings.ToLower(recorder.Body.String())
	for _, word := range []string{"exist", "unknown", "no such", "not found"} {
		if strings.Contains(body, word) {
			t.Errorf("the answer hints at whether the address is known: %s", recorder.Body)
		}
	}
}

func TestSignInRejectsAMalformedBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader("{not json"))
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
	request.Header.Set(RequestTokenHeader, "a-token")

	recorder := httptest.NewRecorder()
	withAuth(&stubAuth{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestSignOutClearsAndRevokes(t *testing.T) {
	auth := &stubAuth{}
	recorder := call(t, withAuth(auth), http.MethodPost, "/api/sign-out", nil,
		&http.Cookie{Name: SessionCookie, Value: "a-session-value"})

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if len(auth.signedOut) != 1 {
		t.Errorf("signed out %v, want the session", auth.signedOut)
	}

	cookie := cookieNamed(recorder.Result(), SessionCookie)
	if cookie == nil || cookie.Value != "" || cookie.MaxAge >= 0 {
		t.Errorf("session cookie = %+v, want it cleared", cookie)
	}
}

func TestSignOutWithNoSessionStillSucceeds(t *testing.T) {
	auth := &stubAuth{}
	recorder := call(t, withAuth(auth), http.MethodPost, "/api/sign-out", nil)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", recorder.Code)
	}
	if len(auth.signedOut) != 0 {
		t.Errorf("something was signed out: %v", auth.signedOut)
	}
}

func TestCurrentPersonNeedsASession(t *testing.T) {
	auth := &stubAuth{known: map[string]Actor{
		"a-session-value": {ID: "1", Email: "joris@example.org", DisplayName: "Joris"},
	}}

	anonymous := call(t, withAuth(auth), http.MethodGet, "/api/me", nil)
	if anonymous.Code != http.StatusUnauthorized {
		t.Errorf("status without a session = %d, want 401", anonymous.Code)
	}

	signedIn := call(t, withAuth(auth), http.MethodGet, "/api/me", nil,
		&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	if signedIn.Code != http.StatusOK {
		t.Fatalf("status with a session = %d, want 200", signedIn.Code)
	}

	var person personResponse
	if err := json.Unmarshal(signedIn.Body.Bytes(), &person); err != nil {
		t.Fatal(err)
	}
	if person.Email != "joris@example.org" {
		t.Errorf("email = %q", person.Email)
	}
}

func TestARequestWithAnUnusableSessionSeesNobody(t *testing.T) {
	auth := &stubAuth{known: map[string]Actor{}}

	recorder := call(t, withAuth(auth), http.MethodGet, "/api/me", nil,
		&http.Cookie{Name: SessionCookie, Value: "not-a-session"})

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}

func TestAFailingSessionLookupIsNotFatal(t *testing.T) {
	auth := &stubAuth{lookupErr: context.DeadlineExceeded}

	recorder := call(t, withAuth(auth), http.MethodGet, "/api/health", nil,
		&http.Cookie{Name: SessionCookie, Value: "a-session-value"})

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want the request to be served as anonymous", recorder.Code)
	}
}

func TestResetAnswersTheSameWayAlways(t *testing.T) {
	auth := &stubAuth{}

	var answers []string
	for _, address := range []string{"joris@example.org", "nobody@example.org"} {
		recorder := call(t, withAuth(auth), http.MethodPost, "/api/password-reset",
			map[string]any{"email": address})
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
		answers = append(answers, recorder.Body.String())
	}

	if answers[0] != answers[1] {
		t.Errorf("the answers differ:\n%s\n%s", answers[0], answers[1])
	}
	if strings.Contains(answers[0], "token") {
		t.Errorf("the answer carries a link: %s", answers[0])
	}
	if len(auth.resetAsked) != 2 {
		t.Errorf("the service was asked %d times, want 2", len(auth.resetAsked))
	}
}

func TestResetIsRateLimited(t *testing.T) {
	recorder := call(t, withAuth(&stubAuth{resetErr: ErrTooManyAttempts}), http.MethodPost,
		"/api/password-reset", map[string]any{"email": "joris@example.org"})

	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", recorder.Code)
	}
}

func TestSetPasswordReportsEachOutcome(t *testing.T) {
	cases := map[string]struct {
		err      error
		status   int
		contains string
	}{
		// Setting a password now signs that person in, so what comes back is
		// who they are rather than an instruction to sign in.
		"accepted":   {nil, http.StatusOK, "invited@example.org"},
		"bad link":   {ErrLinkUnusable, http.StatusBadRequest, "new one"},
		"bad word":   {&PasswordRejectedError{Reason: "a password needs at least 12 characters"}, http.StatusBadRequest, "12"},
		"gone wrong": {context.DeadlineExceeded, http.StatusInternalServerError, "went wrong"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := call(t, withAuth(&stubAuth{setErr: tc.err}), http.MethodPost, "/api/set-password",
				map[string]any{"token": "a-token", "password": "a long enough passphrase"})

			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, tc.status, recorder.Body)
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Errorf("body = %s, want it to contain %q", recorder.Body, tc.contains)
			}
		})
	}
}

func TestSettingAPasswordSignsThemIn(t *testing.T) {
	// An invitation that ends at a sign-in form asks somebody to prove twice
	// what they just proved once.
	recorder := call(t, withAuth(&stubAuth{}), http.MethodPost, "/api/set-password",
		map[string]any{"token": "a-token", "password": "a long enough passphrase"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var carried bool
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == SessionCookie && cookie.Value != "" {
			carried = true
		}
	}
	if !carried {
		t.Error("setting a password did not hand back a session")
	}
}

func TestAuthEndpointsRefuseTheWrongMethod(t *testing.T) {
	handler := withAuth(&stubAuth{})

	for _, path := range []string{"/api/sign-in", "/api/sign-out", "/api/password-reset", "/api/set-password"} {
		recorder := call(t, handler, http.MethodDelete, path, nil)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s with DELETE = %d, want 405", path, recorder.Code)
		}
	}

	if recorder := call(t, handler, http.MethodPost, "/api/me", nil); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("/api/me with POST = %d, want 405", recorder.Code)
	}
}

func TestWithoutAnAuthServiceTheEndpointsAreNotThere(t *testing.T) {
	handler := New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
	})

	recorder := call(t, handler, http.MethodPost, "/api/sign-in", map[string]any{})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 when no identity service is configured", recorder.Code)
	}
}

func TestCallerOfStripsThePort(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/sign-in", nil)
	request.RemoteAddr = "192.0.2.4:51234"

	if got := callerOf(request); got != "192.0.2.4" {
		t.Errorf("callerOf() = %q, want the address without the port", got)
	}

	request.RemoteAddr = "no-port-here"
	if got := callerOf(request); got != "no-port-here" {
		t.Errorf("callerOf() = %q, want the value as it was", got)
	}
}

func TestCapitalise(t *testing.T) {
	cases := map[string]string{
		"a password needs more": "A password needs more",
		"Already capital":       "Already capital",
		"":                      "",
		"12 characters":         "12 characters",
	}

	for in, want := range cases {
		if got := capitalise(in); got != want {
			t.Errorf("capitalise(%q) = %q, want %q", in, got, want)
		}
	}
}


func TestSigningOutEverywhereEndsTheSessionAsking(t *testing.T) {
	// Somebody who thinks their account is not theirs any more wants every
	// door shut, including the one they are standing in.
	auth := &stubAuth{known: map[string]Actor{"a-session-value": {ID: "person-1"}}}

	recorder := signedInCall(t, withAuth(auth), http.MethodPost, "/api/sign-out/everywhere", nil)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", recorder.Code, recorder.Body)
	}
	if len(auth.signedOut) != 1 || auth.signedOut[0] != "everywhere for person-1" {
		t.Errorf("the service was asked %v", auth.signedOut)
	}

	cookie := cookieNamed(recorder.Result(), SessionCookie)
	if cookie == nil || cookie.Value != "" {
		t.Error("the session cookie was not cleared, so the screen stays signed in")
	}
}

func TestSigningOutEverywhereNeedsSomebody(t *testing.T) {
	recorder := call(t, withAuth(&stubAuth{}), http.MethodPost, "/api/sign-out/everywhere", nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}
