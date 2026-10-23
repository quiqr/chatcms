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

// stubBrowse stands in for what a site holds and what can be done to it.
type stubBrowse struct {
	sections []SectionView
	page     PageView
	err      error
	calls    []string
}

func (s *stubBrowse) ListContent(context.Context, Actor, string, string) ([]SectionView, error) {
	s.calls = append(s.calls, "list")
	return s.sections, s.err
}

func (s *stubBrowse) CreatePage(_ context.Context, _ Actor, _, _, section, title string) (PageView, error) {
	s.calls = append(s.calls, "create:"+section+":"+title)
	return s.page, s.err
}

func (s *stubBrowse) RenamePage(_ context.Context, _ Actor, _, _, page, title string) (PageView, error) {
	s.calls = append(s.calls, "rename:"+page+":"+title)
	return s.page, s.err
}

func (s *stubBrowse) DuplicatePage(_ context.Context, _ Actor, _, _, page, title string) (PageView, error) {
	s.calls = append(s.calls, "copy:"+page+":"+title)
	return s.page, s.err
}

func (s *stubBrowse) DeletePage(_ context.Context, _ Actor, _, _, page string) error {
	s.calls = append(s.calls, "delete:"+page)
	return s.err
}

func withBrowse(service BrowseService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}

	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Browse: service,
	})
}

func TestBrowseEndpointsNeedAPerson(t *testing.T) {
	handler := withBrowse(&stubBrowse{}, false)

	for _, pair := range [][3]string{
		{http.MethodGet, projectBase + "/content", ""},
		{http.MethodPost, projectBase + "/content", "body"},
		{http.MethodPost, projectBase + "/rename/news/one", "body"},
		{http.MethodDelete, projectBase + "/pages/news/one", ""},
	} {
		var body any
		if pair[2] != "" {
			body = map[string]any{}
		}
		recorder := call(t, handler, pair[0], pair[1], body)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", pair[0], pair[1], recorder.Code)
		}
	}
}

func TestWhatASiteHoldsIsServed(t *testing.T) {
	service := &stubBrowse{sections: []SectionView{
		{Key: "pages", Title: "Pages", Count: 1, Items: []ItemView{{Key: "home", Title: "Home page"}}},
		{Key: "news", Title: "News", ItemName: "article", Count: 1,
			Items: []ItemView{{Key: "news/one", Title: "One", Waiting: true, By: "Anna"}}},
	}}
	handler := withBrowse(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, projectBase+"/content", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var envelope struct {
		Sections []SectionView `json:"sections"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Sections) != 2 {
		t.Fatalf("sections = %+v", envelope.Sections)
	}
	if envelope.Sections[1].ItemName != "article" {
		t.Errorf("a collection's item name did not survive: %+v", envelope.Sections[1])
	}
	if !envelope.Sections[1].Items[0].Waiting {
		t.Error("the mark on an unpublished page did not survive")
	}
}

func TestMakingAPage(t *testing.T) {
	service := &stubBrowse{page: PageView{Key: "news/spring-prices"}}
	handler := withBrowse(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/content",
		map[string]any{"section": "news", "title": "Spring prices"})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if service.calls[0] != "create:news:Spring prices" {
		t.Errorf("calls = %v", service.calls)
	}
}

func TestRenamingAndCopyingUseTheSameEndpoint(t *testing.T) {
	cases := map[string]struct {
		body map[string]any
		want string
	}{
		"rename": {map[string]any{"title": "The first"}, "rename:news/one:The first"},
		"copy":   {map[string]any{"title": "One again", "copy": true}, "copy:news/one:One again"},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			service := &stubBrowse{page: PageView{Key: "news/x"}}
			handler := withBrowse(service, true)

			recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/rename/news/one", test.body)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if service.calls[0] != test.want {
				t.Errorf("calls = %v, want %q", service.calls, test.want)
			}
		})
	}
}

func TestRemovingAPage(t *testing.T) {
	service := &stubBrowse{}
	handler := withBrowse(service, true)

	recorder := signedInCall(t, handler, http.MethodDelete, projectBase+"/pages/news/one", nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if service.calls[0] != "delete:news/one" {
		t.Errorf("calls = %v", service.calls)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want nothing", recorder.Body.String())
	}
}

func TestSomethingRefusedSaysWhy(t *testing.T) {
	handler := withBrowse(&stubBrowse{
		err: &RefusedError{Reason: "there is already a page called \"One\" here"},
	}, true)

	recorder := signedInCall(t, handler, http.MethodPost, projectBase+"/content",
		map[string]any{"section": "news", "title": "One"})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "already a page") {
		t.Errorf("body = %s", recorder.Body.String())
	}
}

func TestAMalformedRequestToMakeAPageIsRefused(t *testing.T) {
	handler := withBrowse(&stubBrowse{}, true)

	for _, path := range []string{projectBase + "/content", projectBase + "/rename/news/one"} {
		recorder := signedInRaw(t, handler, http.MethodPost, path, "{not json")
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400", path, recorder.Code)
		}
	}
}

func TestTheWrongMethodOnTheContentIsSaidToBe(t *testing.T) {
	handler := withBrowse(&stubBrowse{}, true)

	recorder := signedInCall(t, handler, http.MethodPut, projectBase+"/content", nil)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
