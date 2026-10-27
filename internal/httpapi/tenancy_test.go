package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubTenancy stands in for the organisations and projects, so the HTTP layer
// can be tested without a database. What it returns is what the handlers have
// to turn into status codes and wording.
type stubTenancy struct {
	organisations []OrganisationView
	projects      []ProjectView
	people        []MemberView
	member        MemberView
	err           error
	calls         []string
}

func (s *stubTenancy) record(name string) { s.calls = append(s.calls, name) }

func (s *stubTenancy) ListOrganisations(context.Context, Actor) ([]OrganisationView, error) {
	s.record("list-organisations")
	return s.organisations, s.err
}

func (s *stubTenancy) ChangeProject(_ context.Context, _ Actor, _, project, name, domain string) (ProjectView, error) {
	s.record("change-project")
	return ProjectView{Slug: project, Name: name, Domain: domain}, s.err
}

func (s *stubTenancy) RemoveProject(context.Context, Actor, string, string) error {
	s.record("remove-project")
	return s.err
}

func (s *stubTenancy) InviteToOrganisation(_ context.Context, _ Actor, _, email, role string) (MemberView, error) {
	s.record("invite-to-organisation")
	return MemberView{Email: email, Role: role, StillToAccept: true}, s.err
}

func (s *stubTenancy) ChangeOrganisationRole(context.Context, Actor, string, string, string) error {
	s.record("change-organisation-role")
	return s.err
}

func (s *stubTenancy) RemoveFromOrganisation(context.Context, Actor, string, string) error {
	s.record("remove-from-organisation")
	return s.err
}

func (s *stubTenancy) SendTheInvitationAgain(context.Context, Actor, string, string) error {
	s.record("send-again")
	return s.err
}

func (s *stubTenancy) Rename(_ context.Context, _ Actor, slug, name string) (OrganisationView, error) {
	s.record("rename")
	return OrganisationView{Slug: slug, Name: name}, s.err
}

func (s *stubTenancy) OrganisationPeople(context.Context, Actor, string) ([]MemberView, error) {
	s.record("organisation-people")
	return nil, s.err
}

func (s *stubTenancy) CreateOrganisation(_ context.Context, _ Actor, slug, name string) (OrganisationView, error) {
	s.record("create-organisation")
	return OrganisationView{Slug: slug, Name: name}, s.err
}

func (s *stubTenancy) ListProjects(context.Context, Actor, string) ([]ProjectView, error) {
	s.record("list-projects")
	return s.projects, s.err
}

func (s *stubTenancy) CreateProject(_ context.Context, _ Actor, _, slug, name, domain string) (ProjectView, error) {
	s.record("create-project")
	return ProjectView{Slug: slug, Name: name, Domain: domain}, s.err
}

func (s *stubTenancy) GetProject(context.Context, Actor, string, string) (ProjectView, error) {
	s.record("get-project")
	if len(s.projects) == 0 {
		return ProjectView{}, s.err
	}
	return s.projects[0], s.err
}

func (s *stubTenancy) ListPeople(context.Context, Actor, string, string) ([]MemberView, error) {
	s.record("list-people")
	return s.people, s.err
}

func (s *stubTenancy) InvitePerson(_ context.Context, _ Actor, _, _, email, role string) (MemberView, error) {
	s.record("invite")
	return MemberView{Email: email, Role: role, StillToAccept: true}, s.err
}

func (s *stubTenancy) ChangeRole(_ context.Context, _ Actor, _, _, personID, role string) (MemberView, error) {
	s.record("change-role")
	return MemberView{PersonID: personID, Role: role}, s.err
}

func (s *stubTenancy) RemovePerson(context.Context, Actor, string, string, string) error {
	s.record("remove")
	return s.err
}

func withTenancy(service TenancyService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "owner@example.org"}
	}

	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:  func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:    &stubAuth{known: known},
		Tenancy: service,
	})
}

func signedInCall(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return call(t, handler, method, path, body, &http.Cookie{Name: SessionCookie, Value: "a-session-value"})
}

func TestTenancyEndpointsNeedAPerson(t *testing.T) {
	handler := withTenancy(&stubTenancy{}, false)

	paths := [][2]string{
		{http.MethodGet, "/api/organisations"},
		{http.MethodPost, "/api/organisations"},
		{http.MethodGet, "/api/organisations/technative/projects"},
		{http.MethodPost, "/api/organisations/technative/projects"},
		{http.MethodGet, "/api/organisations/technative/projects/korenaar"},
		{http.MethodGet, "/api/organisations/technative/projects/korenaar/people"},
		{http.MethodPost, "/api/organisations/technative/projects/korenaar/people"},
		{http.MethodPut, "/api/organisations/technative/projects/korenaar/people/1"},
		{http.MethodDelete, "/api/organisations/technative/projects/korenaar/people/1"},
	}

	for _, pair := range paths {
		recorder := call(t, handler, pair[0], pair[1], map[string]any{})
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestListingOrganisationsAndProjects(t *testing.T) {
	service := &stubTenancy{
		organisations: []OrganisationView{{Slug: "technative", Name: "TechNative"}},
		projects:      []ProjectView{{Slug: "korenaar", Name: "Korenaar", YourRole: "owner"}},
	}
	handler := withTenancy(service, true)

	organisations := signedInCall(t, handler, http.MethodGet, "/api/organisations", nil)
	if organisations.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", organisations.Code)
	}
	if !strings.Contains(organisations.Body.String(), "TechNative") {
		t.Errorf("body = %s", organisations.Body)
	}

	projects := signedInCall(t, handler, http.MethodGet, "/api/organisations/technative/projects", nil)
	if projects.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", projects.Code)
	}

	var payload struct {
		Projects []ProjectView `json:"projects"`
	}
	if err := json.Unmarshal(projects.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Projects) != 1 || payload.Projects[0].YourRole != "owner" {
		t.Errorf("projects = %+v", payload.Projects)
	}
}

