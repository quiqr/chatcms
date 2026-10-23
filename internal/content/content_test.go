package content_test

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/content"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
)

const organisation = "11111111-1111-1111-1111-111111111111"

func aModel() *model.Model {
	return &model.Model{
		Singles: []model.Single{{
			Key:   "home",
			Title: "Home page",
			File:  "content/_index.md",
			Fields: []model.Field{
				{Key: "title", Type: "string", Title: "Page title"},
				{Key: "intro", Type: "string", Title: "Introduction",
					Config: model.StringConfig{MultiLine: true}},
				{Key: "published", Type: "boolean", Title: "Published"},
				{Key: "colour", Type: "select", Title: "Colour", Config: model.SelectConfig{
					Options: []model.Option{{Value: "red", Text: "Red"}, {Value: "blue"}},
				}},
				{Key: "seo", Type: "section", Title: "Search", Fields: []model.Field{
					{Key: "description", Type: "string", Title: "Description"},
				}},
				{Key: "tags", Type: "leaf-array", Title: "Tags"},
				{Key: "links", Type: "accordion", Title: "Links", Fields: []model.Field{
					{Key: "label", Type: "string", Title: "Label"},
					{Key: "url", Type: "string", Title: "Address"},
				}},
				{Key: "secret", Type: "hidden"},
				{Key: "note", Type: "info", Content: "Keep it short."},
				{Key: "mystery", Type: "eisenhouwer", Title: "Mystery"},
				{Key: "mainContent", Type: "markdown", Title: "Page text"},
			},
		}},
		Collections: []model.Collection{{
			Key: "news", Title: "News", Folder: "content/news",
			Fields: []model.Field{{Key: "title", Type: "string", Title: "Headline"}},
		}},
	}
}

const theHomePage = `---
title: Welcome
intro: We sell apples and pears.
published: true
colour: red
seo:
  description: A bakery.
tags:
  - bread
  - cake
links:
  - label: Shop
    url: /shop/
mystery: something
---

The body of the page.
`

type site struct {
	files    fstest.MapFS
	saved    []gitsync.Save
	model    *model.Model
	modelErr error
	saveErr  error
}

func (s *site) FS() fs.FS { return s.files }

func (s *site) Save(ctx context.Context, save gitsync.Save) (gitsync.Version, error) {
	if s.saveErr != nil {
		return gitsync.Version{}, s.saveErr
	}
	s.saved = append(s.saved, save)
	for name, content := range save.Writes {
		s.files[name] = &fstest.MapFile{Data: content}
	}
	return gitsync.Version{ID: "a-version"}, nil
}

func (s *site) Model() (*model.Model, error) {
	if s.modelErr != nil {
		return nil, s.modelErr
	}
	if s.model == nil {
		return nil, model.ErrNoModel
	}
	return s.model, nil
}

type sources struct {
	site   *site
	closed int
}

func (s *sources) Site(ctx context.Context, projectID uuid.UUID) (content.Site, func() error, error) {
	return s.site, func() error { s.closed++; return nil }, nil
}

func (s *sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(organisation, projectID.String()), nil
}

func (s *sources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	return gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, nil
}

type syncer struct{ asked int }

func (s *syncer) SyncSoon(ctx context.Context, projectID uuid.UUID) error {
	s.asked++
	return nil
}

