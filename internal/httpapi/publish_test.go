package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubPublish stands in for the publish service, so the HTTP layer can be
// tested without git, a sandbox or a database.
type stubPublish struct {
	review  ReviewView
	record  PublishView
	history []PublishView
	err     error
	calls   []string
}

func (s *stubPublish) Review(context.Context, Actor, string, string) (ReviewView, error) {
	s.calls = append(s.calls, "review")
	return s.review, s.err
}

func (s *stubPublish) Publish(_ context.Context, _ Actor, _, _, note string) (PublishView, error) {
	s.calls = append(s.calls, "publish:"+note)
	return s.record, s.err
}

func (s *stubPublish) ReadPublish(_ context.Context, _ Actor, _, _, id string) (PublishView, error) {
	s.calls = append(s.calls, "read:"+id)
	return s.record, s.err
}

func (s *stubPublish) PublishHistory(context.Context, Actor, string, string) ([]PublishView, error) {
	s.calls = append(s.calls, "history")
	return s.history, s.err
}

// stubHistory stands in for the history service.
type stubHistory struct {
	timeline TimelineView
	contents ReviewView
	warning  WarningView
	wentBack WentBackView
	err      error
	calls    []string
}

func (s *stubHistory) Timeline(context.Context, Actor, string, string) (TimelineView, error) {
	s.calls = append(s.calls, "timeline")
	return s.timeline, s.err
}

func (s *stubHistory) VersionContents(_ context.Context, _ Actor, _, _, version string) (ReviewView, error) {
	s.calls = append(s.calls, "contents:"+version)
	return s.contents, s.err
}

func (s *stubHistory) BeforeGoingBack(_ context.Context, _ Actor, _, _, version string) (WarningView, error) {
	s.calls = append(s.calls, "before:"+version)
	return s.warning, s.err
}

func (s *stubHistory) GoBackTo(_ context.Context, _ Actor, _, _, version string, publishRightAway bool) (WentBackView, error) {
	s.calls = append(s.calls, "go-back:"+version+":"+boolWord(publishRightAway))
	return s.wentBack, s.err
}

func boolWord(yes bool) string {
	if yes {
		return "publishing"
	}
	return "waiting"
}

func withPublishing(publishing PublishService, timeline HistoryService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:  func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:    &stubAuth{known: known},
		Publish: publishing,
		History: timeline,
	})
}

func TestPublishEndpointsNeedAPerson(t *testing.T) {
	handler := withPublishing(&stubPublish{}, &stubHistory{}, false)

	for _, pair := range [][2]string{
		{http.MethodGet, projectBase + "/review"},
		{http.MethodGet, projectBase + "/publishes"},
		{http.MethodPost, projectBase + "/publishes"},
		{http.MethodGet, projectBase + "/publishes/a-publish"},
		{http.MethodGet, projectBase + "/versions"},
		{http.MethodGet, projectBase + "/versions/aaa"},
		{http.MethodGet, projectBase + "/versions/aaa/go-back"},
		{http.MethodPost, projectBase + "/versions/aaa/go-back"},
	} {
		recorder := call(t, handler, pair[0], pair[1], map[string]any{})
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestTheReviewIsServed(t *testing.T) {
	service := &stubPublish{review: ReviewView{
		Count:         1,
		LastPublished: "2026-03-01T09:00:00Z",
		Changes: []ChangeView{{
			Kind: "changed-page", Section: "Pages", Page: "Home page", By: "Anna",
			Fields: []FieldChangeView{{Name: "Introduction", Words: []WordView{
				{Part: "kept", Text: "We sell "},
				{Part: "gone", Text: "apples"},
				{Part: "added", Text: "plums"},
			}}},
		}},
	}}
	handler := withPublishing(service, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/review", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var review ReviewView
	if err := json.Unmarshal(recorder.Body.Bytes(), &review); err != nil {
		t.Fatal(err)
	}
	if review.Count != 1 || len(review.Changes) != 1 {
		t.Fatalf("review = %+v", review)
	}
	if len(review.Changes[0].Fields[0].Words) != 3 {
		t.Errorf("the word comparison did not survive: %+v", review.Changes[0].Fields[0])
	}
}

func TestPublishingCarriesTheNote(t *testing.T) {
	service := &stubPublish{record: PublishView{ID: "a-publish", State: "succeeded", ChangeCount: 2}}
	handler := withPublishing(service, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/publishes",
		map[string]any{"note": "Spring prices"})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}

	if len(service.calls) != 1 || service.calls[0] != "publish:Spring prices" {
		t.Errorf("calls = %v, want the note passed on", service.calls)
	}
}

func TestPublishingWithNoNoteIsFine(t *testing.T) {
	service := &stubPublish{record: PublishView{ID: "a-publish", State: "succeeded"}}
	handler := withPublishing(service, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/publishes", map[string]any{})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if service.calls[0] != "publish:" {
		t.Errorf("calls = %v", service.calls)
	}
}

func TestAPublishAndItsHistoryAreServed(t *testing.T) {
	service := &stubPublish{
		record:  PublishView{ID: "a-publish", State: "failed", Trouble: "We couldn't publish this time."},
		history: []PublishView{{ID: "a-publish", State: "succeeded", Note: "Spring prices"}},
	}
	handler := withPublishing(service, &stubHistory{}, true)

	one := signedInCall(t, handler, http.MethodGet, projectBase+"/publishes/a-publish", nil)
	if one.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", one.Code)
	}
	if !strings.Contains(one.Body.String(), "We couldn't publish") {
		t.Errorf("body = %s", one.Body.String())
	}

	all := signedInCall(t, handler, http.MethodGet, projectBase+"/publishes", nil)
	if all.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", all.Code)
	}

	var envelope struct {
		Publishes []PublishView `json:"publishes"`
	}
	if err := json.Unmarshal(all.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Publishes) != 1 || envelope.Publishes[0].Note != "Spring prices" {
		t.Errorf("publishes = %+v", envelope.Publishes)
	}
}

