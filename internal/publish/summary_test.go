package publish_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/publish"
)

// site stands in for a working copy: what differs, what the files say now,
// and what they said when they were published.
type site struct {
	differences []gitsync.Change
	published   map[string]string
	now         fstest.MapFS
	model       *model.Model
}

func (s *site) Changes(ctx context.Context) ([]gitsync.Change, error) {
	return s.differences, nil
}

func (s *site) ReadPublished(ctx context.Context, name string) ([]byte, error) {
	content, found := s.published[name]
	if !found {
		return nil, gitsync.ErrNotPublished
	}
	return []byte(content), nil
}

func (s *site) FS() fs.FS { return s.now }

func (s *site) Model() (*model.Model, error) { return s.model, nil }

func aModel() *model.Model {
	return &model.Model{
		Singles: []model.Single{{
			Key:   "home",
			Title: "Home page",
			File:  "content/_index.md",
			Fields: []model.Field{
				{Key: "title", Type: "string", Title: "Page title"},
				{Key: "intro", Type: "markdown", Title: "Introduction"},
				{Key: "heroImage", Type: "image"},
			},
		}},
		Collections: []model.Collection{{
			Key:    "news",
			Title:  "News",
			Folder: "content/news",
			Fields: []model.Field{
				{Key: "title", Type: "string", Title: "Headline"},
			},
		}},
	}
}

func anEditedHomePage() *site {
	return &site{
		differences: []gitsync.Change{{
			Path: "content/_index.md",
			Kind: gitsync.Edited,
			By:   gitsync.Identity{Name: "Anna"},
			At:   time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC),
		}},
		published: map[string]string{
			"content/_index.md": "---\ntitle: Welcome\nintro: We sell apples and pears.\n---\n\nThe old body.\n",
		},
		now: fstest.MapFS{
			"content/_index.md": {Data: []byte("---\ntitle: Welcome\nintro: We sell plums and pears.\n---\n\nThe new body.\n")},
		},
		model: aModel(),
	}
}

func TestAChangedPageIsDescribedByItsFields(t *testing.T) {
	summary, err := publish.Describe(context.Background(), anEditedHomePage(), aModel())
	if err != nil {
		t.Fatalf("Describe() error = %v", err)
	}

	if summary.Count() != 1 {
		t.Fatalf("Describe() = %v, want one change", summary.Changes)
	}
	change := summary.Changes[0]

	if change.Kind != publish.ChangedPage {
		t.Errorf("Kind = %q, want %q", change.Kind, publish.ChangedPage)
	}
	if change.Page != "Home page" {
		t.Errorf("Page = %q, want the title the model gives it", change.Page)
	}
	if change.By != "Anna" {
		t.Errorf("By = %q, want who made it", change.By)
	}

	named := map[string]publish.FieldChange{}
	for _, field := range change.Fields {
		named[field.Name] = field
	}
	if _, found := named["Introduction"]; !found {
		t.Errorf("fields = %v, want the model's name for the one that changed", change.Fields)
	}
	if _, found := named[publish.BodyField]; !found {
		t.Errorf("fields = %v, want the page's own text listed", change.Fields)
	}
	if _, found := named["Page title"]; found {
		t.Error("a field that did not change was listed")
	}
}

func TestAChangedFieldCarriesAWordDiff(t *testing.T) {
	summary, _ := publish.Describe(context.Background(), anEditedHomePage(), aModel())

	for _, field := range summary.Changes[0].Fields {
		if field.Name != "Introduction" {
			continue
		}
		if field.Was != "We sell apples and pears." || field.Now != "We sell plums and pears." {
			t.Errorf("Was = %q, Now = %q", field.Was, field.Now)
		}
		if len(field.Words) == 0 {
			t.Error("the field carries no word diff")
		}
		return
	}
	t.Fatal("the changed field was not found")
}

func TestEveryKindOfChangeIsRecognised(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{
			{Path: "content/_index.md", Kind: gitsync.Edited},
			{Path: "content/news/launch.md", Kind: gitsync.Added},
			{Path: "content/news/old.md", Kind: gitsync.Removed},
			{Path: "static/images/logo.png", Kind: gitsync.Added},
			{Path: "static/images/old.jpg", Kind: gitsync.Removed},
			{Path: "layouts/index.html", Kind: gitsync.Edited},
		},
		published: map[string]string{"content/_index.md": "---\ntitle: Welcome\n---\n"},
		now: fstest.MapFS{
			"content/_index.md": {Data: []byte("---\ntitle: Welcome home\n---\n")},
		},
		model: aModel(),
	}

	summary, err := publish.Describe(context.Background(), working, aModel())
	if err != nil {
		t.Fatal(err)
	}

	kinds := map[string]publish.Kind{}
	for _, change := range summary.Changes {
		kinds[change.Path] = change.Kind
	}
	want := map[string]publish.Kind{
		"content/_index.md":      publish.ChangedPage,
		"content/news/launch.md": publish.NewPage,
		"content/news/old.md":    publish.RemovedPage,
		"static/images/logo.png": publish.NewImage,
		"static/images/old.jpg":  publish.RemovedImage,
		"layouts/index.html":     publish.OtherFile,
	}
	for path, kind := range want {
		if kinds[path] != kind {
			t.Errorf("%s = %q, want %q", path, kinds[path], kind)
		}
	}
}

