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
)

// stubOnboarding stands in for setting a project up, so the HTTP layer can be
// tested without git, a database or a key.
type stubOnboarding struct {
	key        KeyView
	report     ImportView
	connection ConnectionView
	publishing PublishingView
	err        error
	calls      []string
}

func (s *stubOnboarding) record(name string) { s.calls = append(s.calls, name) }

func (s *stubOnboarding) GenerateKey(context.Context, Actor, string, string) (KeyView, error) {
	s.record("generate-key")
	return s.key, s.err
}

func (s *stubOnboarding) ReadKey(context.Context, Actor, string, string) (KeyView, error) {
	s.record("read-key")
	return s.key, s.err
}

func (s *stubOnboarding) ImportSite(_ context.Context, _ Actor, _, _, kind, address, remote string) (ImportView, error) {
	s.record("import:" + kind + ":" + address + ":" + remote)
	return s.report, s.err
}

func (s *stubOnboarding) InspectSite(context.Context, Actor, string, string) (ImportView, error) {
	s.record("inspect")
	return s.report, s.err
}

func (s *stubOnboarding) TestConnection(context.Context, Actor, string, string) (ConnectionView, error) {
	s.record("connection")
	return s.connection, s.err
}

func (s *stubOnboarding) SetPublishing(_ context.Context, _ Actor, _, _, to, branch, form string) (PublishingView, error) {
	s.record("set-publishing:" + to + ":" + branch + ":" + form)
	return s.publishing, s.err
}

func (s *stubOnboarding) ReadPublishing(context.Context, Actor, string, string) (PublishingView, error) {
	s.record("read-publishing")
	return s.publishing, s.err
}

func withOnboarding(service OnboardingService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "implementer@example.org"}
	}

	return New(Options{
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:     func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:       &stubAuth{known: known},
		Onboarding: service,
	})
}

const projectBase = "/api/organisations/technative/projects/korenaar"

func TestOnboardingEndpointsNeedAPerson(t *testing.T) {
	handler := withOnboarding(&stubOnboarding{}, false)

	for _, pair := range [][2]string{
		{http.MethodGet, projectBase + "/key"},
		{http.MethodPost, projectBase + "/key"},
		{http.MethodPost, projectBase + "/import"},
		{http.MethodGet, projectBase + "/site"},
		{http.MethodPost, projectBase + "/connection"},
		{http.MethodGet, projectBase + "/publishing"},
		{http.MethodPut, projectBase + "/publishing"},
	} {
		recorder := call(t, handler, pair[0], pair[1], map[string]any{})
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestMakingAndReadingAKey(t *testing.T) {
	service := &stubOnboarding{key: KeyView{
		PublicKey: "ssh-ed25519 AAAAexample chatcms",
		Advice:    "Add this key to the git host.",
	}}
	handler := withOnboarding(service, true)

	made := signedInCall(t, handler, http.MethodPost, projectBase+"/key", nil)
	if made.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", made.Code)
	}

	var key KeyView
	if err := json.Unmarshal(made.Body.Bytes(), &key); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key.PublicKey, "ssh-ed25519 ") || key.Advice == "" {
		t.Errorf("key = %+v", key)
	}
	if strings.Contains(made.Body.String(), "PRIVATE") {
		t.Error("the answer carries a private key")
	}

	read := signedInCall(t, handler, http.MethodGet, projectBase+"/key", nil)
	if read.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", read.Code)
	}
}

func TestImportingPassesWhatWasAsked(t *testing.T) {
	service := &stubOnboarding{report: ImportView{
		Kind: "repository", IsHugoSite: true, HasModel: true, Ready: true, Pages: 3, Collections: 2,
	}}
	handler := withOnboarding(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/import",
		map[string]any{"kind": "export", "address": "/srv/export", "remote": "git@example.org:a/b.git"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body)
	}

	want := "import:export:/srv/export:git@example.org:a/b.git"
	found := false
	for _, call := range service.calls {
		if call == want {
			found = true
		}
	}
	if !found {
		t.Errorf("calls = %v, want %q", service.calls, want)
	}
}