func aSite() *site {
	return &site{
		files: fstest.MapFS{
			"content/_index.md":   {Data: []byte(theHomePage)},
			"content/news/one.md": {Data: []byte("---\ntitle: One\n---\n\nFirst.\n")},
		},
		model: aModel(),
	}
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

func service(t *testing.T, src *sources, sync *syncer) *content.Service {
	t.Helper()

	made, err := content.NewService(src, sync)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func controlAt(controls []content.Control, path string) (content.Control, bool) {
	for _, control := range controls {
		if control.Path == path {
			return control, true
		}
		if found, ok := controlAt(control.Controls, path); ok {
			return found, true
		}
	}
	return content.Control{}, false
}

func TestAPageIsReadAsAForm(t *testing.T) {
	project := uuid.New()

	page, err := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if page.Title != "Home page" || page.Section != "Pages" {
		t.Errorf("page = %+v", page)
	}
	if page.Body != "The body of the page.\n" {
		t.Errorf("Body = %q", page.Body)
	}
	if !page.HasBody {
		t.Error("a Markdown page says it has no text of its own")
	}

	title, found := controlAt(page.Controls, "title")
	if !found {
		t.Fatalf("controls = %+v", page.Controls)
	}
	if title.Title != "Page title" || title.Value != "Welcome" {
		t.Errorf("title control = %+v", title)
	}
}

func TestEveryKindOfFieldComesBackWithItsValue(t *testing.T) {
	project := uuid.New()
	page, err := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]any{
		"intro":           "We sell apples and pears.",
		"published":       true,
		"colour":          "red",
		"seo.description": "A bakery.",
	}
	for path, want := range cases {
		control, found := controlAt(page.Controls, path)
		if !found {
			t.Errorf("%s is missing", path)
			continue
		}
		if control.Value != want {
			t.Errorf("%s = %v, want %v", path, control.Value, want)
		}
	}

	tags, _ := controlAt(page.Controls, "tags")
	list, ok := tags.Value.([]any)
	if !ok || len(list) != 2 || list[0] != "bread" {
		t.Errorf("tags = %#v, want the two it has", tags.Value)
	}

	links, _ := controlAt(page.Controls, "links")
	items, ok := links.Value.([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("links = %#v", links.Value)
	}
	first, ok := items[0].(map[string]any)
	if !ok || first["label"] != "Shop" {
		t.Errorf("the first link is %#v", items[0])
	}
	if len(links.Controls) != 2 {
		t.Errorf("a list of items carries %d controls for one item, want two", len(links.Controls))
	}
}

func TestAGroupCarriesTheFieldsInsideIt(t *testing.T) {
	project := uuid.New()
	page, _ := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	group, found := controlAt(page.Controls, "seo")
	if !found {
		t.Fatal("the group is missing")
	}
	if len(group.Controls) != 1 || group.Controls[0].Path != "seo.description" {
		t.Errorf("group = %+v", group)
	}
	if group.Value != nil {
		t.Error("a group carries a value of its own")
	}
}

func TestAFieldThisVersionCannotDrawIsStillShown(t *testing.T) {
	project := uuid.New()
	page, _ := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	mystery, found := controlAt(page.Controls, "mystery")
	if !found {
		t.Fatal("a field of an unknown kind was left out altogether")
	}
	if !mystery.Unknown {
		t.Error("a field of an unknown kind does not say so")
	}
	if mystery.Value != "something" {
		t.Errorf("Value = %v, want what the file holds", mystery.Value)
	}
}

func TestAHiddenFieldIsNotDrawn(t *testing.T) {
	project := uuid.New()
	page, _ := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	if _, found := controlAt(page.Controls, "secret"); found {
		t.Error("a hidden field was drawn")
	}
}

func TestSelectOptionsCarryALabel(t *testing.T) {
	project := uuid.New()
	page, _ := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	colour, _ := controlAt(page.Controls, "colour")
	if len(colour.Options) != 2 {
		t.Fatalf("options = %+v", colour.Options)
	}
	if colour.Options[0].Label != "Red" {
		t.Errorf("the first option reads %q", colour.Options[0].Label)
	}
	if colour.Options[1].Label != "blue" {
		t.Errorf("an option with no words of its own reads %q, want its value", colour.Options[1].Label)
	}
}

func TestAnEditIsSavedAndASyncIsAskedFor(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	sync := &syncer{}

	page, err := service(t, src, sync).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Values: map[string]any{"title": "Welcome home"}})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	title, _ := controlAt(page.Controls, "title")
	if title.Value != "Welcome home" {
		t.Errorf("the page came back with %v", title.Value)
	}
	if len(src.site.saved) != 1 {
		t.Fatalf("saved %d times, want once", len(src.site.saved))
	}
	if src.site.saved[0].Author.Name != "Anna" {
		t.Errorf("saved as %q, want the person who edited", src.site.saved[0].Author.Name)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times, want once", sync.asked)
	}
}

func TestSavingChangesNothingItWasNotAskedTo(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	if _, err := service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Values: map[string]any{"title": "Welcome home"}}); err != nil {
		t.Fatal(err)
	}

	written := string(src.site.saved[0].Writes["content/_index.md"])
	before := strings.ReplaceAll(theHomePage, "title: Welcome\n", "title: Welcome home\n")
	if written != before {
		t.Errorf("the file was written as\n%q\nwant\n%q", written, before)
	}
}

func TestSavingTheBodyLeavesTheFieldsAlone(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	body := "A new body.\n"

	if _, err := service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Body: &body}); err != nil {
		t.Fatal(err)
	}

	written := string(src.site.saved[0].Writes["content/_index.md"])
	if !strings.Contains(written, "title: Welcome\n") {
		t.Error("saving the body changed the fields")
	}
	if !strings.HasSuffix(written, "A new body.\n") {
		t.Errorf("the file ends %q", written[len(written)-40:])
	}
}

func TestClearingAFieldWritesTheAbsentValueTheReferenceWrites(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	if _, err := service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Values: map[string]any{"colour": nil}}); err != nil {
		t.Fatal(err)
	}

	written := string(src.site.saved[0].Writes["content/_index.md"])
	if !strings.Contains(written, "colour: null") {
		t.Errorf("the cleared field was written as\n%s\nwant the absent value the old Quiqr writes", written)
	}
}

