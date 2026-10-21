package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func healthyRouter(health Health) http.Handler {
	return New(Options{
		Health: func(context.Context) Health { return health },
		Assets: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("the application document"))
		}),
	})
}

func get(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder
}

func TestHealthOK(t *testing.T) {
	handler := healthyRouter(Health{
		Status:        HealthStatusOK,
		ServerVersion: "1.2.3",
		SchemaVersion: "7",
		Database:      DatabaseReachable,
	})

	recorder := get(t, handler, http.MethodGet, "/api/health")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var health Health
	if err := json.Unmarshal(recorder.Body.Bytes(), &health); err != nil {
		t.Fatalf("body %q is not a health report: %v", recorder.Body, err)
	}
	if health.ServerVersion != "1.2.3" || health.SchemaVersion != "7" {
		t.Errorf("health = %+v, want the version and schema reported", health)
	}
}

func TestHealthReportsUnreachableDatabase(t *testing.T) {
	handler := healthyRouter(Health{
		Status:        HealthStatusDegraded,
		ServerVersion: "1.2.3",
		SchemaVersion: "7",
		Database:      DatabaseUnreachable,
	})

	recorder := get(t, handler, http.MethodGet, "/api/health")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), DatabaseUnreachable) {
		t.Errorf("body = %q, want it to say the database is unreachable", recorder.Body)
	}
	if strings.Contains(recorder.Body.String(), "postgres://") {
		t.Errorf("body = %q, want no connection string", recorder.Body)
	}
}

func TestHealthWithoutACheckIsUnavailable(t *testing.T) {
	recorder := get(t, New(Options{}), http.MethodGet, "/api/health")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", recorder.Code)
	}
}

func TestUnknownAPIPathReturnsJSON404(t *testing.T) {
	handler := healthyRouter(Health{Database: DatabaseReachable})

	recorder := get(t, handler, http.MethodGet, "/api/nothing-here")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	apiErr := decodeError(t, recorder.Body.Bytes())
	if apiErr.Code != CodeNotFound {
		t.Errorf("code = %q, want %q", apiErr.Code, CodeNotFound)
	}
}

func TestUnknownAPIPathIsNotTheApplicationDocument(t *testing.T) {
	handler := healthyRouter(Health{Database: DatabaseReachable})

	recorder := get(t, handler, http.MethodGet, "/api/nothing-here")

	if strings.Contains(recorder.Body.String(), "the application document") {
		t.Errorf("body = %q, want the JSON error instead of the document", recorder.Body)
	}
}

func TestWrongMethodOnAnAPIPathReturnsJSON405(t *testing.T) {
	handler := healthyRouter(Health{Database: DatabaseReachable})

	recorder := get(t, handler, http.MethodDelete, "/api/health")

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	apiErr := decodeError(t, recorder.Body.Bytes())
	if apiErr.Code != CodeMethodNotAllowed {
		t.Errorf("code = %q, want %q", apiErr.Code, CodeMethodNotAllowed)
	}
}

func TestNonAPIPathReachesTheAssets(t *testing.T) {
	handler := healthyRouter(Health{Database: DatabaseReachable})

	recorder := get(t, handler, http.MethodGet, "/projects/korenaar/content")

	if recorder.Body.String() != "the application document" {
		t.Errorf("body = %q, want the assets handler to have served it", recorder.Body)
	}
}

func TestWithoutAssetsEveryPathIsJSON404(t *testing.T) {
	recorder := get(t, New(Options{}), http.MethodGet, "/anything")

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestEveryResponseCarriesARequestID(t *testing.T) {
	handler := healthyRouter(Health{Database: DatabaseReachable})

	for _, path := range []string{"/api/health", "/api/missing", "/projects"} {
		recorder := get(t, handler, http.MethodGet, path)
		if recorder.Header().Get(RequestIDHeader) == "" {
			t.Errorf("%s carried no request id", path)
		}
	}
}
