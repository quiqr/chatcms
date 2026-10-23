//go:build integration

package history_test

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
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/history"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
	"github.com/quiqr/chatcms/internal/runner"
	"github.com/quiqr/chatcms/internal/sitebuild"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/tenancy"
	"github.com/quiqr/chatcms/internal/testsupport"
	"github.com/quiqr/chatcms/internal/workflows"
)

const theModel = `singles:
  - key: home
    title: Home page
    file: content/_index.md
    fields:
      - key: title
        type: string
        title: Page title
      - key: intro
        type: markdown
        title: Introduction
`

type harness struct {
	sites   *gitsync.Sites
	web     *history.Web
	publish *publish.Web
	project tenancy.Project
	remote  string
	gitCfg  gitsync.Config
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

func hugoBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("CHATCMS_HUGO_BIN"); bin != "" {
		return bin
	}
	found, err := exec.LookPath("hugo")
	if err != nil {
		t.Skipf("Hugo is not available: %v", err)
	}
	return found
}

func theWholeSite() fstest.MapFS {
	return fstest.MapFS{
		"hugo.toml":             {Data: []byte("baseURL = \"https://example.org/\"\ntitle = \"The Customer\"\n")},
		"quiqr/model/base.yaml": {Data: []byte(theModel)},
		"content/_index.md":     {Data: []byte("---\ntitle: Welcome\nintro: The first words.\n---\n\nThe first body.\n")},
		"layouts/index.html":    {Data: []byte("<!doctype html><html><body>{{ .Content }}</body></html>\n")},
	}
}

