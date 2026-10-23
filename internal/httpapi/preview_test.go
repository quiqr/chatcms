package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubPreview stands in for a site's preview, answering with the path it was
// handed so that a test can see what the preview would have been asked for.
type stubPreview struct {
	err   error
	asked []string
	seen  []string
}

func (s *stubPreview) PreviewHandler(_ context.Context, _ Actor, organisation, project, under string) (http.Handler, error) {
	s.asked = append(s.asked, organisation+"/"+project+" under "+under)
	if s.err != nil {
		return nil, s.err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.seen = append(s.seen, r.URL.Path)
		w.Write([]byte("the site at " + r.URL.Path))
	}), nil
}

func withPreview(service PreviewService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:  func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:    &stubAuth{known: known},
		Preview: service,
	})
}

func askPreview(t *testing.T, handler http.Handler, path string, signedIn bool) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	if signedIn {
		request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestAPreviewNeedsAPerson(t *testing.T) {
	handler := withPreview(&stubPreview{}, false)

	recorder := askPreview(t, handler, "/preview/technative/korenaar/", false)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401; an unpublished site must not be readable by anybody", recorder.Code)
	}
}

func TestThePreviewServesTheWholePathItIsReachedAt(t *testing.T) {
	// The preview is told to serve the site under this address, so the path
	// is forwarded whole: stripping it would ask for a page it does not have.
	for _, path := range []string{
		"/preview/technative/korenaar",
		"/preview/technative/korenaar/",
		"/preview/technative/korenaar/news/",
		"/preview/technative/korenaar/css/site.css",
	} {
		service := &stubPreview{}
		handler := withPreview(service, true)

		recorder := askPreview(t, handler, path, true)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s = %d, want 200", path, recorder.Code)
		}
		if len(service.seen) != 1 || service.seen[0] != path {
			t.Errorf("%s was handed %v, want the whole path", path, service.seen)
		}
		want := "technative/korenaar under /preview/technative/korenaar/"
		if len(service.asked) != 1 || service.asked[0] != want {
			t.Errorf("%s asked for %v, want %q", path, service.asked, want)
		}
	}
}

func TestAPreviewPathThatNamesNoProjectIsNotFound(t *testing.T) {
	handler := withPreview(&stubPreview{}, true)

	for _, path := range []string{"/preview/", "/preview/technative", "/preview/technative/"} {
		recorder := askPreview(t, handler, path, true)
		if recorder.Code != http.StatusNotFound {
			t.Errorf("%s = %d, want 404", path, recorder.Code)
		}
	}
}

func TestAPreviewNobodyMaySeeIsNotFound(t *testing.T) {
	handler := withPreview(&stubPreview{err: ErrMissing}, true)

	recorder := askPreview(t, handler, "/preview/technative/korenaar/", true)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestTooManyPreviewsIsSaidWithAReason(t *testing.T) {
	handler := withPreview(&stubPreview{
		err: &RefusedError{Reason: "too many sites are being previewed at once"},
	}, true)

	recorder := askPreview(t, handler, "/preview/technative/korenaar/", true)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if recorder.Body.Len() == 0 {
		t.Error("the refusal says nothing")
	}
}
