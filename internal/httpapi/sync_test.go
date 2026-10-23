package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

// stubSync stands in for how a site stands.
type stubSync struct {
	standing StandingView
	err      error
	answered []AnswerRequest
	calls    []string
}

func (s *stubSync) SiteStanding(context.Context, Actor, string, string) (StandingView, error) {
	s.calls = append(s.calls, "standing")
	return s.standing, s.err
}

func (s *stubSync) BringChanges(context.Context, Actor, string, string) (StandingView, error) {
	s.calls = append(s.calls, "bring")
	return s.standing, s.err
}

func (s *stubSync) AnswerChanges(_ context.Context, _ Actor, _, _ string, answers []AnswerRequest) (StandingView, error) {
	s.calls = append(s.calls, "answer")
	s.answered = answers
	return s.standing, s.err
}

func (s *stubSync) AbandonChanges(context.Context, Actor, string, string) (StandingView, error) {
	s.calls = append(s.calls, "abandon")
	return s.standing, s.err
}

func withSync(service SyncService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Sync:   service,
	})
}

func TestSyncEndpointsNeedAPerson(t *testing.T) {
	handler := withSync(&stubSync{}, false)

	for _, pair := range [][2]string{
		{http.MethodGet, projectBase + "/standing"},
		{http.MethodPost, projectBase + "/standing/bring"},
		{http.MethodPost, projectBase + "/standing/answer"},
		{http.MethodPost, projectBase + "/standing/abandon"},
	} {
		body := any(nil)
		if pair[0] == http.MethodPost {
			body = map[string]any{}
		}
		recorder := call(t, handler, pair[0], pair[1], body)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestHowASiteStandsIsServed(t *testing.T) {
	service := &stubSync{standing: StandingView{
		State: "asking", Said: "Two people changed the same thing", Unpublished: 1,
		ArrivedFrom: []string{"Joris"},
		Conflicts: []PageConflictView{{
			Page: "home", Title: "Home page", File: "content/_index.md", Who: "Joris",
			Fields: []FieldConflictView{{
				Path: "summary", Name: "Summary", Mine: "apples", Theirs: "plums",
				Words: []WordView{{Part: "gone", Text: "apples"}, {Part: "added", Text: "plums"}},
			}},
		}},
	}}
	handler := withSync(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/standing", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var standing StandingView
	if err := json.Unmarshal(recorder.Body.Bytes(), &standing); err != nil {
		t.Fatal(err)
	}
	if standing.State != "asking" || standing.Said == "" {
		t.Errorf("standing = %+v", standing)
	}
	if len(standing.Conflicts) != 1 || len(standing.Conflicts[0].Fields) != 1 {
		t.Fatalf("conflicts = %+v", standing.Conflicts)
	}
	if len(standing.Conflicts[0].Fields[0].Words) != 2 {
		t.Error("the comparison between the two answers did not survive")
	}
}

func TestTechnicalDetailIsOnlyThereWhenTheServiceSentIt(t *testing.T) {
	plain := withSync(&stubSync{standing: StandingView{State: "saved"}}, true)
	recorder := signedInCall(t, plain, http.MethodGet, projectBase+"/standing", nil)
	if strings.Contains(recorder.Body.String(), "technical") {
		t.Errorf("body = %s, want no technical detail", recorder.Body.String())
	}

	detailed := withSync(&stubSync{standing: StandingView{
		State:     "saved",
		Technical: &TechnicalView{Branch: "main", Remote: "git@example.org:a/b.git", Head: "abc123"},
	}}, true)
	recorder = signedInCall(t, detailed, http.MethodGet, projectBase+"/standing", nil)
	if !strings.Contains(recorder.Body.String(), "git@example.org") {
		t.Errorf("body = %s, want the detail", recorder.Body.String())
	}
}

func TestBringingAndAbandoningAreAsked(t *testing.T) {
	for path, want := range map[string]string{
		projectBase + "/standing/bring":   "bring",
		projectBase + "/standing/abandon": "abandon",
	} {
		service := &stubSync{standing: StandingView{State: "saved"}}
		handler := withSync(service, true)

		recorder := signedInCall(t, handler, http.MethodPost, path, nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s = %d, want 200", path, recorder.Code)
		}
		if service.calls[0] != want {
			t.Errorf("calls = %v, want %q", service.calls, want)
		}
	}
}

func TestAnswersArePassedOnAsTheyWereWritten(t *testing.T) {
	service := &stubSync{standing: StandingView{State: "saved"}}
	handler := withSync(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/standing/answer", map[string]any{
		"answers": []map[string]any{
			{"file": "content/_index.md", "fields": map[string]string{"summary": "theirs"}},
			{"file": "static/logo.png", "whole": "mine"},
		},
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	if len(service.answered) != 2 {
		t.Fatalf("answers = %+v", service.answered)
	}
	if service.answered[0].Fields["summary"] != "theirs" {
		t.Errorf("the field choice did not survive: %+v", service.answered[0])
	}
	if service.answered[1].Whole != "mine" {
		t.Errorf("the whole-file choice did not survive: %+v", service.answered[1])
	}
}

func TestAnIncompleteAnswerIsRefusedWithAReason(t *testing.T) {
	handler := withSync(&stubSync{
		err: &RefusedError{Reason: "every page that two people changed needs an answer"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/standing/answer", map[string]any{})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "needs an answer") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestAMalformedAnswerIsRefused(t *testing.T) {
	handler := withSync(&stubSync{}, true)

	recorder := signedInRaw(t, handler, http.MethodPost, projectBase+"/standing/answer", "{not json")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestTheWrongMethodOnStandingIsSaidToBe(t *testing.T) {
	handler := withSync(&stubSync{}, true)

	recorder := signedInCall(t, handler, http.MethodDelete, projectBase+"/standing", nil)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
