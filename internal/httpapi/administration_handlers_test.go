package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type stubAdministration struct {
	asked       []string
	err         error
	administers bool
}

func (s *stubAdministration) note(what string) { s.asked = append(s.asked, what) }

func (s *stubAdministration) MakeOrganisation(_ context.Context, _ Actor, slug, name, owner string) (OrganisationView, error) {
	s.note("make " + slug + " " + name + " " + owner)
	return OrganisationView{Slug: slug, Name: name}, s.err
}

func (s *stubAdministration) PutSomebodyInto(_ context.Context, _ Actor, slug, email, role string) error {
	s.note("put " + email + " into " + slug + " as " + role)
	return s.err
}

func (s *stubAdministration) TakeSomebodyOut(_ context.Context, _ Actor, slug, email string) error {
	s.note("take " + email + " out of " + slug)
	return s.err
}

func (s *stubAdministration) Appoint(_ context.Context, _ Actor, email string) error {
	s.note("appoint " + email)
	return s.err
}

func (s *stubAdministration) StandDown(_ context.Context, _ Actor, email string) error {
	s.note("stand down " + email)
	return s.err
}

func (s *stubAdministration) Records(context.Context, Actor, int32) ([]RecordView, error) {
	s.note("records")
	return []RecordView{{Who: "Pim", Did: "made an organisation", To: "Bakery", At: time.Now()}}, s.err
}

func (s *stubAdministration) Administrators(context.Context, Actor) ([]AdministratorView, error) {
	s.note("administrators")
	return []AdministratorView{{Email: "pim@example.org", Yourself: true}}, s.err
}

func (s *stubAdministration) Organisations(context.Context, Actor) ([]OrganisationSummaryView, error) {
	s.note("organisations")
	return []OrganisationSummaryView{{Slug: "bakery", Name: "Bakery", Projects: 2, People: 3}}, s.err
}

func (s *stubAdministration) Organisation(_ context.Context, _ Actor, slug string) (OrganisationWholeView, error) {
	s.note("read " + slug)
	return OrganisationWholeView{
		Slug:     slug,
		Name:     "Bakery",
		People:   []MemberView{{Email: "owner@example.org", Role: "owner"}},
		Projects: []ProjectView{{Slug: "korenaar", Name: "Korenaar Bakery"}},
	}, s.err
}

func (s *stubAdministration) RemoveOrganisation(_ context.Context, _ Actor, slug string) error {
	s.note("remove " + slug)
	return s.err
}

func (s *stubAdministration) TurnAccountOff(_ context.Context, _ Actor, email string) error {
	s.note("turn off " + email)
	return s.err
}

func (s *stubAdministration) TurnAccountOn(_ context.Context, _ Actor, email string) error {
	s.note("turn on " + email)
	return s.err
}

func (s *stubAdministration) MayAdminister(context.Context, Actor) bool {
	s.note("may administer")
	return s.administers
}

func withAdministration(service AdministrationService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "pim@example.org"}
	}

	return New(Options{
		Logger:         slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Auth:           &stubAuth{known: known},
		Administration: service,
	})
}

func TestEveryAdministrationEndpointNeedsSomebody(t *testing.T) {
	// Not one of these may be reached by somebody who is not signed in, and
	// a missing check here would be a door left open on every tenant.
	handler := withAdministration(&stubAdministration{}, false)

	for _, each := range []struct {
		method, path string
		body         any
	}{
		{http.MethodGet, "/api/administration/me", nil},
		{http.MethodGet, "/api/administration/organisations", nil},
		{http.MethodPost, "/api/administration/organisations", map[string]any{"slug": "a"}},
		{http.MethodPost, "/api/administration/organisations/a/people", map[string]any{"email": "a@b.c"}},
		{http.MethodDelete, "/api/administration/organisations/a/people/a@b.c", nil},
		{http.MethodGet, "/api/administration/administrators", nil},
		{http.MethodPost, "/api/administration/administrators", map[string]any{"email": "a@b.c"}},
		{http.MethodDelete, "/api/administration/administrators/a@b.c", nil},
		{http.MethodGet, "/api/administration/records", nil},
		{http.MethodGet, "/api/administration/organisations/bakery", nil},
		{http.MethodDelete, "/api/administration/organisations/bakery", nil},
		{http.MethodPost, "/api/administration/accounts/a@b.c/off", nil},
		{http.MethodPost, "/api/administration/accounts/a@b.c/on", nil},
	} {
		recorder := call(t, handler, each.method, each.path, each.body)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", each.method, each.path, recorder.Code)
		}
	}
}

