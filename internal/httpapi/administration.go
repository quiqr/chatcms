package httpapi

import (
	"context"
	"net/http"
	"time"
)

// AdministrationService is what somebody who administers the installation may
// do. Every method refuses anybody else, which the handlers do not check for
// themselves.
type AdministrationService interface {
	MakeOrganisation(ctx context.Context, who Actor, slug, name, ownerEmail string) (OrganisationView, error)
	PutSomebodyInto(ctx context.Context, who Actor, slug, email, role string) error
	TakeSomebodyOut(ctx context.Context, who Actor, slug, email string) error
	Appoint(ctx context.Context, who Actor, email string) error
	StandDown(ctx context.Context, who Actor, email string) error
	Records(ctx context.Context, who Actor, most int32) ([]RecordView, error)
	Administrators(ctx context.Context, who Actor) ([]AdministratorView, error)
	Organisations(ctx context.Context, who Actor) ([]OrganisationSummaryView, error)
	// MayAdminister answers the one question the interface needs in order to
	// know whether to offer any of this, without making it guess from a
	// refusal.
	MayAdminister(ctx context.Context, who Actor) bool
}

// RecordView is one thing an administrator did.
type RecordView struct {
	Who    string         `json:"who"`
	Did    string         `json:"did"`
	To     string         `json:"to"`
	Detail map[string]any `json:"detail,omitempty"`
	At     time.Time      `json:"at"`
}

// AdministratorView is somebody who administers the installation.
type AdministratorView struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Yourself    bool   `json:"yourself"`
}

// OrganisationSummaryView is one organisation as an administrator sees it,
// which includes ones they are not in.
type OrganisationSummaryView struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Projects int64  `json:"projects"`
	People   int64  `json:"people"`
}

type makeOrganisationRequest struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	OwnerEmail string `json:"ownerEmail"`
}

type organisationPersonRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type administratorRequest struct {
	Email string `json:"email"`
}

func administrationHandlers(mux *http.ServeMux, service AdministrationService) {
	mux.Handle("GET /api/administration/me", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WriteJSON(w, http.StatusOK, map[string]any{
				"administers": service.MayAdminister(r.Context(), ActorFrom(r.Context())),
			})
		})))

	mux.Handle("GET /api/administration/organisations", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			organisations, err := service.Organisations(r.Context(), ActorFrom(r.Context()))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"organisations": organisations})
		})))

	mux.Handle("POST /api/administration/organisations", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body makeOrganisationRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			made, err := service.MakeOrganisation(r.Context(), ActorFrom(r.Context()), body.Slug, body.Name, body.OwnerEmail)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusCreated, made)
		})))

	mux.HandleFunc("/api/administration/organisations", MethodNotAllowed)

	mux.Handle("POST /api/administration/organisations/{organisation}/people", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body organisationPersonRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.PutSomebodyInto(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), body.Email, body.Role); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("DELETE /api/administration/organisations/{organisation}/people/{email}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.TakeSomebodyOut(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET /api/administration/administrators", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			people, err := service.Administrators(r.Context(), ActorFrom(r.Context()))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"administrators": people})
		})))

	mux.Handle("POST /api/administration/administrators", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body administratorRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.Appoint(r.Context(), ActorFrom(r.Context()), body.Email); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc("/api/administration/administrators", MethodNotAllowed)

	mux.Handle("DELETE /api/administration/administrators/{email}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.StandDown(r.Context(), ActorFrom(r.Context()), r.PathValue("email")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET /api/administration/records", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			records, err := service.Records(r.Context(), ActorFrom(r.Context()), 100)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"records": records})
		})))
}