func TestInspectingASite(t *testing.T) {
	service := &stubOnboarding{report: ImportView{
		Kind: "repository", IsHugoSite: true, HasModel: false,
		Problem:   "This is a Hugo site with no Quiqr model yet.",
		LookedFor: []string{"quiqr/model/base.yaml"},
	}}

	recorder := signedInCall(t, withOnboarding(service, true), http.MethodGet, projectBase+"/site", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "no Quiqr model yet") {
		t.Errorf("body = %s, want the problem in it", recorder.Body)
	}
	if !strings.Contains(recorder.Body.String(), "quiqr/model/base.yaml") {
		t.Errorf("body = %s, want what was looked for", recorder.Body)
	}
}

func TestTestingTheConnection(t *testing.T) {
	service := &stubOnboarding{connection: ConnectionView{
		Good: false, Message: "The host would not accept this key.",
		PublicKey: "ssh-ed25519 AAAAexample",
	}}

	recorder := signedInCall(t, withOnboarding(service, true), http.MethodPost, projectBase+"/connection", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var connection ConnectionView
	if err := json.Unmarshal(recorder.Body.Bytes(), &connection); err != nil {
		t.Fatal(err)
	}
	if connection.Good || connection.Message == "" {
		t.Errorf("connection = %+v", connection)
	}
	if strings.Contains(recorder.Body.String(), "PRIVATE") {
		t.Error("the answer carries a private key")
	}
}

func TestReadingAndSettingPublishing(t *testing.T) {
	service := &stubOnboarding{publishing: PublishingView{
		To: "its-own-line", Branch: "quiqr/publish", Form: "source",
	}}
	handler := withOnboarding(service, true)

	read := signedInCall(t, handler, http.MethodGet, projectBase+"/publishing", nil)
	if read.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", read.Code)
	}

	set := signedInCall(t, handler, http.MethodPut, projectBase+"/publishing",
		map[string]any{"to": "its-own-line", "branch": "releases", "form": "built"})
	if set.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", set.Code)
	}

	want := "set-publishing:its-own-line:releases:built"
	found := false
	for _, call := range service.calls {
		if call == want {
			found = true
		}
	}
	if !found {
		t.Errorf("calls = %v, want %q", service.calls, want)
	}
}

func TestOnboardingFailuresBecomeTheRightAnswers(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
	}{
		"not theirs":  {ErrMissing, http.StatusNotFound},
		"not allowed": {ErrNotAllowed, http.StatusForbidden},
		"refused":     {&RefusedError{Reason: "that address could not be reached"}, http.StatusConflict},
		"broken":      {context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			handler := withOnboarding(&stubOnboarding{err: tc.err}, true)

			for _, pair := range [][2]string{
				{http.MethodPost, projectBase + "/key"},
				{http.MethodGet, projectBase + "/key"},
				{http.MethodGet, projectBase + "/site"},
				{http.MethodPost, projectBase + "/connection"},
				{http.MethodGet, projectBase + "/publishing"},
			} {
				recorder := signedInCall(t, handler, pair[0], pair[1], nil)
				if recorder.Code != tc.status {
					t.Errorf("%s %s = %d, want %d", pair[0], pair[1], recorder.Code, tc.status)
				}
			}
		})
	}
}

func TestOnboardingRejectsAMalformedBody(t *testing.T) {
	handler := withOnboarding(&stubOnboarding{}, true)

	for _, pair := range [][2]string{
		{http.MethodPost, projectBase + "/import"},
		{http.MethodPut, projectBase + "/publishing"},
	} {
		request := httptest.NewRequest(pair[0], pair[1], strings.NewReader("{not json"))
		request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
		request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
		request.Header.Set(RequestTokenHeader, "a-token")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("%s %s = %d, want 400", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestOnboardingRefusesTheWrongMethod(t *testing.T) {
	handler := withOnboarding(&stubOnboarding{}, true)

	for _, path := range []string{"/key", "/import", "/site", "/connection", "/publishing"} {
		recorder := signedInCall(t, handler, http.MethodPatch, projectBase+path, nil)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("PATCH %s = %d, want 405", path, recorder.Code)
		}
	}
}

func TestWithoutAnOnboardingServiceTheEndpointsAreNotThere(t *testing.T) {
	handler := New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: map[string]Actor{"a-session-value": {ID: "person-1"}}},
	})

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/key", nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}
