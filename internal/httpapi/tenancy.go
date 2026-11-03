package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// TenancyService is what the HTTP layer needs in order to serve
// organisations, projects and the people on them. The decisions live behind
// it; the status codes and the wording live here.
type TenancyService interface {
	ListOrganisations(ctx context.Context, actor Actor) ([]OrganisationView, error)
	Rename(ctx context.Context, actor Actor, slug, name string) (OrganisationView, error)
	OrganisationPeople(ctx context.Context, actor Actor, slug string) ([]MemberView, error)
	InviteToOrganisation(ctx context.Context, actor Actor, slug, email, role string) (MemberView, error)
	ChangeOrganisationRole(ctx context.Context, actor Actor, slug, personID, role string) error
	RemoveFromOrganisation(ctx context.Context, actor Actor, slug, personID string) error
	SendTheInvitationAgain(ctx context.Context, actor Actor, slug, personID string) error
	ChangeProject(ctx context.Context, actor Actor, organisation, project, name, domain string) (ProjectView, error)
	RemoveProject(ctx context.Context, actor Actor, organisation, project string) error
	CreateOrganisation(ctx context.Context, actor Actor, slug, name string) (OrganisationView, error)
	ListProjects(ctx context.Context, actor Actor, organisationSlug string) ([]ProjectView, error)
	UsageOf(ctx context.Context, actor Actor, organisationSlug string) (UsageView, error)
	ManageableOrganisation(ctx context.Context, actor Actor, organisationSlug string) (OrganisationAreaView, error)
	Duplicate(ctx context.Context, actor Actor, organisation, project, slug, name string, withPeople, withPublishing bool) (ProjectView, error)
	MoveTo(ctx context.Context, actor Actor, organisation, project, destination string) error
	Archive(ctx context.Context, actor Actor, organisation, project string, archived bool) error
	// WriteZip writes the site's files to w, and answers what the file
	// should be called. It takes the writer because a zip of a whole site
	// is not something to hold in memory on the way out.
	WriteZip(ctx context.Context, actor Actor, organisation, project string, w io.Writer) (string, error)
	CreateProject(ctx context.Context, actor Actor, organisationSlug, slug, name, domain string) (ProjectView, error)
	GetProject(ctx context.Context, actor Actor, organisationSlug, projectSlug string) (ProjectView, error)
	ListPeople(ctx context.Context, actor Actor, organisationSlug, projectSlug string) ([]MemberView, error)
	InvitePerson(ctx context.Context, actor Actor, organisationSlug, projectSlug, email, role string) (MemberView, error)
	ChangeRole(ctx context.Context, actor Actor, organisationSlug, projectSlug, personID, role string) (MemberView, error)
	RemovePerson(ctx context.Context, actor Actor, organisationSlug, projectSlug, personID string) error
}

// These are the failures the HTTP layer tells apart.
var (
	// ErrNotAllowed says the person may not do this. It is distinct from
	// ErrMissing, which is also what someone else's project looks like.
	ErrNotAllowed = errors.New("this is not allowed")
	// ErrMissing says the thing does not exist, or belongs to someone the
	// asker is not.
	ErrMissing = errors.New("there is no such thing")
	// ErrRefused says the request was understood and cannot be carried out,
	// with a reason worth showing.
	ErrRefused = errors.New("this cannot be done")
)

// OrganisationView is an organisation as the interface sees it.
type OrganisationView struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	// YouAdminister says this person administers the organisation, which is
	// what decides whether a way into its own screens is offered.
	YouAdminister bool `json:"youAdminister,omitempty"`
}

// OrganisationAreaView is one organisation as whoever administers it sees
// it: what it holds, who is in it, and how full it is. One answer, because
// the shell needs the counts whichever screen of the area is open.
type OrganisationAreaView struct {
	Slug     string               `json:"slug"`
	Name     string               `json:"name"`
	Usage    UsageView            `json:"usage"`
	People   []OrganisationPerson `json:"people"`
	Projects []ProjectView        `json:"projects"`
}

// OrganisationPerson is somebody in an organisation, with the sites inside it
// they hold a role on.
type OrganisationPerson struct {
	PersonID      string          `json:"personId"`
	Email         string          `json:"email"`
	DisplayName   string          `json:"displayName"`
	Role          string          `json:"role"`
	StillToAccept bool            `json:"stillToAccept"`
	LastSeen      *time.Time      `json:"lastSeen,omitempty"`
	Projects      []BelongingView `json:"projects"`
}

// ProjectView is a project as the interface sees it, including what this
// person may do with it, so the interface leaves out what it must rather than
// showing it disabled.
type ProjectView struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Domain      string   `json:"domain"`
	YourRole    string   `json:"yourRole"`
	YouMay      []string `json:"youMay"`
	Unpublished int      `json:"unpublished"`
	// Archived says it is out of the way and not counting against the
	// organisation's limit of sites.
	Archived bool `json:"archived,omitempty"`
}

