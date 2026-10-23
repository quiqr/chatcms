package sitesync_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/sitesync"
)

const organisation = "11111111-1111-1111-1111-111111111111"

func aModel() *model.Model {
	return &model.Model{
		Singles: []model.Single{{
			Key: "home", Title: "Home page", File: "content/_index.md",
			Fields: []model.Field{
				{Key: "title", Type: "string", Title: "Page title"},
				{Key: "summary", Type: "string", Title: "Summary"},
				{Key: "mainContent", Type: "markdown", Title: "Page text"},
			},
		}},
	}
}

type site struct {
	status     gitsync.Status
	syncResult gitsync.SyncResult
	syncErr    error
	resolved   map[string][]byte
	abandoned  int
	model      *model.Model
}

func (s *site) Status(ctx context.Context) (gitsync.Status, error) { return s.status, nil }

func (s *site) Sync(ctx context.Context) (gitsync.SyncResult, error) {
	return s.syncResult, s.syncErr
}

func (s *site) Resolve(ctx context.Context, resolutions map[string][]byte) (gitsync.SyncResult, error) {
	s.resolved = resolutions
	return gitsync.SyncResult{}, nil
}

func (s *site) Abandon(ctx context.Context) error { s.abandoned++; return nil }

func (s *site) History(ctx context.Context, limit int) ([]gitsync.Version, error) {
	return []gitsync.Version{{ID: "abc123"}}, nil
}

func (s *site) Model() (*model.Model, error) {
	if s.model == nil {
		return nil, model.ErrNoModel
	}
	return s.model, nil
}

type sources struct {
	site      *site
	technical sitesync.Technical
	closed    int
}

func (s *sources) Site(ctx context.Context, projectID uuid.UUID) (sitesync.Site, func() error, error) {
	return s.site, func() error { s.closed++; return nil }, nil
}

func (s *sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(organisation, projectID.String()), nil
}

func (s *sources) Technical(ctx context.Context, projectID uuid.UUID) (sitesync.Technical, error) {
	return s.technical, nil
}

func anEditor(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleEditor},
	}
}

func anImplementer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleImplementer},
	}
}

func aViewer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleViewer},
	}
}

