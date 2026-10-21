package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDIsGenerated(t *testing.T) {
	var seen string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if seen == "" {
		t.Fatal("the handler saw no request id in its context")
	}
	if got := recorder.Header().Get(RequestIDHeader); got != seen {
		t.Errorf("response header = %q, want the context id %q", got, seen)
	}
}

func TestRequestIDIsReused(t *testing.T) {
	var seen string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.Header.Set(RequestIDHeader, "caller-supplied-id")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if seen != "caller-supplied-id" {
		t.Errorf("context id = %q, want the caller's id", seen)
	}
	if got := recorder.Header().Get(RequestIDHeader); got != "caller-supplied-id" {
		t.Errorf("response header = %q, want the caller's id", got)
	}
}

func TestRequestIDRejectsUnusableCallerID(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"too long":   strings.Repeat("a", 65),
		"has spaces": "not an id",
		"has quotes": `"injected"`,
	}

	for name, supplied := range cases {
		t.Run(name, func(t *testing.T) {
			var seen string
			handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = RequestIDFrom(r.Context())
			}))

			request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			request.Header.Set(RequestIDHeader, supplied)
			handler.ServeHTTP(httptest.NewRecorder(), request)

			if seen == supplied {
				t.Errorf("context id = %q, want a generated id instead", seen)
			}
		})
	}
}

func TestRequestIDFromAnEmptyContextIsEmpty(t *testing.T) {
	if got := RequestIDFrom(t.Context()); got != "" {
		t.Errorf("RequestIDFrom(empty) = %q, want an empty string", got)
	}
}

func TestRequestLogCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := RequestID(Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/thing", nil)
	request.Header.Set(RequestIDHeader, "trace-me")

	handler.ServeHTTP(recorder, request)

	logged := buf.String()
	for _, want := range []string{"trace-me", "POST", "/api/thing", "201"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log = %q, want it to contain %q", logged, want)
		}
	}
}

func TestLoggingRecordsTheStatusOfAnImplicitOK(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("body without an explicit status"))
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), `"status":200`) {
		t.Errorf("log = %q, want status 200", buf.String())
	}
}

func TestPanicIsRecoveredAndLogged(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	calls := 0
	handler := RequestID(Logging(logger)(Recover(logger)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls == 1 {
				panic("the first request explodes")
			}
			w.WriteHeader(http.StatusOK)
		}))))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/thing", nil))

	if first.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", first.Code)
	}
	apiErr := decodeError(t, first.Body.Bytes())
	if apiErr.Code != CodeInternal {
		t.Errorf("code = %q, want %q", apiErr.Code, CodeInternal)
	}
	if strings.Contains(apiErr.Message, "explodes") {
		t.Errorf("message = %q, want it to leak nothing from the panic", apiErr.Message)
	}

	logged := buf.String()
	if !strings.Contains(logged, "explodes") || !strings.Contains(logged, "stack") {
		t.Errorf("log = %q, want the panic and its stack", logged)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/api/thing", nil))

	if second.Code != http.StatusOK {
		t.Errorf("the next request status = %d, want 200", second.Code)
	}
}

func TestHandlerReadsActorFromContext(t *testing.T) {
	var seen Actor
	handler := AnonymousActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = ActorFrom(r.Context())
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if !seen.Anonymous {
		t.Errorf("actor = %+v, want the anonymous actor", seen)
	}
}

func TestActorFromAnEmptyContextIsAnonymous(t *testing.T) {
	if !ActorFrom(t.Context()).Anonymous {
		t.Error("ActorFrom(empty) is not anonymous")
	}
}
