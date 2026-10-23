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
	// Tenancy serves organisations, projects and the people on them.
	Tenancy TenancyService
	// Onboarding serves setting a project up: its key, its site, and where
	// it publishes.
	Onboarding OnboardingService
	// Publish serves the review of what changed and the act of publishing.
	Publish PublishService
	// History serves what was published when, and going back to a version.
	History HistoryService
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

	if opts.Tenancy != nil {
		mux.Handle("GET /api/organisations", RequireActor(listOrganisationsHandler(opts.Tenancy)))
		mux.Handle("POST /api/organisations", RequireActor(createOrganisationHandler(opts.Tenancy)))
		mux.HandleFunc("/api/organisations", MethodNotAllowed)

		mux.Handle("GET /api/organisations/{organisation}/projects",
			RequireActor(listProjectsHandler(opts.Tenancy)))
		mux.Handle("POST /api/organisations/{organisation}/projects",
			RequireActor(createProjectHandler(opts.Tenancy)))
		mux.HandleFunc("/api/organisations/{organisation}/projects", MethodNotAllowed)

		mux.Handle("GET /api/organisations/{organisation}/projects/{project}",
			RequireActor(getProjectHandler(opts.Tenancy)))
		mux.HandleFunc("/api/organisations/{organisation}/projects/{project}", MethodNotAllowed)

		mux.Handle("GET /api/organisations/{organisation}/projects/{project}/people",
			RequireActor(listPeopleHandler(opts.Tenancy)))
		mux.Handle("POST /api/organisations/{organisation}/projects/{project}/people",
			RequireActor(invitePersonHandler(opts.Tenancy)))
		mux.HandleFunc("/api/organisations/{organisation}/projects/{project}/people", MethodNotAllowed)

		mux.Handle("PUT /api/organisations/{organisation}/projects/{project}/people/{person}",
			RequireActor(changeRoleHandler(opts.Tenancy)))
		mux.Handle("DELETE /api/organisations/{organisation}/projects/{project}/people/{person}",
			RequireActor(removePersonHandler(opts.Tenancy)))
		mux.HandleFunc("/api/organisations/{organisation}/projects/{project}/people/{person}", MethodNotAllowed)
	}

	if opts.Onboarding != nil {
		base := "/api/organisations/{organisation}/projects/{project}"

		mux.Handle("GET "+base+"/key", RequireActor(readKeyHandler(opts.Onboarding)))
		mux.Handle("POST "+base+"/key", RequireActor(generateKeyHandler(opts.Onboarding)))
		mux.HandleFunc(base+"/key", MethodNotAllowed)

		mux.Handle("POST "+base+"/import", RequireActor(importSiteHandler(opts.Onboarding)))
		mux.HandleFunc(base+"/import", MethodNotAllowed)

		mux.Handle("GET "+base+"/site", RequireActor(inspectSiteHandler(opts.Onboarding)))
		mux.HandleFunc(base+"/site", MethodNotAllowed)

		mux.Handle("POST "+base+"/connection", RequireActor(testConnectionHandler(opts.Onboarding)))
		mux.HandleFunc(base+"/connection", MethodNotAllowed)

		mux.Handle("GET "+base+"/publishing", RequireActor(readPublishingHandler(opts.Onboarding)))
		mux.Handle("PUT "+base+"/publishing", RequireActor(setPublishingHandler(opts.Onboarding)))
		mux.HandleFunc(base+"/publishing", MethodNotAllowed)
	}

	if opts.Publish != nil {
		base := "/api/organisations/{organisation}/projects/{project}"

		mux.Handle("GET "+base+"/review", RequireActor(reviewHandler(opts.Publish)))
		mux.HandleFunc(base+"/review", MethodNotAllowed)

		mux.Handle("GET "+base+"/publishes", RequireActor(publishHistoryHandler(opts.Publish)))
		mux.Handle("POST "+base+"/publishes", RequireActor(publishHandler(opts.Publish)))
		mux.HandleFunc(base+"/publishes", MethodNotAllowed)

		mux.Handle("GET "+base+"/publishes/{publish}", RequireActor(readPublishHandler(opts.Publish)))
		mux.HandleFunc(base+"/publishes/{publish}", MethodNotAllowed)
	}

	if opts.History != nil {
		base := "/api/organisations/{organisation}/projects/{project}"

		mux.Handle("GET "+base+"/versions", RequireActor(timelineHandler(opts.History)))
		mux.HandleFunc(base+"/versions", MethodNotAllowed)

		mux.Handle("GET "+base+"/versions/{version}", RequireActor(versionContentsHandler(opts.History)))
		mux.HandleFunc(base+"/versions/{version}", MethodNotAllowed)

		mux.Handle("GET "+base+"/versions/{version}/go-back", RequireActor(beforeGoingBackHandler(opts.History)))
		mux.Handle("POST "+base+"/versions/{version}/go-back", RequireActor(goBackHandler(opts.History)))
		mux.HandleFunc(base+"/versions/{version}/go-back", MethodNotAllowed)
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
