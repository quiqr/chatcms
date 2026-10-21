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
}

func New(opts Options) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()

	mux.Handle("GET /api/health", healthHandler(opts.Health))
	mux.HandleFunc("/api/health", MethodNotAllowed)
	mux.HandleFunc("/api/", NotFound)

	if opts.Assets != nil {
		mux.Handle("/", opts.Assets)
	} else {
		mux.HandleFunc("/", NotFound)
	}

	var handler http.Handler = mux
	handler = AnonymousActor(handler)
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