func TestCreatingAnOrganisationAndAProject(t *testing.T) {
	service := &stubTenancy{}
	handler := withTenancy(service, true)

	organisation := signedInCall(t, handler, http.MethodPost, "/api/organisations",
		map[string]any{"slug": "technative", "name": "TechNative"})
	if organisation.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", organisation.Code)
	}

	project := signedInCall(t, handler, http.MethodPost, "/api/organisations/technative/projects",
		map[string]any{"slug": "korenaar", "name": "Korenaar", "domain": "korenaar.example"})
	if project.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", project.Code)
	}
	if !strings.Contains(project.Body.String(), "korenaar.example") {
		t.Errorf("body = %s", project.Body)
	}
}

func TestInvitingChangingAndRemoving(t *testing.T) {
	service := &stubTenancy{}
	handler := withTenancy(service, true)

	invited := signedInCall(t, handler, http.MethodPost,
		"/api/organisations/technative/projects/korenaar/people",
		map[string]any{"email": "customer@example.org", "role": "editor"})
	if invited.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", invited.Code)
	}

	var member MemberView
	if err := json.Unmarshal(invited.Body.Bytes(), &member); err != nil {
		t.Fatal(err)
	}
	if member.Email != "customer@example.org" || !member.StillToAccept {
		t.Errorf("member = %+v", member)
	}

	changed := signedInCall(t, handler, http.MethodPut,
		"/api/organisations/technative/projects/korenaar/people/person-2",
		map[string]any{"role": "viewer"})
	if changed.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", changed.Code)
	}

	removed := signedInCall(t, handler, http.MethodDelete,
		"/api/organisations/technative/projects/korenaar/people/person-2", nil)
	if removed.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", removed.Code)
	}
}

func TestTenancyFailuresBecomeTheRightAnswers(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
		code   string
	}{
		"not theirs":     {ErrMissing, http.StatusNotFound, CodeNotFound},
		"not allowed":    {ErrNotAllowed, http.StatusForbidden, CodeForbidden},
		"refused":        {&RefusedError{Reason: "that address is already taken"}, http.StatusConflict, CodeRefused},
		"something else": {context.DeadlineExceeded, http.StatusInternalServerError, CodeInternal},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			handler := withTenancy(&stubTenancy{err: tc.err}, true)
			recorder := signedInCall(t, handler, http.MethodGet,
				"/api/organisations/technative/projects/korenaar", nil)

			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, tc.status, recorder.Body)
			}
			if got := decodeError(t, recorder.Body.Bytes()).Code; got != tc.code {
				t.Errorf("code = %q, want %q", got, tc.code)
			}
		})
	}
}

func TestARefusalSaysWhy(t *testing.T) {
	handler := withTenancy(&stubTenancy{err: &RefusedError{Reason: "that address is already taken"}}, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/organisations",
		map[string]any{"slug": "technative", "name": "TechNative"})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "already taken") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestTenancyEndpointsRejectAMalformedBody(t *testing.T) {
	handler := withTenancy(&stubTenancy{}, true)

	for _, path := range []string{
		"/api/organisations",
		"/api/organisations/technative/projects",
		"/api/organisations/technative/projects/korenaar/people",
	} {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{not json"))
		request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
		request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
		request.Header.Set(RequestTokenHeader, "a-token")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400", path, recorder.Code)
		}
	}
}

func TestTenancyEndpointsRefuseTheWrongMethod(t *testing.T) {
	handler := withTenancy(&stubTenancy{}, true)

	for _, path := range []string{
		"/api/organisations",
		"/api/organisations/technative/projects",
		"/api/organisations/technative/projects/korenaar",
		"/api/organisations/technative/projects/korenaar/people",
	} {
		recorder := signedInCall(t, handler, http.MethodPatch, path, nil)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s with PATCH = %d, want 405", path, recorder.Code)
		}
	}
}

func TestWithoutATenancyServiceTheEndpointsAreNotThere(t *testing.T) {
	handler := New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: map[string]Actor{"a-session-value": {ID: "person-1"}}},
	})

	recorder := signedInCall(t, handler, http.MethodGet, "/api/organisations", nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 when organisations are not served", recorder.Code)
	}
}

