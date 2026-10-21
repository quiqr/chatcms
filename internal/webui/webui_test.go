package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func builtAssets() fstest.MapFS {
	return fstest.MapFS{
		"index.html":              {Data: []byte("<!doctype html><title>ChatCMS</title>")},
		"assets/index-abc123.js":  {Data: []byte("console.log('hi')")},
		"assets/index-abc123.css": {Data: []byte("body{}")},
		"favicon.ico":             {Data: []byte("icon")},
	}
}

func request(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder
}

func TestAssetCacheHeaders(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodGet, "/assets/index-abc123.js")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != immutableMaxAge {
		t.Errorf("Cache-Control = %q, want %q", got, immutableMaxAge)
	}
}

func TestDocumentIsNotCached(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodGet, "/")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != documentCache {
		t.Errorf("Cache-Control = %q, want %q", got, documentCache)
	}
	if !strings.Contains(recorder.Body.String(), "ChatCMS") {
		t.Errorf("body = %q, want the application document", recorder.Body)
	}
}

func TestUnknownPathServesDocument(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodGet, "/projects/korenaar/content/articles")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 so the browser router can decide", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "ChatCMS") {
		t.Errorf("body = %q, want the application document", recorder.Body)
	}
}

func TestNonHashedFileIsNotCachedForever(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodGet, "/favicon.ico")

	if got := recorder.Header().Get("Cache-Control"); got == immutableMaxAge {
		t.Errorf("Cache-Control = %q, want a short-lived policy outside assets/", got)
	}
}

func TestUnbuiltFrontendServesNotice(t *testing.T) {
	handler := HandlerFor(fstest.MapFS{".gitkeep": {Data: []byte{}}})

	recorder := request(t, handler, http.MethodGet, "/")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "was not built") {
		t.Errorf("body = %q, want it to say the frontend was not built", recorder.Body)
	}
}

func TestOnlyReadMethodsAreServed(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodPost, "/")

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
	if got := recorder.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want GET, HEAD", got)
	}
}

func TestHeadRequestCarriesNoBody(t *testing.T) {
	handler := HandlerFor(builtAssets())

	recorder := request(t, handler, http.MethodHead, "/")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty for HEAD", recorder.Body)
	}
}

func TestEmbeddedFrontendIsReadable(t *testing.T) {
	if _, err := FS().Open("."); err != nil {
		t.Fatalf("the embedded dist directory is unreadable: %v", err)
	}
	if Built() {
		t.Log("the embedded frontend is built; the notice path is covered by its own test")
	}
}
