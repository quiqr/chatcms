//go:build integration

package content_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/content"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
)

const siteModel = `singles:
  - key: home
    title: Home page
    file: content/_index.md
    fields:
      - key: title
        type: string
        title: Page title
      - key: tags
        type: leaf-array
        title: Tags
        field:
          key: tag
          type: string
      - key: mainContent
        type: markdown
        title: Page text
collections:
  - key: news
    title: News
    folder: content/news
    fields:
      - key: title
        type: string
        title: Headline
`

const siteHomePage = `---
title: Korenaar Bakery
tags:
  - bread
  - cake
---

Come in and smell it.
`

type webHarness struct {
	sites   *gitsync.Sites
	web     *content.Web
	project tenancy.Project
	editor  httpapi.Actor
	viewer  httpapi.Actor
}

func gitBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("CHATCMS_GIT_BIN"); bin != "" {
		return bin
	}
	found, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}
	return found
}

func newWebHarness(t *testing.T) *webHarness {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	identity := auth.NewService(db, slog.New(slog.NewJSONHandler(io.Discard, nil)), "https://chatcms.example")
	owner, err := identity.People().Create(t.Context(), "owner@example.org", "The Owner", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	anna, err := identity.People().Create(t.Context(), "anna@example.org", "Anna de Wit", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	wim, err := identity.People().Create(t.Context(), "wim@example.org", "Wim", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}

	tenants := tenancy.NewStore(db)
	organisation, err := tenants.CreateOrganisation(t.Context(), "technative", "TechNative", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	project, err := tenants.CreateProject(t.Context(), organisation.ID, "korenaar", "Korenaar", "korenaar.example")
	if err != nil {
		t.Fatal(err)
	}
	if err := tenants.SetProjectRole(t.Context(), project.ID, anna.ID, policy.RoleEditor); err != nil {
		t.Fatal(err)
	}
	if err := tenants.SetProjectRole(t.Context(), project.ID, wim.ID, policy.RoleViewer); err != nil {
		t.Fatal(err)
	}

	cfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(cfg, filepath.Join(t.TempDir(), "sites"), nil)

	remote := filepath.Join(t.TempDir(), "remote.git")
	if err := gitsync.InitBareForDevelopment(t.Context(), cfg, remote); err != nil {
		t.Fatal(err)
	}

	repository, err := sites.Create(t.Context(), project.ID.String(), remote, fstest.MapFS{
		"hugo.toml":             {Data: []byte("title = \"Korenaar\"\n")},
		"quiqr/model/base.yaml": {Data: []byte(siteModel)},
		"content/_index.md":     {Data: []byte(siteHomePage)},
		"content/news/one.md":   {Data: []byte("---\ntitle: One\n---\n\nFirst.\n")},
	}, gitsync.Identity{Name: "The Owner", Email: owner.Email}, "Bring the site in")
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	service, err := content.NewService(theSources{sites: sites, organisation: organisation.ID, people: identity}, nil)
	if err != nil {
		t.Fatal(err)
	}
	web, err := content.NewWeb(service, tenants)
	if err != nil {
		t.Fatal(err)
	}

	return &webHarness{
		sites: sites, web: web, project: project,
		editor: httpapi.Actor{ID: anna.ID.String(), Email: anna.Email},
		viewer: httpapi.Actor{ID: wim.ID.String(), Email: wim.Email},
	}
}

type theSources struct {
	sites        *gitsync.Sites
	organisation uuid.UUID
	people       *auth.Service
}

type aWholeSite struct{ gitsync.Repo }

func (s aWholeSite) Model() (*model.Model, error) {
	assembly, err := model.Assemble(s.FS())
	if err != nil {
		return nil, err
	}
	return model.Decode(assembly)
}

func (s theSources) Site(ctx context.Context, projectID uuid.UUID) (content.Site, func() error, error) {
	repository, err := s.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return aWholeSite{Repo: repository}, repository.Close, nil
}

func (s theSources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(s.organisation.String(), projectID.String()), nil
}

func (s theSources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	person, err := s.people.People().ByID(ctx, personID)
	if err != nil {
		return gitsync.Identity{}, err
	}
	return gitsync.Identity{Name: person.DisplayName, Email: person.Email}, nil
}

func (h *webHarness) read(t *testing.T) string {
	t.Helper()

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	content, err := fs.ReadFile(repository.FS(), "content/_index.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func fieldAt(fields []httpapi.ControlView, path string) (httpapi.ControlView, bool) {
	for _, field := range fields {
		if field.Path == path {
			return field, true
		}
		if found, ok := fieldAt(field.Fields, path); ok {
			return found, true
		}
	}
	return httpapi.ControlView{}, false
}

func TestAPageIsReadAndWrittenThroughTheWholeStack(t *testing.T) {
	h := newWebHarness(t)

	page, err := h.web.ReadPage(t.Context(), h.editor, "technative", "korenaar", "home")
	if err != nil {
		t.Fatalf("ReadPage() error = %v", err)
	}
	if page.Title != "Home page" || !page.HasBody {
		t.Fatalf("page = %+v", page)
	}

	title, found := fieldAt(page.Fields, "title")
	if !found || title.Value != "Korenaar Bakery" {
		t.Fatalf("title = %+v", title)
	}
	if _, drawn := fieldAt(page.Fields, "mainContent"); drawn {
		t.Error("the page's own text is drawn as an ordinary field as well")
	}

	body := "Come in, we are open.\n"
	saved, err := h.web.SavePage(t.Context(), h.editor, "technative", "korenaar", "home", httpapi.EditRequest{
		Values: map[string]any{"title": "Korenaar Bakery & Co", "tags": []any{"bread"}},
		Body:   &body,
	})
	if err != nil {
		t.Fatalf("SavePage() error = %v", err)
	}

	after, _ := fieldAt(saved.Fields, "title")
	if after.Value != "Korenaar Bakery & Co" {
		t.Errorf("the page came back with %v", after.Value)
	}

	written := h.read(t)
	for _, want := range []string{"title: Korenaar Bakery & Co", "  - bread", "Come in, we are open."} {
		if !strings.Contains(written, want) {
			t.Errorf("the file does not carry %q:\n%s", want, written)
		}
	}
	if strings.Contains(written, "cake") {
		t.Errorf("the removed tag is still there:\n%s", written)
	}
}

func TestTheVersionIsAttributedToWhoeverEdited(t *testing.T) {
	h := newWebHarness(t)

	if _, err := h.web.SavePage(t.Context(), h.editor, "technative", "korenaar", "home",
		httpapi.EditRequest{Values: map[string]any{"title": "Changed"}}); err != nil {
		t.Fatal(err)
	}

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	versions, err := repository.History(t.Context(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if versions[0].Author.Name != "Anna de Wit" {
		t.Errorf("the version is by %q, want the person who edited", versions[0].Author.Name)
	}
}

func TestAnItemOfACollectionIsEditedThroughItsWholeName(t *testing.T) {
	h := newWebHarness(t)

	page, err := h.web.ReadPage(t.Context(), h.editor, "technative", "korenaar", "news/one")
	if err != nil {
		t.Fatalf("ReadPage() error = %v", err)
	}
	if page.Section != "News" {
		t.Errorf("Section = %q", page.Section)
	}

	if _, err := h.web.SavePage(t.Context(), h.editor, "technative", "korenaar", "news/one",
		httpapi.EditRequest{Values: map[string]any{"title": "The first"}}); err != nil {
		t.Fatal(err)
	}

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	written, err := fs.ReadFile(repository.FS(), "content/news/one.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "title: The first") {
		t.Errorf("the item reads:\n%s", written)
	}
}

func TestAViewerMayOpenAPageButNotChangeIt(t *testing.T) {
	h := newWebHarness(t)

	if _, err := h.web.ReadPage(t.Context(), h.viewer, "technative", "korenaar", "home"); err != nil {
		t.Errorf("ReadPage() as a viewer error = %v, want it allowed", err)
	}

	before := h.read(t)
	if _, err := h.web.SavePage(t.Context(), h.viewer, "technative", "korenaar", "home",
		httpapi.EditRequest{Values: map[string]any{"title": "Changed"}}); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("SavePage() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if h.read(t) != before {
		t.Error("a viewer changed the file")
	}
}

func TestSomebodyElsesProjectIsNotThere(t *testing.T) {
	h := newWebHarness(t)

	stranger := httpapi.Actor{ID: uuid.NewString(), Email: "stranger@example.org"}
	if _, err := h.web.ReadPage(t.Context(), stranger, "technative", "korenaar", "home"); err == nil {
		t.Error("a stranger read a page")
	}
	if _, err := h.web.ReadPage(t.Context(), h.editor, "technative", "somewhere-else", "home"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("a project that is not there error = %v, want ErrMissing", err)
	}
}

func TestAPageThatIsNotThereIsNotFound(t *testing.T) {
	h := newWebHarness(t)

	for _, key := range []string{"nothing", "news/../../etc/passwd", "news/absent"} {
		if _, err := h.web.ReadPage(t.Context(), h.editor, "technative", "korenaar", key); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("ReadPage(%q) error = %v, want ErrMissing", key, err)
		}
	}
}

func TestTheBrowserAndTheWholeLifeOfAPage(t *testing.T) {
	h := newWebHarness(t)

	sections, err := h.web.ListContent(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatalf("ListContent() error = %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("sections = %+v", sections)
	}

	made, err := h.web.CreatePage(t.Context(), h.editor, "technative", "korenaar", "news", "Spring prices")
	if err != nil {
		t.Fatalf("CreatePage() error = %v", err)
	}
	if made.Key != "news/spring-prices" {
		t.Fatalf("Key = %q", made.Key)
	}

	renamed, err := h.web.RenamePage(t.Context(), h.editor, "technative", "korenaar", made.Key, "Summer prices")
	if err != nil {
		t.Fatalf("RenamePage() error = %v", err)
	}
	if renamed.Key != "news/summer-prices" {
		t.Errorf("Key = %q", renamed.Key)
	}

	copied, err := h.web.DuplicatePage(t.Context(), h.editor, "technative", "korenaar", renamed.Key, "Autumn prices")
	if err != nil {
		t.Fatalf("DuplicatePage() error = %v", err)
	}

	after, err := h.web.ListContent(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	named := map[string]bool{}
	for _, section := range after {
		for _, item := range section.Items {
			named[item.Key] = true
			if section.Key == "news" && !item.Waiting {
				t.Errorf("%s is not marked as waiting to be published", item.Key)
			}
		}
	}
	for _, want := range []string{"news/summer-prices", "news/autumn-prices", "news/one"} {
		if !named[want] {
			t.Errorf("%s is missing from the browser", want)
		}
	}
	if named["news/spring-prices"] {
		t.Error("the renamed page is still listed under its old name")
	}

	if err := h.web.DeletePage(t.Context(), h.editor, "technative", "korenaar", copied.Key); err != nil {
		t.Fatalf("DeletePage() error = %v", err)
	}

	finally, err := h.web.ListContent(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range finally {
		for _, item := range section.Items {
			if item.Key == copied.Key {
				t.Error("the removed page is still listed")
			}
		}
	}
}

func TestAViewerMayBrowseAndChangeNothing(t *testing.T) {
	h := newWebHarness(t)

	if _, err := h.web.ListContent(t.Context(), h.viewer, "technative", "korenaar"); err != nil {
		t.Errorf("ListContent() as a viewer error = %v, want it allowed", err)
	}
	if _, err := h.web.CreatePage(t.Context(), h.viewer, "technative", "korenaar", "news", "x"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("CreatePage() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if err := h.web.DeletePage(t.Context(), h.viewer, "technative", "korenaar", "news/one"); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("DeletePage() as a viewer error = %v, want ErrNotAllowed", err)
	}
}
