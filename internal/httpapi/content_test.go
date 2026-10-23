package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"io"
	"log/slog"
)

// stubContent stands in for reading and writing one page, so the HTTP layer
// can be tested without a model, a working copy or git.
type stubContent struct {
	page  PageView
	err   error
	calls []string
	edits []EditRequest
}

func (s *stubContent) ReadPage(_ context.Context, _ Actor, _, _, page string) (PageView, error) {
	s.calls = append(s.calls, "read:"+page)
	return s.page, s.err
}

func (s *stubContent) SavePage(_ context.Context, _ Actor, _, _, page string, edit EditRequest) (PageView, error) {
	s.calls = append(s.calls, "save:"+page)
	s.edits = append(s.edits, edit)
	return s.page, s.err
}

func withContent(service ContentService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:  func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:    &stubAuth{known: known},
		Content: service,
	})
}

func aPageView() PageView {
	return PageView{
		Key: "home", Title: "Home page", Section: "Pages", HasBody: true,
		Body: "The body.\n",
		Fields: []ControlView{
			{Path: "title", Type: "string", Title: "Page title", Value: "Welcome"},
			{Path: "seo", Type: "section", Title: "Search", Fields: []ControlView{
				{Path: "seo.description", Type: "string", Title: "Description"},
			}},
		},
	}
}

func TestContentEndpointsNeedAPerson(t *testing.T) {
	handler := withContent(&stubContent{}, false)

	for _, pair := range [][2]string{
		{http.MethodGet, projectBase + "/pages/home"},
		{http.MethodPut, projectBase + "/pages/home"},
	} {
		recorder := call(t, handler, pair[0], pair[1], map[string]any{})
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestAPageIsServedAsAForm(t *testing.T) {
	service := &stubContent{page: aPageView()}
	handler := withContent(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/pages/home", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var page PageView
	if err := json.Unmarshal(recorder.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Title != "Home page" || len(page.Fields) != 2 {
		t.Fatalf("page = %+v", page)
	}
	if len(page.Fields[1].Fields) != 1 {
		t.Error("the fields inside a group did not survive")
	}
	if service.calls[0] != "read:home" {
		t.Errorf("calls = %v", service.calls)
	}
}

func TestAPageInACollectionIsAddressedByItsWholeName(t *testing.T) {
	service := &stubContent{page: aPageView()}
	handler := withContent(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/pages/news/spring-prices", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if service.calls[0] != "read:news/spring-prices" {
		t.Errorf("calls = %v, want the collection and the item", service.calls)
	}
}

func TestAnEditIsPassedOnAsItWasWritten(t *testing.T) {
	service := &stubContent{page: aPageView()}
	handler := withContent(service, true)

	recorder := signedInCall(t, handler, http.MethodPut, projectBase+"/pages/home", map[string]any{
		"values": map[string]any{"title": "Welcome home", "published": true, "tags": []string{"a", "b"}},
		"body":   "A new body.\n",
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	edit := service.edits[0]
	if edit.Values["title"] != "Welcome home" || edit.Values["published"] != true {
		t.Errorf("values = %+v", edit.Values)
	}
	if edit.Body == nil || *edit.Body != "A new body.\n" {
		t.Errorf("body = %v", edit.Body)
	}
}

func TestAnEditThatSaysNothingAboutTheBodyLeavesItAlone(t *testing.T) {
	service := &stubContent{page: aPageView()}
	handler := withContent(service, true)

	signedInCall(t, handler, http.MethodPut, projectBase+"/pages/home", map[string]any{
		"values": map[string]any{"title": "Welcome home"},
	})

	if service.edits[0].Body != nil {
		t.Error("an edit that said nothing about the text carried a text to write")
	}
}

func TestClearingAFieldIsSaidWithNothing(t *testing.T) {
	service := &stubContent{page: aPageView()}
	handler := withContent(service, true)

	signedInRaw(t, handler, http.MethodPut, projectBase+"/pages/home",
		`{"values":{"colour":null}}`)

	value, named := service.edits[0].Values["colour"]
	if !named {
		t.Fatal("the cleared field did not reach the service")
	}
	if value != nil {
		t.Errorf("value = %v, want nothing", value)
	}
}

func TestAPageThatIsNotThereIsNotFound(t *testing.T) {
	handler := withContent(&stubContent{err: ErrMissing}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/pages/nothing", nil)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestAnEditNobodyMayMakeIsRefused(t *testing.T) {
	handler := withContent(&stubContent{err: ErrNotAllowed}, true)

	recorder := signedInCall(t, handler, http.MethodPut, projectBase+"/pages/home", map[string]any{})

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
}

func TestASiteWithNoModelSaysSoWithAReason(t *testing.T) {
	handler := withContent(&stubContent{
		err: &RefusedError{Reason: "this site has no Quiqr model yet, so there are no forms to fill in"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/pages/home", nil)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "no Quiqr model") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestAMalformedEditIsRefused(t *testing.T) {
	handler := withContent(&stubContent{}, true)

	recorder := signedInRaw(t, handler, http.MethodPut, projectBase+"/pages/home", "{not json")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", recorder.Code)
	}
}

func TestTheWrongMethodOnAPageIsSaidToBe(t *testing.T) {
	handler := withContent(&stubContent{}, true)

	recorder := signedInCall(t, handler, http.MethodDelete, projectBase+"/pages/home", nil)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
