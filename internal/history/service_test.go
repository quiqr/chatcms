package history_test

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/history"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
)

const organisation = "11111111-1111-1111-1111-111111111111"

type site struct {
	versions    []gitsync.Version
	unpublished int
	changes     map[string][]gitsync.Change
	content     map[string]fstest.MapFS
	wentBackTo  string
	wentBackAs  gitsync.Identity
	message     string
	goBackErr   error
	model       *model.Model
}

func (s *site) History(ctx context.Context, limit int) ([]gitsync.Version, error) {
	return s.versions, nil
}

func (s *site) Status(ctx context.Context) (gitsync.Status, error) {
	return gitsync.Status{Unpublished: s.unpublished}, nil
}

func (s *site) ChangesIn(ctx context.Context, version string) ([]gitsync.Change, error) {
	changes, found := s.changes[version]
	if !found {
		return nil, gitsync.ErrNoSuchVersion
	}
	return changes, nil
}

func (s *site) At(ctx context.Context, version string) (fs.FS, error) {
	content, found := s.content[version]
	if !found {
		return nil, gitsync.ErrNoSuchVersion
	}
	return content, nil
}

func (s *site) Before(ctx context.Context, version string) (fs.FS, error) {
	content, found := s.content[version+"-before"]
	if !found {
		return fstest.MapFS{}, nil
	}
	return content, nil
}

func (s *site) GoBackTo(ctx context.Context, version string, author gitsync.Identity, message string) (gitsync.Version, error) {
	if s.goBackErr != nil {
		return gitsync.Version{}, s.goBackErr
	}
	s.wentBackTo, s.wentBackAs, s.message = version, author, message
	return gitsync.Version{ID: "new-version", Author: author, Message: message}, nil
}

func (s *site) ReadPublished(ctx context.Context, name string) ([]byte, error) {
	return nil, gitsync.ErrNotPublished
}

func (s *site) Model() (*model.Model, error) { return s.model, nil }

func (s *site) FS() fs.FS { return fstest.MapFS{} }

type sources struct {
	site   *site
	notes  map[string]string
	closed int
	who    gitsync.Identity
}

func (s *sources) Site(ctx context.Context, projectID uuid.UUID) (history.Site, func() error, error) {
	return s.site, func() error { s.closed++; return nil }, nil
}

func (s *sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(organisation, projectID.String()), nil
}

func (s *sources) Notes(ctx context.Context, projectID uuid.UUID) (map[string]string, error) {
	return s.notes, nil
}

func (s *sources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	if s.who.Name == "" {
		return gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, nil
	}
	return s.who, nil
}

type publisher struct {
	ran  int
	note string
	err  error
}

func (p *publisher) Run(ctx context.Context, actor policy.Actor, projectID uuid.UUID, note string) (publish.Record, error) {
	p.ran++
	p.note = note
	if p.err != nil {
		return publish.Record{}, p.err
	}
	return publish.Record{ID: uuid.New(), State: publish.Succeeded, Note: note}, nil
}

func anEditor(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleEditor},
	}
}

func aViewer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleViewer},
	}
}

func aSite() *site {
	return &site{
		versions: []gitsync.Version{
			{ID: "ccc", Time: at(20, 11), Author: gitsync.Identity{Name: "Anna"}, Message: "Not published yet", Paths: []string{"content/a.md"}},
			{ID: "bbb", Time: at(20, 9), Author: gitsync.Identity{Name: "Anna"}, Message: "Second", Paths: []string{"content/a.md"}},
			{ID: "aaa", Time: at(19, 9), Author: gitsync.Identity{Name: "Bert"}, Message: "First", Paths: []string{"content/a.md"}},
		},
		unpublished: 1,
		changes: map[string][]gitsync.Change{
			"bbb": {{Path: "content/a.md", Kind: gitsync.Edited, By: gitsync.Identity{Name: "Anna"}}},
		},
		content: map[string]fstest.MapFS{
			"bbb":        {"content/a.md": {Data: []byte("---\ntitle: Now\n---\n")}},
			"bbb-before": {"content/a.md": {Data: []byte("---\ntitle: Then\n---\n")}},
			"aaa":        {"content/a.md": {Data: []byte("---\ntitle: Then\n---\n")}},
		},
	}
}