func newHarness(t *testing.T) *harness {
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
	anna, err := identity.People().Create(t.Context(), "anna@example.org", "Anna", "a long enough passphrase")
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

	gitCfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(gitCfg, filepath.Join(t.TempDir(), "sites"), nil)

	remote := filepath.Join(t.TempDir(), "remote.git")
	if err := gitsync.RunForTest(t.Context(), gitCfg, "", t.TempDir(),
		"init", "--bare", "--initial-branch=main", remote); err != nil {
		t.Fatal(err)
	}

	repository, err := sites.Create(t.Context(), project.ID.String(), remote, theWholeSite(),
		gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, "Bring the site in")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}
	repository.Close()

	sandbox, err := runner.NewSandbox(runner.Config{WorkRoot: t.TempDir()})
	if err != nil {
		if errors.Is(err, runner.ErrNoIsolation) {
			t.Skipf("no isolation available here: %v", err)
		}
		t.Fatal(err)
	}
	builder, err := sitebuild.New(sandbox, hugoBinary(t), sitebuild.Limits{Timeout: 2 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	publisher, err := publish.NewPublisher(sites, func(ctx context.Context, projectID uuid.UUID) (publish.Target, error) {
		return publish.Target{Author: gitsync.Identity{Name: "Anna", Email: "anna@example.org"}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	engine, err := workflows.NewEngine(sandbox, builder, publisher, nil, workflows.Limits{})
	if err != nil {
		t.Fatal(err)
	}

	records := publish.NewRecords(db.Queries)
	sources := theSources{sites: sites, organisation: organisation.ID, records: records, people: identity}
	publishing, err := publish.NewService(publishSide{sources}, engine, records)
	if err != nil {
		t.Fatal(err)
	}
	publishWeb, err := publish.NewWeb(publishing, tenants)
	if err != nil {
		t.Fatal(err)
	}

	timeline, err := history.NewService(sources, publishing)
	if err != nil {
		t.Fatal(err)
	}
	web, err := history.NewWeb(timeline, tenants)
	if err != nil {
		t.Fatal(err)
	}

	return &harness{
		sites: sites, web: web, publish: publishWeb, project: project, remote: remote, gitCfg: gitCfg,
		editor: httpapi.Actor{ID: anna.ID.String(), Email: anna.Email},
		viewer: httpapi.Actor{ID: wim.ID.String(), Email: wim.Email},
	}
}

type theSources struct {
	sites        *gitsync.Sites
	organisation uuid.UUID
	records      *publish.Records
	people       *auth.Service
}

type theSite struct{ gitsync.Repo }

func (s theSite) Model() (*model.Model, error) {
	assembly, err := model.Assemble(s.FS())
	if err != nil {
		return nil, err
	}
	return model.Decode(assembly)
}

func (s theSources) open(ctx context.Context, projectID uuid.UUID) (theSite, func() error, error) {
	repository, err := s.sites.Open(ctx, projectID.String())
	if err != nil {
		return theSite{}, nil, err
	}
	return theSite{Repo: repository}, repository.Close, nil
}

func (s theSources) Site(ctx context.Context, projectID uuid.UUID) (history.Site, func() error, error) {
	site, done, err := s.open(ctx, projectID)
	return site, done, err
}

// publishSide is the same sources answering the publish service, whose Site
// returns a smaller interface. The server does the same thing: one method
// name cannot return two types.
type publishSide struct{ theSources }

func (s publishSide) Site(ctx context.Context, projectID uuid.UUID) (publish.Site, func() error, error) {
	site, done, err := s.open(ctx, projectID)
	return site, done, err
}

func (s theSources) Settings(ctx context.Context, projectID uuid.UUID) (workflows.Settings, error) {
	return workflows.Settings{}, nil
}

func (s theSources) BaseURL(ctx context.Context, projectID uuid.UUID) (string, error) {
	return "https://korenaar.example/", nil
}

func (s theSources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(s.organisation.String(), projectID.String()), nil
}

func (s theSources) Notes(ctx context.Context, projectID uuid.UUID) (map[string]string, error) {
	return map[string]string{}, nil
}

func (s theSources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	person, err := s.people.People().ByID(ctx, personID)
	if err != nil {
		return gitsync.Identity{}, err
	}
	return gitsync.Identity{Name: person.DisplayName, Email: person.Email}, nil
}

func (h *harness) edit(t *testing.T, writes map[string]string) {
	t.Helper()

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	contents := map[string][]byte{}
	for name, body := range writes {
		contents[name] = []byte(body)
	}
	if _, err := repository.Save(t.Context(), gitsync.Save{
		Message: "Edit the site",
		Author:  gitsync.Identity{Name: "Anna", Email: "anna@example.org"},
		Writes:  contents,
	}); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) onTheRemote(t *testing.T, args ...string) string {
	t.Helper()

	out, err := gitsync.OutputForTest(t.Context(), h.gitCfg, h.remote, t.TempDir(), args...)
	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
	return out
}

func (h *harness) timeline(t *testing.T) httpapi.TimelineView {
	t.Helper()

	timeline, err := h.web.Timeline(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatalf("Timeline() error = %v", err)
	}
	return timeline
}

func versionTitled(timeline httpapi.TimelineView, title string) (httpapi.VersionView, bool) {
	for _, group := range timeline.Groups {
		for _, entry := range group.Entries {
			if strings.Contains(entry.Title, title) {
				return entry, true
			}
		}
	}
	return httpapi.VersionView{}, false
}

func TestGoingBackToAnEarlierVersionAndThenUndoingIt(t *testing.T) {
	h := newHarness(t)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: The second words.\n---\n\nThe second body.\n",
	})
	if _, err := h.publish.Publish(t.Context(), h.editor, "technative", "korenaar", "The second update"); err != nil {
		t.Fatal(err)
	}

	live := h.onTheRemote(t, "cat-file", "blob", "refs/heads/main:content/_index.md")
	if !strings.Contains(live, "The second words") {
		t.Fatalf("the live site holds %q, want the second version", live)
	}

	timeline := h.timeline(t)
	first, found := versionTitled(timeline, "Bring the site in")
	if !found {
		t.Fatalf("the first version is not in the timeline: %v", timeline.Groups)
	}

	warning, err := h.web.BeforeGoingBack(t.Context(), h.editor, "technative", "korenaar", first.Version)
	if err != nil {
		t.Fatalf("BeforeGoingBack() error = %v", err)
	}
	if !strings.Contains(warning.Reassurance, "Nothing is lost") {
		t.Errorf("Reassurance = %q", warning.Reassurance)
	}

	back, err := h.web.GoBackTo(t.Context(), h.editor, "technative", "korenaar", first.Version, true)
	if err != nil {
		t.Fatalf("GoBackTo() error = %v", err)
	}
	if back.Published == nil || back.Published.State != string(publish.Succeeded) {
		t.Fatalf("going back did not publish: %+v", back.Published)
	}

	live = h.onTheRemote(t, "cat-file", "blob", "refs/heads/main:content/_index.md")
	if !strings.Contains(live, "The first words") {
		t.Fatalf("the live site holds %q, want the first version back", live)
	}

	// Going back is itself something to undo: the version it moved away from
	// is still in the history, and going back to that restores it.
	after := h.timeline(t)
	second, found := versionTitled(after, "Edit the site")
	if !found {
		t.Fatalf("the version it moved away from is gone: %v", after.Groups)
	}

	if _, err := h.web.GoBackTo(t.Context(), h.editor, "technative", "korenaar", second.Version, true); err != nil {
		t.Fatalf("undoing the rollback error = %v", err)
	}

	live = h.onTheRemote(t, "cat-file", "blob", "refs/heads/main:content/_index.md")
	if !strings.Contains(live, "The second words") {
		t.Fatalf("the live site holds %q, want the second version restored", live)
	}
}

func TestUnpublishedWorkIsSetAsideRatherThanLost(t *testing.T) {
	h := newHarness(t)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: The second words.\n---\n\nThe second body.\n",
	})
	if _, err := h.publish.Publish(t.Context(), h.editor, "technative", "korenaar", "The second update"); err != nil {
		t.Fatal(err)
	}

	h.edit(t, map[string]string{
		"content/draft.md": "---\ntitle: A draft\n---\n\nNot ready yet.\n",
	})

	timeline := h.timeline(t)
	if timeline.Unpublished != 1 {
		t.Fatalf("Unpublished = %d, want the one draft", timeline.Unpublished)
	}

	first, _ := versionTitled(timeline, "Bring the site in")
	warning, err := h.web.BeforeGoingBack(t.Context(), h.editor, "technative", "korenaar", first.Version)
	if err != nil {
		t.Fatal(err)
	}
	if warning.SetAside != 1 {
		t.Errorf("SetAside = %d, want the unpublished draft named", warning.SetAside)
	}
	if !strings.Contains(warning.Reassurance, "set aside") {
		t.Errorf("Reassurance = %q, want it to say what happens to unpublished work", warning.Reassurance)
	}

	if _, err := h.web.GoBackTo(t.Context(), h.editor, "technative", "korenaar", first.Version, false); err != nil {
		t.Fatal(err)
	}

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	if _, err := fs.Stat(repository.FS(), "content/draft.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the draft is still in the site: %v", err)
	}

	// Set aside, not lost: the version that holds it is still in the history,
	// and going back to it brings the draft out again.
	after := h.timeline(t)
	withDraft, found := versionTitled(after, "Edit the site")
	if !found {
		t.Fatalf("the version holding the draft is gone: %v", after.Groups)
	}
	if _, err := h.web.GoBackTo(t.Context(), h.editor, "technative", "korenaar", withDraft.Version, false); err != nil {
		t.Fatal(err)
	}

	again, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()

	draft, err := fs.ReadFile(again.FS(), "content/draft.md")
	if err != nil {
		t.Fatalf("the draft could not be brought back: %v", err)
	}
	if !strings.Contains(string(draft), "Not ready yet") {
		t.Errorf("the draft reads %q", draft)
	}
}