func TestRefusedErrorWithNoReason(t *testing.T) {
	handler := withTenancy(&stubTenancy{err: &RefusedError{}}, true)

	recorder := signedInCall(t, handler, http.MethodGet,
		"/api/organisations/technative/projects/korenaar", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if recorder.Body.Len() == 0 {
		t.Error("a refusal with no reason has no message at all")
	}
}

func TestAnOrganisationsPeopleAreManagedThroughTheirOwnEndpoints(t *testing.T) {
	// An organisation's people are not a project's people, and each of these
	// is a door that must ask before it opens.
	service := &stubTenancy{}
	handler := withTenancy(service, true)

	const people = "/api/organisations/bakery/people"

	invited := signedInCall(t, handler, http.MethodPost, people,
		map[string]any{"email": "marieke@example.org", "role": "member"})
	if invited.Code != http.StatusCreated {
		t.Errorf("inviting = %d, want 201: %s", invited.Code, invited.Body)
	}

	changed := signedInCall(t, handler, http.MethodPut, people+"/person-1",
		map[string]any{"role": "owner"})
	if changed.Code != http.StatusNoContent {
		t.Errorf("changing a role = %d, want 204: %s", changed.Code, changed.Body)
	}

	again := signedInCall(t, handler, http.MethodPost, people+"/person-1/invitation", nil)
	if again.Code != http.StatusNoContent {
		t.Errorf("sending again = %d, want 204: %s", again.Code, again.Body)
	}

	removed := signedInCall(t, handler, http.MethodDelete, people+"/person-1", nil)
	if removed.Code != http.StatusNoContent {
		t.Errorf("removing = %d, want 204: %s", removed.Code, removed.Body)
	}

	want := []string{
		"invite-to-organisation",
		"change-organisation-role",
		"send-again",
		"remove-from-organisation",
	}
	if len(service.calls) != len(want) {
		t.Fatalf("the service was asked %v, want %v", service.calls, want)
	}
	for i, each := range want {
		if service.calls[i] != each {
			t.Errorf("asked[%d] = %q, want %q", i, service.calls[i], each)
		}
	}
}

func TestEveryOrganisationPeopleEndpointNeedsSomebody(t *testing.T) {
	handler := withTenancy(&stubTenancy{}, true)

	for _, each := range []struct {
		method, path string
		body         any
	}{
		{http.MethodPost, "/api/organisations/bakery/people", map[string]any{"email": "a@b.c"}},
		{http.MethodPut, "/api/organisations/bakery/people/person-1", map[string]any{"role": "owner"}},
		{http.MethodDelete, "/api/organisations/bakery/people/person-1", nil},
		{http.MethodPost, "/api/organisations/bakery/people/person-1/invitation", nil},
	} {
		recorder := call(t, handler, each.method, each.path, each.body)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", each.method, each.path, recorder.Code)
		}
	}
}

func TestManagingAnOrganisationsPeopleReportsARefusal(t *testing.T) {
	handler := withTenancy(&stubTenancy{err: ErrNotAllowed}, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/organisations/bakery/people",
		map[string]any{"email": "sneaky@example.org", "role": "owner"})

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403: %s", recorder.Code, recorder.Body)
	}
}

func TestTheLastOwnerRefusalReachesTheScreen(t *testing.T) {
	handler := withTenancy(&stubTenancy{
		err: &RefusedError{Reason: "an organisation needs an owner"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodDelete,
		"/api/organisations/bakery/people/person-1", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "needs an owner") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestRenamingAndReadingAnOrganisation(t *testing.T) {
	service := &stubTenancy{}
	handler := withTenancy(service, true)

	renamed := signedInCall(t, handler, http.MethodPut, "/api/organisations/bakery",
		map[string]any{"name": "Bakery BV"})
	if renamed.Code != http.StatusOK {
		t.Errorf("renaming = %d, want 200: %s", renamed.Code, renamed.Body)
	}
	if !strings.Contains(renamed.Body.String(), "Bakery BV") {
		t.Errorf("body = %s, want the new name", renamed.Body)
	}

	listed := signedInCall(t, handler, http.MethodGet, "/api/organisations/bakery/people", nil)
	if listed.Code != http.StatusOK {
		t.Errorf("reading the people = %d, want 200", listed.Code)
	}
}

func TestChangingAndRemovingAProjectThroughItsEndpoints(t *testing.T) {
	service := &stubTenancy{}
	handler := withTenancy(service, true)

	const project = "/api/organisations/bakery/projects/korenaar"

	changed := signedInCall(t, handler, http.MethodPut, project,
		map[string]any{"name": "Korenaar BV", "domain": "korenaar.nl"})
	if changed.Code != http.StatusOK {
		t.Errorf("changing = %d, want 200: %s", changed.Code, changed.Body)
	}

	removed := signedInCall(t, handler, http.MethodDelete, project, nil)
	if removed.Code != http.StatusNoContent {
		t.Errorf("removing = %d, want 204: %s", removed.Code, removed.Body)
	}
}
