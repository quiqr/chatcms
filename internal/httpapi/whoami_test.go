package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubWhoAmI struct {
	view WhoAmIView
	err  error
}

func (s *stubWhoAmI) WhoAmI(context.Context, Actor) (WhoAmIView, error) {
	return s.view, s.err
}

func withWhoAmI(service WhoAmIService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "joris@example.org"}
	}

	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Auth:   &stubAuth{known: known},
		WhoAmI: service,
	})
}

func TestWhoAmICarriesEverythingAScreenNeedsBeforeItDraws(t *testing.T) {
	handler := withWhoAmI(&stubWhoAmI{view: WhoAmIView{
		Person:        PersonView{ID: "person-1", Email: "joris@example.org", DisplayName: "Joris Bakker"},
		Administers:   true,
		Organisations: []OrganisationView{{Slug: "technative", Name: "TechNative"}},
	}}, true)

	recorder := signedInCall(t, handler, http.MethodGet, "/api/whoami", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}

	var answer WhoAmIView
	if err := json.Unmarshal(recorder.Body.Bytes(), &answer); err != nil {
		t.Fatal(err)
	}

	if answer.Person.DisplayName != "Joris Bakker" {
		t.Errorf("name = %q, want the person's own", answer.Person.DisplayName)
	}
	if !answer.Administers {
		t.Error("the answer does not say they administer the installation")
	}
	if len(answer.Organisations) != 1 || answer.Organisations[0].Slug != "technative" {
		t.Errorf("organisations = %v, want the one they may switch to", answer.Organisations)
	}
}

func TestWhoAmINeedsSomebody(t *testing.T) {
	// Asked by nobody, it says to sign in rather than describing nobody.
	recorder := call(t, withWhoAmI(&stubWhoAmI{}, false), http.MethodGet, "/api/whoami", nil)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}

func TestWhoAmIPassesOnAFailure(t *testing.T) {
	handler := withWhoAmI(&stubWhoAmI{err: errors.New("the database fell over")}, true)

	recorder := signedInCall(t, handler, http.MethodGet, "/api/whoami", nil)
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", recorder.Code)
	}
}

func TestWhoAmIRefusesTheWrongMethod(t *testing.T) {
	recorder := signedInCall(t, withWhoAmI(&stubWhoAmI{}, true), http.MethodPost, "/api/whoami", map[string]any{})

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