func TestMakingAnOrganisationPassesOnWhatWasAsked(t *testing.T) {
	service := &stubAdministration{}
	handler := withAdministration(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/administration/organisations",
		map[string]any{"slug": "bakery", "name": "Bakery", "ownerEmail": "marieke@example.org"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body)
	}
	if len(service.asked) == 0 || service.asked[0] != "make bakery Bakery marieke@example.org" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestAdministrationReportsARefusal(t *testing.T) {
	handler := withAdministration(&stubAdministration{err: ErrNotAllowed}, true)

	recorder := signedInCall(t, handler, http.MethodPost, "/api/administration/organisations",
		map[string]any{"slug": "bakery", "name": "Bakery", "ownerEmail": "a@b.c"})

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
}

func TestPuttingSomebodyInAndTakingThemOut(t *testing.T) {
	service := &stubAdministration{}
	handler := withAdministration(service, true)

	put := signedInCall(t, handler, http.MethodPost, "/api/administration/organisations/bakery/people",
		map[string]any{"email": "marieke@example.org", "role": "owner"})
	if put.Code != http.StatusNoContent {
		t.Errorf("putting somebody in = %d, want 204: %s", put.Code, put.Body)
	}

	out := signedInCall(t, handler, http.MethodDelete,
		"/api/administration/organisations/bakery/people/marieke@example.org", nil)
	if out.Code != http.StatusNoContent {
		t.Errorf("taking somebody out = %d, want 204: %s", out.Code, out.Body)
	}

	if len(service.asked) != 2 {
		t.Fatalf("the service was asked %v", service.asked)
	}
	if service.asked[0] != "put marieke@example.org into bakery as owner" {
		t.Errorf("putting in asked %q", service.asked[0])
	}
	if service.asked[1] != "take marieke@example.org out of bakery" {
		t.Errorf("taking out asked %q", service.asked[1])
	}
}

func TestAppointingAndStandingDown(t *testing.T) {
	service := &stubAdministration{}
	handler := withAdministration(service, true)

	appointed := signedInCall(t, handler, http.MethodPost, "/api/administration/administrators",
		map[string]any{"email": "second@example.org"})
	if appointed.Code != http.StatusNoContent {
		t.Errorf("appointing = %d, want 204", appointed.Code)
	}

	stood := signedInCall(t, handler, http.MethodDelete,
		"/api/administration/administrators/second@example.org", nil)
	if stood.Code != http.StatusNoContent {
		t.Errorf("standing down = %d, want 204", stood.Code)
	}
}

func TestStandingDownTheLastOneIsRefusedWithAReason(t *testing.T) {
	handler := withAdministration(&stubAdministration{
		err: &RefusedError{Reason: "this is the last administrator"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodDelete,
		"/api/administration/administrators/only@example.org", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", recorder.Code, recorder.Body)
	}
	if !strings.Contains(recorder.Body.String(), "last administrator") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestWhetherSomebodyMayAdministerIsAnswered(t *testing.T) {
	// The interface asks once rather than offering everything and being
	// refused afterwards.
	for _, administers := range []bool{true, false} {
		handler := withAdministration(&stubAdministration{administers: administers}, true)

		recorder := signedInCall(t, handler, http.MethodGet, "/api/administration/me", nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}

		want := "\"administers\":" + map[bool]string{true: "true", false: "false"}[administers]
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("body = %s, want %s", recorder.Body, want)
		}
	}
}

func TestReadingWhatWasDone(t *testing.T) {
	handler := withAdministration(&stubAdministration{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, "/api/administration/records", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "made an organisation") {
		t.Errorf("body = %s, want the record in it", recorder.Body)
	}
}

func TestAnAccountIsTurnedOffAndOnAgain(t *testing.T) {
	service := &stubAdministration{}
	handler := withAdministration(service, true)

	off := signedInCall(t, handler, http.MethodPost, "/api/administration/accounts/gone@example.org/off", nil)
	if off.Code != http.StatusNoContent {
		t.Errorf("turning off = %d, want 204: %s", off.Code, off.Body)
	}

	on := signedInCall(t, handler, http.MethodPost, "/api/administration/accounts/gone@example.org/on", nil)
	if on.Code != http.StatusNoContent {
		t.Errorf("turning on = %d, want 204: %s", on.Code, on.Body)
	}

	if len(service.asked) != 2 ||
		service.asked[0] != "turn off gone@example.org" ||
		service.asked[1] != "turn on gone@example.org" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestTurningOffYourOwnAccountSaysWhyNot(t *testing.T) {
	handler := withAdministration(&stubAdministration{
		err: &RefusedError{Reason: "you cannot turn off your own account"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodPost,
		"/api/administration/accounts/me@example.org/off", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "your own account") {
		t.Errorf("body = %s, want it to say why", recorder.Body)
	}
}

func TestAnOrganisationIsOpenedWhole(t *testing.T) {
	handler := withAdministration(&stubAdministration{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, "/api/administration/organisations/bakery", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}

	// Its people and its projects come in the one answer, because opening
	// an organisation is one question.
	for _, want := range []string{"owner@example.org", "Korenaar Bakery"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("body = %s, want it to carry %q", recorder.Body, want)
		}
	}
}

func TestAnOrganisationIsRemoved(t *testing.T) {
	service := &stubAdministration{}
	handler := withAdministration(service, true)

	recorder := signedInCall(t, handler, http.MethodDelete,
		"/api/administration/organisations/bakery", nil)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", recorder.Code, recorder.Body)
	}
	if len(service.asked) != 1 || service.asked[0] != "remove bakery" {
		t.Errorf("the service was asked %v", service.asked)
	}
}

func TestRemovingOneThatStillHoldsProjectsSaysWhatIsInTheWay(t *testing.T) {
	handler := withAdministration(&stubAdministration{
		err: &RefusedError{Reason: "this organisation still holds Korenaar Bakery; remove them first"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodDelete,
		"/api/administration/organisations/bakery", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Korenaar Bakery") {
		t.Errorf("body = %s, want it to name what is in the way", recorder.Body)
	}
}

func TestAdministrationRefusesTheWrongMethod(t *testing.T) {
	handler := withAdministration(&stubAdministration{}, true)

	for _, path := range []string{
		"/api/administration/organisations",
		"/api/administration/administrators",
	} {
		recorder := signedInCall(t, handler, http.MethodPut, path, map[string]any{})
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("PUT %s = %d, want 405", path, recorder.Code)
		}
	}
}
