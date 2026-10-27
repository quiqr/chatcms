package httpapi

import (
	"context"
	"errors"
	"net/http"
)

// TenancyService is what the HTTP layer needs in order to serve
// organisations, projects and the people on them. The decisions live behind
// it; the status codes and the wording live here.
type TenancyService interface {
	ListOrganisations(ctx context.Context, actor Actor) ([]OrganisationView, error)
	Rename(ctx context.Context, actor Actor, slug, name string) (OrganisationView, error)
	OrganisationPeople(ctx context.Context, actor Actor, slug string) ([]MemberView, error)
	ChangeProject(ctx context.Context, actor Actor, organisation, project, name, domain string) (ProjectView, error)
	RemoveProject(ctx context.Context, actor Actor, organisation, project string) error
	CreateOrganisation(ctx context.Context, actor Actor, slug, name string) (OrganisationView, error)
	ListProjects(ctx context.Context, actor Actor, organisationSlug string) ([]ProjectView, error)
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
}

// MemberView is one person on a project.
type MemberView struct {
	PersonID      string `json:"personId"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Role          string `json:"role"`
	StillToAccept bool   `json:"stillToAccept"`
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
		WriteError(w, http.StatusConflict, CodeRefused, capitalise(unwrapReason(err))+".")
	default:
		WriteError(w, http.StatusInternalServerError, CodeInternal,
			"Something went wrong. Nothing was lost; please try again.")
	}
}

func unwrapReason(err error) string {
	var refused *RefusedError
	if errors.As(err, &refused) && refused.Reason != "" {
		return refused.Reason
	}
	return "this cannot be done"
}

// RefusedError carries why something understood could not be done.
type RefusedError struct {
	Reason string
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
		WriteJSON(w, http.StatusOK, map[string]any{"projects": projects})
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