func TestAPageWhoseModelGivesItsTextNoFieldIsNotEdited(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	src.site.model.Singles[0].Fields = src.site.model.Singles[0].Fields[:len(src.site.model.Singles[0].Fields)-1]

	page, err := service(t, src, &syncer{}).Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatal(err)
	}
	if page.HasBody {
		t.Error("a page whose model gives its text no field says it can be edited")
	}
	if page.Body != "The body of the page.\n" {
		t.Errorf("the text is still shown as %q, want it readable even when it cannot be edited", page.Body)
	}

	body := "Something else.\n"
	_, err = service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Body: &body})
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Save() error = %v, want it refused rather than quietly ignored", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("the text was written anyway")
	}
}

func TestAFieldThePageDoesNotHaveIsRefused(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	_, err := service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "home",
		content.Edit{Values: map[string]any{"somethingElse": "x"}})

	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Save() error = %v, want it refused", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("something was saved anyway")
	}
}

func TestACollectionItemIsReadAndSaved(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	page, err := service(t, src, &syncer{}).
		Read(context.Background(), anEditor(project), project, "news/one")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if page.Section != "News" || page.Title != "One" {
		t.Errorf("page = %+v", page)
	}

	if _, err := service(t, src, &syncer{}).Save(context.Background(), anEditor(project), project, "news/one",
		content.Edit{Values: map[string]any{"title": "The first"}}); err != nil {
		t.Fatal(err)
	}
	if _, wrote := src.site.saved[0].Writes["content/news/one.md"]; !wrote {
		t.Errorf("wrote %v, want the item's own file", src.site.saved[0].Writes)
	}
}

func TestAPageNameCannotReachOutOfItsCollection(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &syncer{})

	for _, key := range []string{
		"news/../../etc/passwd",
		"news/..",
		"news/",
		`news/a\b`,
		"news/a/b",
		"nothing",
		"",
	} {
		if _, err := made.Read(context.Background(), anEditor(project), project, key); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("Read(%q) error = %v, want ErrMissing", key, err)
		}
	}
}

func TestAViewerMayReadButNotEdit(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	made := service(t, src, &syncer{})

	if _, err := made.Read(context.Background(), aViewer(project), project, "home"); err != nil {
		t.Errorf("Read() as a viewer error = %v, want it allowed", err)
	}
	if _, err := made.Save(context.Background(), aViewer(project), project, "home",
		content.Edit{Values: map[string]any{"title": "x"}}); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Save() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("a viewer changed the site")
	}
}

func TestSomebodyWithNoPartInTheProjectIsToldItIsNotThere(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &syncer{})
	stranger := policy.Actor{PersonID: uuid.NewString()}

	if _, err := made.Read(context.Background(), stranger, project, "home"); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Read() error = %v, want ErrMissing", err)
	}
	if _, err := made.Save(context.Background(), stranger, project, "home", content.Edit{}); !errors.Is(err, httpapi.ErrMissing) {
		t.Errorf("Save() error = %v, want ErrMissing", err)
	}
}

func TestAModelWithAMistakeSaysWhereTheMistakeIs(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.modelErr = errors.New("quiqr/model/base.yaml: singles[0].fields[5].field: expected the single field every item of this array has")

	_, err := service(t, &sources{site: working}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Read() error = %v, want it refused with a reason", err)
	}
	if !strings.Contains(err.Error(), "base.yaml") {
		t.Errorf("error = %q, want it to name the file the mistake is in", err)
	}
	if errors.Is(err, content.ErrNoModel) {
		t.Error("a model with a mistake was reported as no model at all")
	}
}

func TestASiteWithNoModelSaysSo(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.model = nil

	_, err := service(t, &sources{site: working}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")

	if !errors.Is(err, content.ErrNoModel) {
		t.Errorf("Read() error = %v, want ErrNoModel", err)
	}
}

func TestTheSiteIsReleasedEveryTime(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	made := service(t, src, &syncer{})

	made.Read(context.Background(), anEditor(project), project, "home")
	made.Save(context.Background(), anEditor(project), project, "home", content.Edit{})

	if src.closed != 2 {
		t.Errorf("the site was released %d times, want once for each call", src.closed)
	}
}

func TestNewServiceRefusesWithNoSources(t *testing.T) {
	if _, err := content.NewService(nil, &syncer{}); !errors.Is(err, content.ErrIncomplete) {
		t.Errorf("NewService() error = %v, want ErrIncomplete", err)
	}
	if _, err := content.NewService(&sources{}, nil); err != nil {
		t.Errorf("NewService() with no syncer error = %v, want it allowed", err)
	}
}

func TestThePagesOwnTextIsNotDrawnTwice(t *testing.T) {
	project := uuid.New()

	page, err := service(t, &sources{site: aSite()}, &syncer{}).
		Read(context.Background(), anEditor(project), project, "home")
	if err != nil {
		t.Fatal(err)
	}

	if !page.HasBody {
		t.Fatal("this model declares a field for the page's text")
	}
	if _, found := controlAt(page.Controls, "mainContent"); found {
		t.Error("the page's own text is drawn as an ordinary field as well as by its own editor")
	}
	if page.Body == "" {
		t.Error("the page's own text is not there at all")
	}
}