func service(t *testing.T, src *sources, pub *publisher) *history.Service {
	t.Helper()

	made, err := history.NewService(src, pub)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func TestTheTimelineIsBuiltFromVersionsAndNotes(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(), notes: map[string]string{"bbb": "Spring prices"}}

	timeline, err := service(t, src, &publisher{}).Timeline(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("Timeline() error = %v", err)
	}

	if timeline.Unpublished != 1 {
		t.Errorf("Unpublished = %d, want 1", timeline.Unpublished)
	}
	found := false
	for _, group := range timeline.Groups {
		for _, entry := range group.Entries {
			if entry.Title == "Spring prices" {
				found = true
				if !entry.Live {
					t.Error("the version the live site holds is not marked")
				}
			}
		}
	}
	if !found {
		t.Errorf("the note was not used: %v", timeline.Groups)
	}
	if src.closed != 1 {
		t.Errorf("the site was released %d times, want once", src.closed)
	}
}

func TestTheLiveVersionIsTheMostRecentPublishedOne(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	timeline, err := service(t, src, &publisher{}).Timeline(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	for _, group := range timeline.Groups {
		for _, entry := range group.Entries {
			wantLive := entry.Version == "bbb"
			if entry.Live != wantLive {
				t.Errorf("version %q Live = %v, want %v", entry.Version, entry.Live, wantLive)
			}
		}
	}
}

func TestNothingIsLiveWhenNothingHasBeenPublished(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.unpublished = len(working.versions)

	timeline, err := service(t, &sources{site: working}, &publisher{}).
		Timeline(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	for _, group := range timeline.Groups {
		for _, entry := range group.Entries {
			if entry.Live {
				t.Errorf("version %q is marked live though nothing has been published", entry.Version)
			}
		}
	}
}

func TestWhatAVersionContainsIsSaidAsPagesAndFields(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.model = &model.Model{Singles: []model.Single{{
		Key: "a", Title: "A page", File: "content/a.md",
		Fields: []model.Field{{Key: "title", Type: "string", Title: "Page title"}},
	}}}

	summary, err := service(t, &sources{site: working}, &publisher{}).
		Contents(context.Background(), anEditor(project), project, "bbb")
	if err != nil {
		t.Fatalf("Contents() error = %v", err)
	}

	if summary.Count() != 1 {
		t.Fatalf("Contents() = %v, want one change", summary.Changes)
	}
	if summary.Changes[0].Page != "A page" {
		t.Errorf("Page = %q, want the model's title", summary.Changes[0].Page)
	}
	if len(summary.Changes[0].Fields) == 0 {
		t.Fatal("no fields were named")
	}
	if summary.Changes[0].Fields[0].Was != "Then" || summary.Changes[0].Fields[0].Now != "Now" {
		t.Errorf("the field went from %q to %q, want Then to Now",
			summary.Changes[0].Fields[0].Was, summary.Changes[0].Fields[0].Now)
	}
}

func TestAVersionThatIsNotThereIsNotFound(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &publisher{})

	if _, err := made.Contents(context.Background(), anEditor(project), project, "zzz"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Contents() error = %v, want ErrMissing", err)
	}
	if _, _, err := made.AsItWas(context.Background(), anEditor(project), project, "zzz"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("AsItWas() error = %v, want ErrMissing", err)
	}
}

func TestTheSiteAsItWasCanBeRead(t *testing.T) {
	project := uuid.New()

	asItWas, done, err := service(t, &sources{site: aSite()}, &publisher{}).
		AsItWas(context.Background(), anEditor(project), project, "aaa")
	if err != nil {
		t.Fatalf("AsItWas() error = %v", err)
	}
	defer done()

	page, err := fs.ReadFile(asItWas, "content/a.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(page) != "---\ntitle: Then\n---\n" {
		t.Errorf("the page reads %q, want it as it was", page)
	}
}

func TestGoingBackSaysWhatWillHappenFirst(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(), notes: map[string]string{"aaa": "The first update"}}

	warning, err := service(t, src, &publisher{}).
		Before(context.Background(), anEditor(project), project, "aaa")
	if err != nil {
		t.Fatalf("Before() error = %v", err)
	}

	if warning.SetAside != 1 {
		t.Errorf("SetAside = %d, want the one unpublished version", warning.SetAside)
	}
	if warning.Title != "The first update" {
		t.Errorf("Title = %q, want what that version is called", warning.Title)
	}
	if warning.When.IsZero() {
		t.Error("When is not set")
	}
}

func TestGoingBackMakesANewVersionAndCanPublishItAtOnce(t *testing.T) {
	project := uuid.New()
	working := aSite()
	pub := &publisher{}

	result, err := service(t, &sources{site: working}, pub).
		GoBackTo(context.Background(), anEditor(project), project, "aaa", true)
	if err != nil {
		t.Fatalf("GoBackTo() error = %v", err)
	}

	if working.wentBackTo != "aaa" {
		t.Errorf("went back to %q, want aaa", working.wentBackTo)
	}
	if working.wentBackAs.Name != "Anna" {
		t.Errorf("went back as %q, want the person who asked", working.wentBackAs.Name)
	}
	if result.Version != "new-version" {
		t.Errorf("Version = %q, want the new one", result.Version)
	}
	if pub.ran != 1 {
		t.Errorf("published %d times, want once", pub.ran)
	}
	if result.Published == nil {
		t.Error("the result does not carry what the publish did")
	}
}

