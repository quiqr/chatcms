package httpapi

import (
	"context"
	"log/slog"
	"net/http"
)

type Health struct {
	Status        string `json:"status"`
	ServerVersion string `json:"serverVersion"`
	SchemaVersion string `json:"schemaVersion"`
	Database      string `json:"database"`
}

const (
	HealthStatusOK       = "ok"
	HealthStatusDegraded = "degraded"
	DatabaseReachable    = "reachable"
	DatabaseUnreachable  = "unreachable"
)

type HealthFunc func(context.Context) Health

type Options struct {
	Logger *slog.Logger
	Health HealthFunc
	Assets http.Handler
	// Auth answers who someone is. With none, every request carries nobody
	// and the endpoints that need a person answer 401.
	Auth AuthService
	// Secure says the site is served over a secure connection, which decides
	// whether cookies are marked as requiring one.
	Secure bool
}

func New(opts Options) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()

	mux.Handle("GET /api/health", healthHandler(opts.Health))
	mux.HandleFunc("/api/health", MethodNotAllowed)

	if opts.Auth != nil {
		mux.Handle("POST /api/sign-in", signInHandler(opts.Auth, opts.Secure))
		mux.HandleFunc("/api/sign-in", MethodNotAllowed)

		mux.Handle("POST /api/sign-out", signOutHandler(opts.Auth, opts.Secure))
		mux.HandleFunc("/api/sign-out", MethodNotAllowed)

		mux.Handle("GET /api/me", RequireActor(currentPersonHandler()))
		mux.HandleFunc("/api/me", MethodNotAllowed)

		mux.Handle("POST /api/password-reset", requestResetHandler(opts.Auth))
		mux.HandleFunc("/api/password-reset", MethodNotAllowed)

		mux.Handle("POST /api/set-password", setPasswordHandler(opts.Auth))
		mux.HandleFunc("/api/set-password", MethodNotAllowed)
	}

	mux.HandleFunc("/api/", NotFound)

	if opts.Assets != nil {
		mux.Handle("/", opts.Assets)
	} else {
		mux.HandleFunc("/", NotFound)
	}

	var lookup LookupSession
	if opts.Auth != nil {
		lookup = opts.Auth.Lookup
	}

	var handler http.Handler = mux
	handler = RequireRequestToken(opts.Secure)(handler)
	handler = WithSession(lookup, logger)(handler)
	handler = Recover(logger)(handler)
	handler = Logging(logger)(handler)
	handler = RequestID(handler)
	return handler
}

func healthHandler(check HealthFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			WriteError(w, http.StatusServiceUnavailable, CodeUnavailable,
				"Health reporting is not configured.")
			return
		}

		health := check(r.Context())

		status := http.StatusOK
		if health.Database != DatabaseReachable {
			status = http.StatusServiceUnavailable
		}
		WriteJSON(w, status, health)
	})
}
