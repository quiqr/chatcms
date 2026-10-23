package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubMedia stands in for the picture library.
type stubMedia struct {
	images  []ImageView
	image   ImageView
	content []byte
	kind    string
	err     error
	calls   []string
}

func (s *stubMedia) ListMedia(context.Context, Actor, string, string) ([]ImageView, error) {
	s.calls = append(s.calls, "list")
	return s.images, s.err
}

func (s *stubMedia) AddMedia(_ context.Context, _ Actor, _, _, name string, content []byte) (ImageView, error) {
	s.calls = append(s.calls, "add:"+name+":"+string(content))
	return s.image, s.err
}

func (s *stubMedia) DescribeMedia(_ context.Context, _ Actor, _, _, name, said string) (ImageView, error) {
	s.calls = append(s.calls, "describe:"+name+":"+said)
	return s.image, s.err
}

func (s *stubMedia) RemoveMedia(_ context.Context, _ Actor, _, _, name string, evenIfUsed bool) error {
	s.calls = append(s.calls, "remove:"+name+":"+boolWord(evenIfUsed))
	return s.err
}

func (s *stubMedia) ReadMedia(_ context.Context, _ Actor, _, _, name string, small bool) ([]byte, string, error) {
	s.calls = append(s.calls, "read:"+name+":"+boolWord(small))
	return s.content, s.kind, s.err
}

func withMedia(service MediaService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Media:  service,
	})
}

// sendPicture sends a file the way a browser sends one.
func sendPicture(t *testing.T, handler http.Handler, path, name string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	body := bytes.Buffer{}
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(content)
	form.Close()

	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", form.FormDataContentType())
	request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
	request.Header.Set(RequestTokenHeader, "a-token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestMediaEndpointsNeedAPerson(t *testing.T) {
	handler := withMedia(&stubMedia{}, false)

	for _, pair := range [][2]string{
		{http.MethodGet, projectBase + "/images"},
		{http.MethodGet, projectBase + "/images/logo.png"},
		{http.MethodDelete, projectBase + "/images/logo.png"},
	} {
		recorder := call(t, handler, pair[0], pair[1], nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestThePicturesAreServed(t *testing.T) {
	service := &stubMedia{images: []ImageView{
		{Name: "logo.png", Path: "static/images/logo.png", Size: 2048, Alt: "The sign", UsedOn: []string{"home"}},
	}}
	handler := withMedia(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/images", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var envelope struct {
		Images []ImageView `json:"images"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Images) != 1 || envelope.Images[0].Alt != "The sign" {
		t.Errorf("images = %+v", envelope.Images)
	}
	if len(envelope.Images[0].UsedOn) != 1 {
		t.Error("what uses a picture did not survive")
	}
}

func TestAPictureIsSentAsAForm(t *testing.T) {
	service := &stubMedia{image: ImageView{Name: "logo.png"}}
	handler := withMedia(service, true)

	recorder := sendPicture(t, handler, projectBase+"/images", "Our Logo.png", []byte("not really a picture"))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	if service.calls[0] != "add:Our Logo.png:not really a picture" {
		t.Errorf("calls = %v, want the name and the bytes as sent", service.calls)
	}
}

func TestARequestWithNoPictureIsRefused(t *testing.T) {
	handler := withMedia(&stubMedia{}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/images", map[string]any{})

	if recorder.Code != http.StatusRequestEntityTooLarge && recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want it refused", recorder.Code)
	}
}

func TestAPictureIsReadBackWithItsKind(t *testing.T) {
	service := &stubMedia{content: []byte("the bytes"), kind: "image/png"}
	handler := withMedia(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/images/logo.png", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Content-Type = %q", recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("a picture is served without telling the browser not to guess its kind")
	}
	if recorder.Body.String() != "the bytes" {
		t.Errorf("body = %q", recorder.Body.String())
	}
}

func TestASmallCopyIsAskedForWithASetting(t *testing.T) {
	service := &stubMedia{content: []byte("small"), kind: "image/png"}
	handler := withMedia(service, true)

	signedInCall(t, handler, http.MethodGet, projectBase+"/images/logo.png?small=true", nil)

	if service.calls[0] != "read:logo.png:publishing" {
		t.Errorf("calls = %v, want it asked for small", service.calls)
	}
}

func TestADescriptionIsWritten(t *testing.T) {
	service := &stubMedia{image: ImageView{Name: "logo.png", Alt: "The sign"}}
	handler := withMedia(service, true)

	recorder := signedInCall(t, handler, http.MethodPut, projectBase+"/images/logo.png",
		map[string]any{"alt": "The sign above the door"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if service.calls[0] != "describe:logo.png:The sign above the door" {
		t.Errorf("calls = %v", service.calls)
	}
}

func TestRemovingSaysWhetherToDoItAnyway(t *testing.T) {
	cases := map[string]string{
		projectBase + "/images/logo.png":                 "remove:logo.png:waiting",
		projectBase + "/images/logo.png?evenIfUsed=true": "remove:logo.png:publishing",
	}

	for path, want := range cases {
		service := &stubMedia{}
		handler := withMedia(service, true)

		recorder := signedInCall(t, handler, http.MethodDelete, path, nil)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", recorder.Code)
		}
		if service.calls[0] != want {
			t.Errorf("calls = %v, want %q", service.calls, want)
		}
	}
}

func TestAPictureInUseIsRefusedWithAReason(t *testing.T) {
	handler := withMedia(&stubMedia{
		err: &RefusedError{Reason: "that picture is on 2 pages that are still using it"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodDelete, projectBase+"/images/logo.png", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "still using it") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestAPictureThatIsNotThereIsNotFound(t *testing.T) {
	handler := withMedia(&stubMedia{err: ErrMissing}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/images/absent.png", nil)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}