func TestAPublishThatIsRefusedSaysWhy(t *testing.T) {
	service := &stubPublish{err: &RefusedError{Reason: "this site is already being published"}}
	handler := withPublishing(service, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/publishes", map[string]any{})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "already being published") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestAPublishNobodyMaySeeIsNotFound(t *testing.T) {
	handler := withPublishing(&stubPublish{err: ErrMissing}, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/publishes/a-publish", nil)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestAMalformedPublishRequestIsRefused(t *testing.T) {
	handler := withPublishing(&stubPublish{}, &stubHistory{}, true)

	recorder := signedInRaw(t, handler, http.MethodPost, projectBase+"/publishes", "{not json")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestTheTimelineIsServed(t *testing.T) {
	service := &stubHistory{timeline: TimelineView{
		Unpublished: 2,
		Groups: []GroupView{{
			Period: "Today",
			Entries: []VersionView{{
				Version: "ccc", Title: "Spring prices", By: "Anna", Changes: 3, Live: true,
			}},
		}},
	}}
	handler := withPublishing(&stubPublish{}, service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/versions", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var timeline TimelineView
	if err := json.Unmarshal(recorder.Body.Bytes(), &timeline); err != nil {
		t.Fatal(err)
	}
	if timeline.Unpublished != 2 || len(timeline.Groups) != 1 {
		t.Fatalf("timeline = %+v", timeline)
	}
	if !timeline.Groups[0].Entries[0].Live {
		t.Error("the live version is not marked in the answer")
	}
}

func TestWhatAVersionContainsIsServed(t *testing.T) {
	service := &stubHistory{contents: ReviewView{
		Count:   1,
		Changes: []ChangeView{{Kind: "changed-page", Page: "Home page"}},
	}}
	handler := withPublishing(&stubPublish{}, service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/versions/aaa", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if service.calls[0] != "contents:aaa" {
		t.Errorf("calls = %v, want the version passed on", service.calls)
	}
}

func TestTheConfirmationBeforeGoingBackIsServed(t *testing.T) {
	service := &stubHistory{warning: WarningView{
		Title:       "The first update",
		SetAside:    2,
		Reassurance: "2 changes you have not published yet will be set aside. Nothing is lost.",
	}}
	handler := withPublishing(&stubPublish{}, service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/versions/aaa/go-back", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Nothing is lost") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestGoingBackPublishesRightAwayUnlessTurnedOff(t *testing.T) {
	cases := map[string]struct {
		body map[string]any
		want string
	}{
		"nothing said": {map[string]any{}, "go-back:aaa:publishing"},
		"asked for":    {map[string]any{"publishRightAway": true}, "go-back:aaa:publishing"},
		"turned off":   {map[string]any{"publishRightAway": false}, "go-back:aaa:waiting"},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			service := &stubHistory{wentBack: WentBackView{Title: "The first update"}}
			handler := withPublishing(&stubPublish{}, service, true)

			recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/versions/aaa/go-back", test.body)
			if recorder.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201", recorder.Code)
			}
			if service.calls[0] != test.want {
				t.Errorf("calls = %v, want %q", service.calls, test.want)
			}
		})
	}
}

func TestGoingBackToWhereTheSiteIsIsRefused(t *testing.T) {
	service := &stubHistory{err: &RefusedError{Reason: "your site is already exactly as that version left it"}}
	handler := withPublishing(&stubPublish{}, service, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/versions/aaa/go-back", map[string]any{})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "already") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestGoingBackWhenNotAllowedIsRefused(t *testing.T) {
	handler := withPublishing(&stubPublish{}, &stubHistory{err: ErrNotAllowed}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/versions/aaa/go-back", map[string]any{})

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
}

func TestAMalformedGoBackRequestIsRefused(t *testing.T) {
	handler := withPublishing(&stubPublish{}, &stubHistory{}, true)

	recorder := signedInRaw(t, handler, http.MethodPost, projectBase+"/versions/aaa/go-back", "{not json")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestSomethingUnexpectedIsNotLeakedToTheCaller(t *testing.T) {
	handler := withPublishing(&stubPublish{err: errors.New("the database is on fire")}, &stubHistory{}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/review", nil)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "on fire") {
		t.Errorf("body = %s, want nothing of what went wrong", recorder.Body.String())
	}
}

func TestTheWrongMethodIsSaidToBeTheWrongMethod(t *testing.T) {
	handler := withPublishing(&stubPublish{}, &stubHistory{}, true)

	for _, pair := range [][2]string{
		{http.MethodDelete, projectBase + "/review"},
		{http.MethodDelete, projectBase + "/publishes"},
		{http.MethodDelete, projectBase + "/versions"},
		{http.MethodDelete, projectBase + "/versions/aaa/go-back"},
	} {
		recorder := signedInCall(t, handler, pair[0], pair[1], nil)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s = %d, want 405", pair[0], pair[1], recorder.Code)
		}
	}
}

// signedInRaw sends a body this package did not encode, which is how a
// malformed request arrives in practice.
func signedInRaw(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	request.AddCookie(&http.Cookie{Name: RequestTokenCookie, Value: "a-token"})
	request.Header.Set(RequestTokenHeader, "a-token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