func TestAPageInACollectionIsNamedByItsSection(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{{Path: "content/news/the-launch.md", Kind: gitsync.Added}},
		now:         fstest.MapFS{"content/news/the-launch.md": {Data: []byte("---\ntitle: The launch\n---\n")}},
		model:       aModel(),
	}

	summary, _ := publish.Describe(context.Background(), working, aModel())

	if summary.Changes[0].Section != "News" {
		t.Errorf("Section = %q, want the collection's title", summary.Changes[0].Section)
	}
	if summary.Changes[0].Page != "The launch" {
		t.Errorf("Page = %q, want the page's own name in words", summary.Changes[0].Page)
	}
}

func TestAFileTheModelSaysNothingAboutIsStillListed(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{{Path: "assets/js/search.js", Kind: gitsync.Edited}},
		now:         fstest.MapFS{"assets/js/search.js": {Data: []byte("// code")}},
	}

	summary, err := publish.Describe(context.Background(), working, nil)
	if err != nil {
		t.Fatal(err)
	}

	if summary.Count() != 1 {
		t.Fatalf("Describe() = %v, want the file to be listed anyway", summary.Changes)
	}
	if summary.Changes[0].Page == "" {
		t.Error("the file was listed with no name")
	}
	if summary.Changes[0].Kind != publish.OtherFile {
		t.Errorf("Kind = %q, want %q", summary.Changes[0].Kind, publish.OtherFile)
	}
}

func TestAnIndexFileIsNamedAfterItsFolder(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{{Path: "content/about-us/index.md", Kind: gitsync.Added}},
		now:         fstest.MapFS{"content/about-us/index.md": {Data: []byte("---\ntitle: About\n---\n")}},
	}

	summary, _ := publish.Describe(context.Background(), working, nil)

	if summary.Changes[0].Page != "About us" {
		t.Errorf("Page = %q, want it named after the folder", summary.Changes[0].Page)
	}
}

func TestNothingToPublishIsAnEmptySummary(t *testing.T) {
	summary, err := publish.Describe(context.Background(), &site{now: fstest.MapFS{}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count() != 0 {
		t.Errorf("Describe() = %v, want nothing", summary.Changes)
	}
}

func TestAPageThatWasNeverPublishedComparesAgainstNothing(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{{Path: "content/_index.md", Kind: gitsync.Edited}},
		now:         fstest.MapFS{"content/_index.md": {Data: []byte("---\ntitle: Welcome\n---\n\nBody.\n")}},
		model:       aModel(),
	}

	summary, err := publish.Describe(context.Background(), working, aModel())
	if err != nil {
		t.Fatal(err)
	}

	if len(summary.Changes[0].Fields) == 0 {
		t.Fatal("no fields were listed for a page with no published version")
	}
	for _, field := range summary.Changes[0].Fields {
		if field.Was != "" {
			t.Errorf("%s was %q, want nothing before it existed", field.Name, field.Was)
		}
	}
}

func TestANestedFieldKeepsItsParentsName(t *testing.T) {
	nested := &model.Model{Singles: []model.Single{{
		Key:   "contact",
		Title: "Contact",
		File:  "content/contact.md",
		Fields: []model.Field{{
			Key: "address", Type: "section", Title: "Address",
			Fields: []model.Field{{Key: "street", Type: "string", Title: "Street"}},
		}},
	}}}

	working := &site{
		differences: []gitsync.Change{{Path: "content/contact.md", Kind: gitsync.Edited}},
		published:   map[string]string{"content/contact.md": "---\naddress:\n  street: Oude Gracht 1\n---\n"},
		now:         fstest.MapFS{"content/contact.md": {Data: []byte("---\naddress:\n  street: Nieuwe Gracht 2\n---\n")}},
		model:       nested,
	}

	summary, _ := publish.Describe(context.Background(), working, nested)

	if len(summary.Changes[0].Fields) != 1 {
		t.Fatalf("fields = %v, want one", summary.Changes[0].Fields)
	}
	if summary.Changes[0].Fields[0].Name != "Address, Street" {
		t.Errorf("Name = %q, want the parent in front", summary.Changes[0].Fields[0].Name)
	}
}

func TestAFieldWithNoTitleIsNamedFromItsKey(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{{Path: "content/_index.md", Kind: gitsync.Edited}},
		published:   map[string]string{"content/_index.md": "---\nheroImage: old.png\n---\n"},
		now:         fstest.MapFS{"content/_index.md": {Data: []byte("---\nheroImage: new.png\n---\n")}},
		model:       aModel(),
	}

	summary, _ := publish.Describe(context.Background(), working, aModel())

	if summary.Changes[0].Fields[0].Name != "Hero image" {
		t.Errorf("Name = %q, want the key turned into words", summary.Changes[0].Fields[0].Name)
	}
}

func TestChangesAreGroupedAndOrdered(t *testing.T) {
	working := &site{
		differences: []gitsync.Change{
			{Path: "content/news/zebra.md", Kind: gitsync.Added},
			{Path: "content/_index.md", Kind: gitsync.Edited},
			{Path: "content/news/apple.md", Kind: gitsync.Added},
		},
		now: fstest.MapFS{
			"content/_index.md":     {Data: []byte("---\ntitle: Home\n---\n")},
			"content/news/apple.md": {Data: []byte("---\ntitle: Apple\n---\n")},
			"content/news/zebra.md": {Data: []byte("---\ntitle: Zebra\n---\n")},
		},
		model: aModel(),
	}

	summary, _ := publish.Describe(context.Background(), working, aModel())

	sections := []string{}
	for _, change := range summary.Changes {
		sections = append(sections, change.Section+"/"+change.Page)
	}
	want := []string{"News/Apple", "News/Zebra", "Pages/Home page"}
	for index, entry := range want {
		if sections[index] != entry {
			t.Errorf("change %d = %q, want %q (all: %v)", index, sections[index], entry, sections)
		}
	}
}