func TestTheTimelineReadsAsPlainLanguage(t *testing.T) {
	h := newHarness(t)

	h.edit(t, map[string]string{"content/news.md": "---\ntitle: News\n---\n"})
	if _, err := h.publish.Publish(t.Context(), h.editor, "technative", "korenaar", "Spring prices"); err != nil {
		t.Fatal(err)
	}

	timeline := h.timeline(t)

	if len(timeline.Groups) == 0 {
		t.Fatal("the timeline is empty")
	}
	for _, group := range timeline.Groups {
		if group.Period == "" {
			t.Error("a group has no period")
		}
		for _, entry := range group.Entries {
			if entry.Title == "" {
				t.Error("a version has no title")
			}
			for _, word := range []string{"commit", "SHA", "branch", "HEAD", "ref"} {
				if strings.Contains(strings.ToLower(entry.Title), strings.ToLower(word)) {
					t.Errorf("the title %q carries the word %q", entry.Title, word)
				}
			}
			if strings.Contains(entry.Title, entry.Version) {
				t.Errorf("the title %q carries the version identifier", entry.Title)
			}
		}
	}

	live := 0
	for _, group := range timeline.Groups {
		for _, entry := range group.Entries {
			if entry.Live {
				live++
			}
		}
	}
	if live != 1 {
		t.Errorf("%d versions are marked live, want exactly one", live)
	}
}

func TestWhatAVersionContainsIsShownAsPagesAndFields(t *testing.T) {
	h := newHarness(t)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: The second words.\n---\n\nThe second body.\n",
	})

	timeline := h.timeline(t)
	latest := timeline.Groups[0].Entries[0]

	contents, err := h.web.VersionContents(t.Context(), h.editor, "technative", "korenaar", latest.Version)
	if err != nil {
		t.Fatalf("VersionContents() error = %v", err)
	}

	if contents.Count != 1 {
		t.Fatalf("that version changed %d things, want one", contents.Count)
	}
	if contents.Changes[0].Page != "Home page" {
		t.Errorf("Page = %q, want the model's title", contents.Changes[0].Page)
	}

	named := map[string]bool{}
	for _, field := range contents.Changes[0].Fields {
		named[field.Name] = true
	}
	if !named["Introduction"] {
		t.Errorf("fields = %v, want the one the model calls Introduction", contents.Changes[0].Fields)
	}
}

func TestAViewerMayLookButNotGoBack(t *testing.T) {
	h := newHarness(t)

	h.edit(t, map[string]string{"content/news.md": "---\ntitle: News\n---\n"})

	if _, err := h.web.Timeline(t.Context(), h.viewer, "technative", "korenaar"); err != nil {
		t.Errorf("Timeline() as a viewer error = %v, want it allowed", err)
	}

	timeline := h.timeline(t)
	first, _ := versionTitled(timeline, "Bring the site in")
	if _, err := h.web.GoBackTo(t.Context(), h.viewer, "technative", "korenaar", first.Version, false); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("GoBackTo() as a viewer error = %v, want ErrNotAllowed", err)
	}
}

func TestGoingBackToWhereTheSiteAlreadyIsIsRefusedKindly(t *testing.T) {
	h := newHarness(t)

	timeline := h.timeline(t)
	latest := timeline.Groups[0].Entries[0]

	_, err := h.web.GoBackTo(t.Context(), h.editor, "technative", "korenaar", latest.Version, false)

	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("GoBackTo() error = %v, want it refused", err)
	}
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("the refusal says %q, want it to say why", err)
	}
}
