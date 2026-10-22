//go:build integration

package publish_test

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
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
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
	db        *store.DB
	sites     *gitsync.Sites
	tenancy   *tenancy.Store
	web       *publish.Web
	project   tenancy.Project
	remote    string
	gitCfg    gitsync.Config
	editor    httpapi.Actor
	implement httpapi.Actor
	settings  workflows.Settings
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

func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

// theSite is a Hugo site with a Quiqr model, which is what a customer brings.
func theSite() fstest.MapFS {
	return fstest.MapFS{
		"hugo.toml":             {Data: []byte("baseURL = \"https://example.org/\"\ntitle = \"The Customer\"\n")},
		"quiqr/model/base.yaml": {Data: []byte(theModel)},
		"content/_index.md":     {Data: []byte("---\ntitle: Welcome\nintro: We sell apples and pears.\n---\n\nThe old body.\n")},
		"layouts/index.html":    {Data: []byte("<!doctype html><html><head><title>{{ .Title }}</title></head><body>{{ .Content }}</body></html>\n")},
	}
}

func newHarness(t *testing.T, builds bool) *harness {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	people := auth.NewService(db, quiet(), "https://chatcms.example").People()
	owner, err := people.Create(t.Context(), "owner@example.org", "The Owner", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	anna, err := people.Create(t.Context(), "anna@example.org", "Anna", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	bram, err := people.Create(t.Context(), "bram@example.org", "Bram", "a long enough passphrase")
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
	if err := tenants.SetProjectRole(t.Context(), project.ID, bram.ID, policy.RoleImplementer); err != nil {
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

	repository, err := sites.Create(t.Context(), project.ID.String(), remote, theSite(),
		gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, "Bring the site in")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Publish(t.Context()); err != nil {
		t.Fatal(err)
	}
	repository.Close()

	form := "source"
	line := ""
	if builds {
		form = "built"
		line = "live"
	}
	if _, err := db.Pool().Exec(t.Context(),
		`UPDATE projects SET publish_form = $2, publish_to = $3, publish_branch = $4 WHERE id = $1`,
		pgtype.UUID{Bytes: project.ID, Valid: true}, form, lineKind(line), line); err != nil {
		t.Fatal(err)
	}

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
		return publish.Target{
			Line:       line,
			BuiltPages: builds,
			Author:     gitsync.Identity{Name: "Anna", Email: "anna@example.org"},
			Message:    "Published from Quiqr ChatCMS",
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	engine, err := workflows.NewEngine(sandbox, builder, publisher, toolbox(t),
		workflows.Limits{CheckTimeout: 30 * time.Second, CheckMemory: 512 << 20})
	if err != nil {
		t.Fatal(err)
	}

	service, err := publish.NewService(theSources{sites: sites, organisation: organisation.ID,
		builds: builds, baseURL: "https://korenaar.example/"}, engine, publish.NewRecords(db.Queries))
	if err != nil {
		t.Fatal(err)
	}
	web, err := publish.NewWeb(service, tenants)
	if err != nil {
		t.Fatal(err)
	}

	return &harness{
		db: db, sites: sites, tenancy: tenants, web: web, project: project,
		remote: remote, gitCfg: gitCfg,
		editor:    httpapi.Actor{ID: anna.ID.String(), Email: anna.Email},
		implement: httpapi.Actor{ID: bram.ID.String(), Email: bram.Email},
	}
}

func lineKind(line string) string {
	if line == "" {
		return "main-line"
	}
	return "its-own-line"
}

func toolbox(t *testing.T) []string {
	t.Helper()

	found := []string{}
	for _, name := range []string{"sh", "grep", "cat", "ls", "echo", "test"} {
		if path, err := exec.LookPath(name); err == nil {
			found = append(found, path)
		}
	}
	return found
}

type theSources struct {
	sites        *gitsync.Sites
	organisation uuid.UUID
	builds       bool
	baseURL      string
}

func (s theSources) Site(ctx context.Context, projectID uuid.UUID) (publish.Site, func() error, error) {
	repository, err := s.sites.Open(ctx, projectID.String())
	if err != nil {
		return nil, nil, err
	}
	return siteOf{Repo: repository}, repository.Close, nil
}

type siteOf struct{ gitsync.Repo }

func (s siteOf) Model() (*model.Model, error) {
	assembly, err := model.Assemble(s.FS())
	if err != nil {
		return nil, err
	}
	return model.Decode(assembly)
}

func (s theSources) Settings(ctx context.Context, projectID uuid.UUID) (workflows.Settings, error) {
	return workflows.Settings{ChatCMSBuilds: s.builds}, nil
}

func (s theSources) BaseURL(ctx context.Context, projectID uuid.UUID) (string, error) {
	return s.baseURL, nil
}

func (s theSources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(s.organisation.String(), projectID.String()), nil
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

func TestPublishingShowsWhatChangedAndThenPublishes(t *testing.T) {
	h := newHarness(t, false)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: We sell plums and pears.\n---\n\nThe new body.\n",
		"content/news.md":   "---\ntitle: News\n---\n\nSomething happened.\n",
	})

	review, err := h.web.Review(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if review.Count != 2 {
		t.Fatalf("Review() shows %d changes, want two", review.Count)
	}

	var home httpapi.ChangeView
	for _, change := range review.Changes {
		if change.Page == "Home page" {
			home = change
		}
	}
	if home.Kind != string(publish.ChangedPage) {
		t.Errorf("the home page is %q, want %q", home.Kind, publish.ChangedPage)
	}
	if home.By != "Anna" {
		t.Errorf("By = %q, want the person who edited it", home.By)
	}

	named := map[string]httpapi.FieldChangeView{}
	for _, field := range home.Fields {
		named[field.Name] = field
	}
	introduction, found := named["Introduction"]
	if !found {
		t.Fatalf("fields = %v, want the one the model calls Introduction", home.Fields)
	}
	added, gone := "", ""
	for _, word := range introduction.Words {
		switch word.Part {
		case string(publish.Added):
			added += word.Text
		case string(publish.Gone):
			gone += word.Text
		}
	}
	if strings.TrimSpace(gone) != "apples" || strings.TrimSpace(added) != "plums" {
		t.Errorf("the word diff removed %q and added %q", gone, added)
	}

	record, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "Spring prices")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if record.State != string(publish.Succeeded) {
		t.Fatalf("State = %q, trouble %q", record.State, record.Trouble)
	}

	onTheRemote := h.onTheRemote(t, "cat-file", "blob", "refs/heads/main:content/_index.md")
	if !strings.Contains(onTheRemote, "plums") {
		t.Errorf("the live site holds %q, want the edit", onTheRemote)
	}

	after, err := h.web.Review(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	if after.Count != 0 {
		t.Errorf("Review() still shows %d changes after publishing", after.Count)
	}
	if after.LastPublished == "" {
		t.Error("the review does not say when the site was last published")
	}
}

func TestPublishingBuiltPagesSendsTheHTML(t *testing.T) {
	h := newHarness(t, true)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: Now with plums.\n---\n\nThe new body.\n",
	})

	record, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if record.State != string(publish.Succeeded) {
		t.Fatalf("State = %q, trouble %q", record.State, record.Trouble)
	}

	page := h.onTheRemote(t, "cat-file", "blob", "refs/heads/live:index.html")
	if !strings.Contains(page, "The new body") {
		t.Errorf("the published page is %q, want the built HTML", page)
	}

	listed := h.onTheRemote(t, "ls-tree", "-r", "--name-only", "refs/heads/live")
	if strings.Contains(listed, "quiqr/model") {
		t.Error("the site's own files were published alongside the built pages")
	}

	steps := []string{}
	for _, step := range record.Steps {
		steps = append(steps, step.Title)
	}
	if len(steps) != 2 {
		t.Errorf("the publish ran %v, want building then putting it live", steps)
	}
}

func TestAFailedCheckLeavesTheLiveSiteAlone(t *testing.T) {
	h := newHarness(t, false)

	before := h.onTheRemote(t, "rev-parse", "refs/heads/main")

	h.edit(t, map[string]string{
		"quiqr/workflows/publish.yaml": "on: publish\nsteps:\n  - name: Checking pages and links\n    action: check\n    run: sh scripts/check.sh\n  - action: publish\n",
		"scripts/check.sh":             "#!/bin/sh\necho 'A page still says TODO.'\nexit 1\n",
		"content/news.md":              "---\ntitle: News\n---\n\nTODO: finish this.\n",
	})

	record, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if record.State != string(publish.Failed) {
		t.Fatalf("State = %q, want it to have failed", record.State)
	}
	if !strings.Contains(record.Trouble, "Nothing changed on your live site") {
		t.Errorf("Trouble = %q, want the reassurance", record.Trouble)
	}

	after := h.onTheRemote(t, "rev-parse", "refs/heads/main")
	if after != before {
		t.Error("the live site changed even though the check failed")
	}
}

func TestOnlyAnImplementerSeesTheRawOutput(t *testing.T) {
	h := newHarness(t, false)

	h.edit(t, map[string]string{
		"quiqr/workflows/publish.yaml": "on: publish\nsteps:\n  - action: check\n    run: sh scripts/check.sh\n  - action: publish\n",
		"scripts/check.sh":             "#!/bin/sh\necho 'fatal: cannot stat .git/HEAD'\nexit 1\n",
	})

	record, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "")
	if err != nil {
		t.Fatal(err)
	}

	forTheEditor, err := h.web.ReadPublish(t.Context(), h.editor, "technative", "korenaar", record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if forTheEditor.Detail != "" {
		t.Errorf("an editor was given the raw output: %q", forTheEditor.Detail)
	}

	forTheImplementer, err := h.web.ReadPublish(t.Context(), h.implement, "technative", "korenaar", record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(forTheImplementer.Detail, "cannot stat") {
		t.Errorf("an implementer was not given the raw output: %q", forTheImplementer.Detail)
	}
}

func TestEditingDuringAPublishCorruptsNeither(t *testing.T) {
	h := newHarness(t, true)

	h.edit(t, map[string]string{
		"content/_index.md": "---\ntitle: Welcome\nintro: The first version.\n---\n\nFirst.\n",
	})

	done := make(chan httpapi.PublishView, 1)
	go func() {
		record, err := h.web.Publish(context.Background(), h.editor, "technative", "korenaar", "")
		if err != nil {
			t.Error(err)
		}
		done <- record
	}()

	h.edit(t, map[string]string{
		"content/news.md": "---\ntitle: News\n---\n\nWritten while publishing.\n",
	})

	record := <-done
	if record.State != string(publish.Succeeded) {
		t.Fatalf("State = %q, trouble %q", record.State, record.Trouble)
	}

	repository, err := h.sites.Open(t.Context(), h.project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	page, err := fs.ReadFile(repository.FS(), "content/news.md")
	if err != nil {
		t.Fatalf("the edit made during the publish is gone: %v", err)
	}
	if !strings.Contains(string(page), "Written while publishing") {
		t.Errorf("the edit is %q", page)
	}
}

func TestThePublishHistoryIsKept(t *testing.T) {
	h := newHarness(t, false)

	h.edit(t, map[string]string{"content/news.md": "---\ntitle: News\n---\n"})
	if _, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "The first update"); err != nil {
		t.Fatal(err)
	}
	h.edit(t, map[string]string{"content/more.md": "---\ntitle: More\n---\n"})
	if _, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "The second update"); err != nil {
		t.Fatal(err)
	}

	history, err := h.web.PublishHistory(t.Context(), h.editor, "technative", "korenaar")
	if err != nil {
		t.Fatalf("PublishHistory() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("the history holds %d publishes, want two", len(history))
	}
	if history[0].Note != "The second update" {
		t.Errorf("the first listed is %q, want the most recent", history[0].Note)
	}
	if history[0].ChangeCount != 1 {
		t.Errorf("ChangeCount = %d, want what it carried", history[0].ChangeCount)
	}
}

func TestSomebodyElsesProjectIsNotThere(t *testing.T) {
	h := newHarness(t, false)

	stranger := httpapi.Actor{ID: uuid.NewString(), Email: "stranger@example.org"}
	if _, err := h.web.Review(t.Context(), stranger, "technative", "korenaar"); err == nil {
		t.Error("Review() for a stranger error = nil, want it to be refused")
	}
}

func TestPublishingNothingIsNotAFailure(t *testing.T) {
	h := newHarness(t, false)

	record, err := h.web.Publish(t.Context(), h.editor, "technative", "korenaar", "")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if record.State != string(publish.Succeeded) {
		t.Errorf("State = %q, trouble %q", record.State, record.Trouble)
	}
	if record.ChangeCount != 0 {
		t.Errorf("ChangeCount = %d, want nothing", record.ChangeCount)
	}
}