func service(t *testing.T, src *sources) *sitesync.Service {
	t.Helper()

	made, err := sitesync.NewService(src)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

const minePage = `---
title: Korenaar Bakery
summary: We bake with apples.
---

My body.
`

const theirPage = `---
title: Korenaar Bakery
summary: We bake with plums.
---

Their body.
`

const basePage = `---
title: Korenaar Bakery
summary: We bake.
---

A body.
`

func aConflict() *gitsync.Conflict {
	return &gitsync.Conflict{Paths: []gitsync.ConflictedPath{{
		Path:   "content/_index.md",
		Base:   []byte(basePage),
		Local:  []byte(minePage),
		Remote: []byte(theirPage),
	}}}
}

func TestEveryStateIsReachable(t *testing.T) {
	project := uuid.New()

	cases := map[sitesync.State]*site{
		sitesync.Saved:  {status: gitsync.Status{}},
		sitesync.Saving: {status: gitsync.Status{UnsavedWork: true}},
		sitesync.Asking: {status: gitsync.Status{ExchangeInProgress: true}},
	}

	for want, working := range cases {
		standing, err := service(t, &sources{site: working}).
			Standing(context.Background(), anEditor(project), project)
		if err != nil {
			t.Fatalf("Standing() error = %v", err)
		}
		if standing.State != want {
			t.Errorf("State = %q, want %q", standing.State, want)
		}
	}
}

func TestBeingUnreachableIsSaidAsOffline(t *testing.T) {
	project := uuid.New()
	working := &site{syncErr: gitsync.ErrRemoteUnreachable}

	standing, err := service(t, &sources{site: working}).
		Bring(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("Bring() error = %v, want it reported as offline rather than failing", err)
	}
	if standing.State != sitesync.Offline {
		t.Errorf("State = %q, want %q", standing.State, sitesync.Offline)
	}
}

func TestWorkThatArrivedNamesWhoeverSentIt(t *testing.T) {
	project := uuid.New()
	working := &site{syncResult: gitsync.SyncResult{Arrived: []gitsync.Version{
		{Author: gitsync.Identity{Name: "Joris"}},
		{Author: gitsync.Identity{Name: "Joris"}},
		{Author: gitsync.Identity{Name: "Anna"}},
	}}}

	standing, err := service(t, &sources{site: working}).
		Bring(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	if standing.State != sitesync.Arrived {
		t.Errorf("State = %q, want %q", standing.State, sitesync.Arrived)
	}
	if len(standing.ArrivedFrom) != 2 {
		t.Errorf("ArrivedFrom = %v, want each person once", standing.ArrivedFrom)
	}
}

func TestTheSameFieldChangedTwiceBecomesAQuestionAboutThatField(t *testing.T) {
	project := uuid.New()
	working := &site{
		model:      aModel(),
		syncResult: gitsync.SyncResult{Conflict: aConflict(), Arrived: []gitsync.Version{{Author: gitsync.Identity{Name: "Joris"}}}},
	}

	standing, err := service(t, &sources{site: working}).
		Bring(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("Bring() error = %v", err)
	}

	if standing.State != sitesync.Asking {
		t.Fatalf("State = %q, want %q", standing.State, sitesync.Asking)
	}
	if len(standing.Conflicts) != 1 {
		t.Fatalf("conflicts = %+v", standing.Conflicts)
	}

	page := standing.Conflicts[0]
	if page.Title != "Home page" || page.Page != "home" {
		t.Errorf("page = %+v", page)
	}
	if page.Who != "Joris" {
		t.Errorf("Who = %q, want whoever sent it", page.Who)
	}
	if page.WholeFile {
		t.Error("a difference this can name was reported as a whole-file choice")
	}

	named := map[string]sitesync.FieldConflict{}
	for _, field := range page.Fields {
		named[field.Name] = field
	}

	summary, found := named["Summary"]
	if !found {
		t.Fatalf("fields = %+v, want the one the model calls Summary", page.Fields)
	}
	if summary.Mine != "We bake with apples." || summary.Theirs != "We bake with plums." {
		t.Errorf("summary = %+v", summary)
	}
	if len(summary.Words) == 0 {
		t.Error("the two answers carry no comparison")
	}
	if _, found := named["Page title"]; found {
		t.Error("a field both people left alone was reported as a question")
	}
	if _, found := named[sitesync.BodyField]; !found {
		t.Errorf("fields = %+v, want the page's own text listed too", page.Fields)
	}
}

func TestAFileThisCannotReadIsAWholeFileChoice(t *testing.T) {
	project := uuid.New()
	working := &site{
		model: aModel(),
		syncResult: gitsync.SyncResult{Conflict: &gitsync.Conflict{Paths: []gitsync.ConflictedPath{{
			Path:   "static/logo.png",
			Local:  []byte{0x89, 'P', 'N', 'G'},
			Remote: []byte{0x89, 'P', 'N', 'G', 0x01},
		}}}},
	}

	standing, err := service(t, &sources{site: working}).
		Bring(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	if len(standing.Conflicts) != 1 || !standing.Conflicts[0].WholeFile {
		t.Errorf("conflicts = %+v, want one whole-file choice", standing.Conflicts)
	}
}

func TestKeepingMineAndTheirsFieldByField(t *testing.T) {
	project := uuid.New()
	working := &site{model: aModel(), syncResult: gitsync.SyncResult{Conflict: aConflict()}}

	if _, err := service(t, &sources{site: working}).Answered(context.Background(), anEditor(project), project,
		[]sitesync.Answer{{
			File:   "content/_index.md",
			Fields: map[string]sitesync.Choice{"summary": sitesync.Theirs},
		}}); err != nil {
		t.Fatalf("Answered() error = %v", err)
	}

	written := string(working.resolved["content/_index.md"])
	if !strings.Contains(written, "We bake with plums.") {
		t.Errorf("the field somebody wanted from the other version is not there:\n%s", written)
	}
	if !strings.Contains(written, "My body.") {
		t.Errorf("a field nobody asked about did not keep what was written here:\n%s", written)
	}
}

func TestKeepingAWholeVersion(t *testing.T) {
	project := uuid.New()

	for _, test := range []struct {
		choice sitesync.Choice
		want   string
	}{
		{sitesync.Theirs, "Their body."},
		{sitesync.Mine, "My body."},
	} {
		working := &site{model: aModel(), syncResult: gitsync.SyncResult{Conflict: aConflict()}}

		if _, err := service(t, &sources{site: working}).Answered(context.Background(), anEditor(project), project,
			[]sitesync.Answer{{File: "content/_index.md", Whole: test.choice}}); err != nil {
			t.Fatal(err)
		}

		written := string(working.resolved["content/_index.md"])
		if !strings.Contains(written, test.want) {
			t.Errorf("keeping %q gave:\n%s", test.choice, written)
		}
	}
}

func TestEveryPageNeedsAnAnswer(t *testing.T) {
	project := uuid.New()
	working := &site{
		model: aModel(),
		syncResult: gitsync.SyncResult{Conflict: &gitsync.Conflict{Paths: []gitsync.ConflictedPath{
			{Path: "content/_index.md", Local: []byte(minePage), Remote: []byte(theirPage)},
			{Path: "content/other.md", Local: []byte(minePage), Remote: []byte(theirPage)},
		}}},
	}

	_, err := service(t, &sources{site: working}).Answered(context.Background(), anEditor(project), project,
		[]sitesync.Answer{{File: "content/_index.md", Whole: sitesync.Mine}})

	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Answered() error = %v, want it refused", err)
	}
	if working.resolved != nil {
		t.Error("something was written with a page left unanswered")
	}
}

func TestGivingUpPutsTheSiteBack(t *testing.T) {
	project := uuid.New()
	working := &site{status: gitsync.Status{ExchangeInProgress: true}}

	if _, err := service(t, &sources{site: working}).
		Abandon(context.Background(), anEditor(project), project); err != nil {
		t.Fatalf("Abandon() error = %v", err)
	}
	if working.abandoned != 1 {
		t.Errorf("gave up %d times, want once", working.abandoned)
	}
}

func TestOnlySomebodyAllowedTechnicalDetailSeesIt(t *testing.T) {
	project := uuid.New()
	src := &sources{
		site:      &site{},
		technical: sitesync.Technical{Line: "main", Remote: "git@example.org:a/b.git", Head: "abc123"},
	}
	made := service(t, src)

	forTheEditor, err := made.Standing(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}
	if forTheEditor.Technical != nil {
		t.Errorf("an editor was given technical detail: %+v", forTheEditor.Technical)
	}

	forTheImplementer, err := made.Standing(context.Background(), anImplementer(project), project)
	if err != nil {
		t.Fatal(err)
	}
	if forTheImplementer.Technical == nil {
		t.Fatal("an implementer was given no technical detail")
	}
	if forTheImplementer.Technical.Remote == "" || forTheImplementer.Technical.Head == "" {
		t.Errorf("Technical = %+v, want it filled in", forTheImplementer.Technical)
	}
}

func TestAViewerMayLookAndNotSettleAnything(t *testing.T) {
	project := uuid.New()
	working := &site{model: aModel(), syncResult: gitsync.SyncResult{Conflict: aConflict()}}
	made := service(t, &sources{site: working})

	if _, err := made.Standing(context.Background(), aViewer(project), project); err != nil {
		t.Errorf("Standing() as a viewer error = %v, want it allowed", err)
	}
	for name, call := range map[string]func() error{
		"bring": func() error { _, err := made.Bring(context.Background(), aViewer(project), project); return err },
		"answer": func() error {
			_, err := made.Answered(context.Background(), aViewer(project), project, nil)
			return err
		},
		"abandon": func() error { _, err := made.Abandon(context.Background(), aViewer(project), project); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, httpapi.ErrNotAllowed) {
				t.Errorf("error = %v, want ErrNotAllowed", err)
			}
		})
	}
	if working.resolved != nil || working.abandoned != 0 {
		t.Error("a viewer changed something")
	}
}

func TestSomebodyWithNoPartInTheProjectIsToldItIsNotThere(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: &site{}})
	stranger := policy.Actor{PersonID: uuid.NewString()}

	if _, err := made.Standing(context.Background(), stranger, project); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Standing() error = %v, want ErrMissing", err)
	}
}

func TestNewServiceRefusesWithNoSources(t *testing.T) {
	if _, err := sitesync.NewService(nil); !errors.Is(err, sitesync.ErrIncomplete) {
		t.Errorf("NewService() error = %v, want ErrIncomplete", err)
	}
}
