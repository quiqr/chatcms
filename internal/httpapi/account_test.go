package httpapi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubAccount struct {
	asked   []string
	err     error
	picture []byte
	has     bool
}

func (s *stubAccount) note(what string) { s.asked = append(s.asked, what) }

func (s *stubAccount) ChangeOwnName(_ context.Context, _ Actor, name string) (Actor, error) {
	s.note("name " + name)
	return Actor{ID: "person-1", DisplayName: name}, s.err
}

func (s *stubAccount) ChangeOwnPassword(context.Context, Actor, string, string) error {
	s.note("password")
	return s.err
}

func (s *stubAccount) TurnAccountOff(context.Context, Actor, uuid.UUID) error { return s.err }
func (s *stubAccount) TurnAccountOn(context.Context, uuid.UUID) error        { return s.err }

func (s *stubAccount) SetPicture(_ context.Context, _ Actor, content []byte) error {
	s.note("picture of " + string(rune('0'+len(content)%10)))
	return s.err
}

func (s *stubAccount) RemovePicture(context.Context, Actor) error {
	s.note("remove picture")
	return s.err
}

func (s *stubAccount) ReadPicture(context.Context, string) ([]byte, string, bool, error) {
	s.note("read picture")
	return s.picture, "image/png", s.has, s.err
}

func (s *stubAccount) AskForANewAddress(_ context.Context, _ Actor, address string) error {
	s.note("address " + address)
	return s.err
}

func (s *stubAccount) SessionViews(context.Context, Actor, string) ([]SessionView, error) {
	s.note("sessions")
	return []SessionView{
		{ID: "session-1", LastSeen: time.Now(), ThisOne: true},
		{ID: "session-2", LastSeen: time.Now()},
	}, s.err
}

func (s *stubAccount) EndSession(_ context.Context, _ Actor, sessionID string) error {
	s.note("end " + sessionID)
	return s.err
}

func withAccount(service AccountService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "joris@example.org"}
	}

	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Auth:    &stubAuth{known: known},
		Account: service,
	})
}

func aPictureUpload(t *testing.T, content []byte) (io.Reader, string) {
	t.Helper()

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	part, err := form.CreateFormFile("picture", "me.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	form.Close()
	return body, form.FormDataContentType()
}

func uploading(t *testing.T, handler http.Handler, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	body, kind := aPictureUpload(t, content)
	request := httptest.NewRequest(http.MethodPut, "/api/account/picture", body)
	request.Header.Set("Content-Type", kind)
	request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
	request.Header.Set(RequestTokenHeader, "a-token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestEveryAccountEndpointNeedsSomebody(t *testing.T) {
	handler := withAccount(&stubAccount{}, false)

	for _, each := range []struct {
		method, path string
		body         any
	}{
		{http.MethodPut, "/api/account/name", map[string]any{"displayName": "A"}},
		{http.MethodPut, "/api/account/password", map[string]any{"current": "a", "next": "b"}},
		{http.MethodDelete, "/api/account/picture", nil},
		{http.MethodPost, "/api/account/address", map[string]any{"email": "a@b.c"}},
		{http.MethodGet, "/api/account/sessions", nil},
		{http.MethodDelete, "/api/account/sessions/session-1", nil},
		{http.MethodGet, "/api/people/person-1/picture", nil},
	} {
		recorder := call(t, handler, each.method, each.path, each.body)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", each.method, each.path, recorder.Code)
		}
	}
}

func TestAPictureIsUploadedAndRemoved(t *testing.T) {
	service := &stubAccount{}
	handler := withAccount(service, true)

	uploaded := uploading(t, handler, []byte("some bytes"))
	if uploaded.Code != http.StatusNoContent {
		t.Errorf("uploading = %d, want 204: %s", uploaded.Code, uploaded.Body)
	}

	removed := signedInCall(t, handler, http.MethodDelete, "/api/account/picture", nil)
	if removed.Code != http.StatusNoContent {
		t.Errorf("removing = %d, want 204: %s", removed.Code, removed.Body)
	}

	if len(service.asked) != 2 || service.asked[1] != "remove picture" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestAPictureThatIsRefusedSaysWhy(t *testing.T) {
	handler := withAccount(&stubAccount{
		err: &RefusedError{Reason: "that is not a picture"},
	}, true)

	recorder := uploading(t, handler, []byte("not a picture at all"))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", recorder.Code, recorder.Body)
	}
	if !strings.Contains(recorder.Body.String(), "not a picture") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestAPictureIsServedAndMissingOneIsNothingThere(t *testing.T) {
	with := withAccount(&stubAccount{picture: []byte("png bytes"), has: true}, true)
	served := signedInCall(t, with, http.MethodGet, "/api/people/person-1/picture", nil)
	if served.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", served.Code)
	}
	if served.Header().Get("Content-Type") != "image/png" {
		t.Errorf("content type = %q, want image/png", served.Header().Get("Content-Type"))
	}

	without := withAccount(&stubAccount{}, true)
	missing := signedInCall(t, without, http.MethodGet, "/api/people/person-2/picture", nil)
	if missing.Code != http.StatusNotFound {
		t.Errorf("a person with no picture = %d, want 404", missing.Code)
	}
}

func TestAskingForANewAddressPassesItOn(t *testing.T) {
	service := &stubAccount{}
	handler := withAccount(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/account/address",
		map[string]any{"email": "new@example.org"})

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", recorder.Code, recorder.Body)
	}
	if len(service.asked) != 1 || service.asked[0] != "address new@example.org" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestAnAddressSomebodyElseHoldsIsRefusedWithAReason(t *testing.T) {
	handler := withAccount(&stubAccount{
		err: &RefusedError{Reason: "that address already belongs to an account"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/account/address",
		map[string]any{"email": "taken@example.org"})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "already belongs") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestWhereSomebodyIsSignedInMarksTheOneReading(t *testing.T) {
	handler := withAccount(&stubAccount{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, "/api/account/sessions", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"thisOne":true`) {
		t.Errorf("body = %s, want the session being read marked", recorder.Body)
	}
}

func TestOneSessionIsEnded(t *testing.T) {
	service := &stubAccount{}
	handler := withAccount(service, true)

	recorder := signedInCall(t, handler, http.MethodDelete, "/api/account/sessions/session-2", nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", recorder.Code, recorder.Body)
	}
	if len(service.asked) != 1 || service.asked[0] != "end session-2" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestAccountEndpointsRefuseTheWrongMethod(t *testing.T) {
	handler := withAccount(&stubAccount{}, true)

	for _, path := range []string{
		"/api/account/name",
		"/api/account/password",
		"/api/account/picture",
		"/api/account/address",
		"/api/account/sessions",
	} {
		recorder := signedInCall(t, handler, http.MethodPatch, path, map[string]any{})
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("PATCH %s = %d, want 405", path, recorder.Code)
		}
	}
}