// MemberView is one person on a project.
type MemberView struct {
	PersonID      string `json:"personId"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Role          string `json:"role"`
	StillToAccept bool   `json:"stillToAccept"`
	// LastSeen is when they were last here, and nothing when they never
	// were.
	LastSeen *time.Time `json:"lastSeen,omitempty"`
}

type changeProjectRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type renameOrganisationRequest struct {
	Name string `json:"name"`
}

type createOrganisationRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type createProjectRequest struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type changeRoleRequest struct {
	Role string `json:"role"`
}

// writeTenancyError turns a failure into the answer the specification asks
// for: 403 when the person may not, and the same answer for something that is
// not theirs as for something that is not there.
func writeTenancyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMissing):
		WriteError(w, http.StatusNotFound, CodeNotFound, "There is nothing here.")
	case errors.Is(err, ErrNotAllowed):
		WriteError(w, http.StatusForbidden, CodeForbidden, "You are not allowed to do that.")
	case errors.Is(err, ErrRefused):
		writeRefusal(w, err)
	default:
		WriteError(w, http.StatusInternalServerError, CodeInternal,
			"Something went wrong. Nothing was lost; please try again.")
	}
}

// writeRefusal says why, in English and by key, so that the screen can say
// it in the language whoever is reading chose.
func writeRefusal(w http.ResponseWriter, err error) {
	said := Error{
		Code:    CodeRefused,
		Message: capitalise(unwrapReason(err)) + ".",
	}

	var refused *RefusedError
	if errors.As(err, &refused) {
		said.Key = refused.Key
		said.Detail = refused.Detail
	}

	WriteJSON(w, http.StatusConflict, errorEnvelope{Error: said})
}

func unwrapReason(err error) string {
	var refused *RefusedError
	if errors.As(err, &refused) && refused.Reason != "" {
		return refused.Reason
	}
	return "this cannot be done"
}

// RefusedError carries why something understood could not be done.
//
// Reason is English, which is what a log and whoever reads one want. Key is
// what the screen says it with, in whatever language the person reading is
// in; a refusal with no key, or one nobody has translated yet, is shown in
// the Reason's own words rather than not at all. Detail is what the sentence
// has to put in: a name, a size, a list.
type RefusedError struct {
	Key    string
	Reason string
	Detail map[string]any
}

func (e *RefusedError) Error() string { return e.Reason }

func (e *RefusedError) Is(target error) bool { return target == ErrRefused }

func listOrganisationsHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organisations, err := service.ListOrganisations(r.Context(), ActorFrom(r.Context()))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"organisations": organisations})
	})
}

func createOrganisationHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createOrganisationRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		organisation, err := service.CreateOrganisation(r.Context(), ActorFrom(r.Context()), body.Slug, body.Name)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, organisation)
	})
}

func listProjectsHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projects, err := service.ListProjects(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		usage, err := service.UsageOf(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"projects": projects, "usage": usage})
	})
}

func createProjectHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body createProjectRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		project, err := service.CreateProject(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), body.Slug, body.Name, body.Domain)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, project)
	})
}

func getProjectHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project, err := service.GetProject(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, project)
	})
}

func listPeopleHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		people, err := service.ListPeople(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"people": people})
	})
}

func invitePersonHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body inviteRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		member, err := service.InvitePerson(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Email, body.Role)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, member)
	})
}

func changeRoleHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body changeRoleRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		member, err := service.ChangeRole(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("person"), body.Role)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, member)
	})
}

func removePersonHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := service.RemovePerson(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("person"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusNoContent, nil)
	})
}

func renameOrganisationHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body renameOrganisationRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		renamed, err := service.Rename(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), body.Name)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, renamed)
	})
}

func organisationPeopleHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		people, err := service.OrganisationPeople(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"people": people})
	})
}

func changeProjectHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body changeProjectRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		changed, err := service.ChangeProject(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Name, body.Domain)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, changed)
	})
}

func removeProjectHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := service.RemoveProject(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project")); err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

type duplicateRequest struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	WithPeople     bool   `json:"withPeople"`
	WithPublishing bool   `json:"withPublishing"`
}

type moveRequest struct {
	To string `json:"to"`
}

type archiveRequest struct {
	Archived bool `json:"archived"`
}

func wholeSiteHandlers(mux *http.ServeMux, service TenancyService) {
	const project = "/api/organisations/{organisation}/projects/{project}"

	mux.Handle("POST "+project+"/copies", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body duplicateRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			made, err := service.Duplicate(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("project"),
				body.Slug, body.Name, body.WithPeople, body.WithPublishing)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusCreated, made)
		})))

	mux.Handle("POST "+project+"/organisation", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body moveRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.MoveTo(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("project"), body.To); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("GET "+project+"/zip", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// The name is asked for before anything is written, so that a
			// refusal is still a refusal rather than half a file.
			var body bytes.Buffer
			name, err := service.WriteZip(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("project"), &body)
			if err != nil {
				writeTenancyError(w, err)
				return
			}

			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.zip"`)
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(body.Bytes()); err != nil {
				slog.Error("writing the zip failed", "error", err)
			}
		})))

	mux.Handle("PUT "+project+"/archived", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body archiveRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}
			if err := service.Archive(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("project"), body.Archived); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))
}

func manageOrganisationHandler(service TenancyService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		area, err := service.ManageableOrganisation(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, area)
	})
}

func organisationPeopleHandlers(mux *http.ServeMux, service TenancyService) {
	const people = "/api/organisations/{organisation}/people"

	mux.Handle("POST "+people, RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body inviteRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}

			invited, err := service.InviteToOrganisation(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), body.Email, body.Role)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusCreated, invited)
		})))

	mux.Handle("PUT "+people+"/{person}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body changeRoleRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}

			if err := service.ChangeOrganisationRole(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("person"), body.Role); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("DELETE "+people+"/{person}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.RemoveFromOrganisation(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("person")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("POST "+people+"/{person}/invitation", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.SendTheInvitationAgain(r.Context(), ActorFrom(r.Context()),
				r.PathValue("organisation"), r.PathValue("person")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc(people+"/{person}", MethodNotAllowed)
}