func TestGoingBackWithoutPublishingLeavesItWaiting(t *testing.T) {
	project := uuid.New()
	pub := &publisher{}

	result, err := service(t, &sources{site: aSite()}, pub).
		GoBackTo(context.Background(), anEditor(project), project, "aaa", false)
	if err != nil {
		t.Fatal(err)
	}

	if pub.ran != 0 {
		t.Error("it published even though it was not asked to")
	}
	if result.Published != nil {
		t.Error("the result claims a publish that did not happen")
	}
}

func TestTheMessageSaysWhatHappenedInWords(t *testing.T) {
	project := uuid.New()
	working := aSite()
	src := &sources{site: working, notes: map[string]string{"aaa": "The first update"}}
	pub := &publisher{}

	if _, err := service(t, src, pub).
		GoBackTo(context.Background(), anEditor(project), project, "aaa", true); err != nil {
		t.Fatal(err)
	}

	for _, said := range []string{working.message, pub.note} {
		if said == "" {
			t.Fatal("nothing was said about what happened")
		}
		if !contains(said, "The first update") {
			t.Errorf("%q does not say which version it went back to", said)
		}
		for _, word := range []string{"commit", "revert", "reset", "branch", "SHA"} {
			if contains(said, word) {
				t.Errorf("%q carries the word %q", said, word)
			}
		}
	}
}

func contains(said, word string) bool {
	return strings.Contains(said, word)
}

func TestGoingBackToWhereTheSiteAlreadyIsSaysSo(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.goBackErr = gitsync.ErrAlreadyThere

	_, err := service(t, &sources{site: working}, &publisher{}).
		GoBackTo(context.Background(), anEditor(project), project, "aaa", false)

	if !errors.Is(err, history.ErrNothingToUndo) {
		t.Errorf("GoBackTo() error = %v, want ErrNothingToUndo", err)
	}
}

func TestOnlySomebodyAllowedMayGoBack(t *testing.T) {
	project := uuid.New()
	working := aSite()
	made := service(t, &sources{site: working}, &publisher{})

	if _, err := made.Timeline(context.Background(), aViewer(project), project); err != nil {
		t.Errorf("Timeline() as a viewer error = %v, want it allowed", err)
	}
	if _, err := made.GoBackTo(context.Background(), aViewer(project), project, "aaa", false); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("GoBackTo() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if working.wentBackTo != "" {
		t.Error("a viewer moved the site")
	}
}

func TestSomebodyWithNoPartInTheProjectIsToldItIsNotThere(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &publisher{})
	stranger := policy.Actor{PersonID: uuid.NewString()}

	for name, call := range map[string]func() error{
		"timeline": func() error { _, err := made.Timeline(context.Background(), stranger, project); return err },
		"contents": func() error { _, err := made.Contents(context.Background(), stranger, project, "bbb"); return err },
		"go back": func() error {
			_, err := made.GoBackTo(context.Background(), stranger, project, "aaa", false)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, httpapi.ErrMissing) {
				t.Errorf("error = %v, want ErrMissing", err)
			}
		})
	}
}

func TestGoingBackToAVersionTheSiteDoesNotHaveIsNotFound(t *testing.T) {
	project := uuid.New()

	_, err := service(t, &sources{site: aSite()}, &publisher{}).
		GoBackTo(context.Background(), anEditor(project), project, "zzz", false)

	if !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("GoBackTo() error = %v, want ErrMissing", err)
	}
}

func TestAFailedPublishStillReportsThatTheSiteWentBack(t *testing.T) {
	project := uuid.New()
	pub := &publisher{err: errors.New("the remote is unreachable")}

	result, err := service(t, &sources{site: aSite()}, pub).
		GoBackTo(context.Background(), anEditor(project), project, "aaa", true)

	if err == nil {
		t.Fatal("GoBackTo() error = nil, want the publish failure reported")
	}
	if result.Version == "" {
		t.Error("the result does not say that the site did go back")
	}
}

func TestNewServiceRefusesAMissingPart(t *testing.T) {
	if _, err := history.NewService(nil, &publisher{}); !errors.Is(err, history.ErrIncomplete) {
		t.Errorf("NewService() with no sources error = %v, want ErrIncomplete", err)
	}
	if _, err := history.NewService(&sources{}, nil); !errors.Is(err, history.ErrIncomplete) {
		t.Errorf("NewService() with no publisher error = %v, want ErrIncomplete", err)
	}
}
