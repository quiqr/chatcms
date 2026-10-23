package content_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/content"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
)

func TestSlugIsWhatTheOldQuiqrWouldName(t *testing.T) {
	cases := map[string]string{
		"Spring prices":       "spring-prices",
		"  Spring   prices  ": "spring-prices",
		"Café au lait":        "cafe-au-lait",
		"Grüße aus Köln":      "grusse-aus-koln",
		"Prijzen 2026":        "prijzen-2026",
		"A/B test":            "a-b-test",
		"Hello!!!":            "hello",
		"---":                 "",
		"":                    "",
		"ÉÉN":                 "een",
	}

	for said, want := range cases {
		if got := content.Slug(said); got != want {
			t.Errorf("Slug(%q) = %q, want %q", said, got, want)
		}
	}
}

func sectionNamed(sections []content.Section, key string) (content.Section, bool) {
	for _, section := range sections {
		if section.Key == key {
			return section, true
		}
	}
	return content.Section{}, false
}

func TestTheBrowserListsPagesAndCollections(t *testing.T) {
	project := uuid.New()

	sections, err := service(t, &sources{site: aSite()}, &syncer{}).
		List(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	pages, found := sectionNamed(sections, "pages")
	if !found || len(pages.Items) != 1 || pages.Items[0].Title != "Home page" {
		t.Fatalf("pages = %+v", pages)
	}

	news, found := sectionNamed(sections, "news")
	if !found {
		t.Fatalf("sections = %+v", sections)
	}
	if news.Title != "News" || news.Count != 1 || news.Items[0].Key != "news/one" {
		t.Errorf("news = %+v", news)
	}
	if news.ItemName == "" {
		t.Error("a collection has no word for one of its items")
	}
}

func TestAnItemWaitingToBePublishedIsMarked(t *testing.T) {
	project := uuid.New()
	working := aSite()
	working.changes = []gitsync.Change{{
		Path: "content/news/one.md",
		Kind: gitsync.Edited,
		By:   gitsync.Identity{Name: "Anna"},
		At:   time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC),
	}}

	sections, err := service(t, &sources{site: working}, &syncer{}).
		List(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatal(err)
	}

	news, _ := sectionNamed(sections, "news")
	if !news.Items[0].Waiting {
		t.Error("a page with an unpublished change is not marked")
	}
	if news.Items[0].By != "Anna" {
		t.Errorf("By = %q, want who changed it", news.Items[0].By)
	}

	pages, _ := sectionNamed(sections, "pages")
	if pages.Items[0].Waiting {
		t.Error("a page with nothing waiting is marked anyway")
	}
}

func TestAPageIsMade(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	sync := &syncer{}

	page, err := service(t, src, sync).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "news", Title: "Spring prices"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if page.Key != "news/spring-prices" {
		t.Errorf("Key = %q, want it named from the title", page.Key)
	}
	written, wrote := src.site.saved[0].Writes["content/news/spring-prices.md"]
	if !wrote {
		t.Fatalf("wrote %v", src.site.saved[0].Writes)
	}
	if string(written) != "---\ntitle: Spring prices\n---\n\n" {
		t.Errorf("the new page holds %q", written)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times, want once", sync.asked)
	}
}

func TestAPageCannotBeMadeWhereThereIsOne(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	_, err := service(t, src, &syncer{}).Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "news", Title: "One"})

	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Create() error = %v, want it refused", err)
	}
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("error = %q, want it to say why", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("something was written anyway")
	}
}

func TestAPageNeedsAName(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &syncer{})

	for _, title := range []string{"", "   ", "---", "!!!"} {
		if _, err := made.Create(context.Background(), anEditor(project), project,
			content.NewPage{Section: "news", Title: title}); !errors.Is(err, httpapi.ErrRefused) {
			t.Errorf("Create(%q) error = %v, want it refused", title, err)
		}
	}
}

func TestOnlyACollectionCanGrow(t *testing.T) {
	project := uuid.New()
	made := service(t, &sources{site: aSite()}, &syncer{})

	for _, section := range []string{"pages", "nothing", ""} {
		if _, err := made.Create(context.Background(), anEditor(project), project,
			content.NewPage{Section: section, Title: "Something"}); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("Create(section %q) error = %v, want ErrMissing", section, err)
		}
	}
}

func TestAPageIsRenamedAndItsTitleFollows(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	page, err := service(t, src, &syncer{}).Rename(context.Background(), anEditor(project), project,
		"news/one", "The very first")
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if page.Key != "news/the-very-first" {
		t.Errorf("Key = %q", page.Key)
	}
	save := src.site.saved[0]
	if _, wrote := save.Writes["content/news/the-very-first.md"]; !wrote {
		t.Errorf("wrote %v", save.Writes)
	}
	if len(save.Deletes) != 1 || save.Deletes[0] != "content/news/one.md" {
		t.Errorf("deleted %v, want the old file", save.Deletes)
	}
	if !strings.Contains(string(save.Writes["content/news/the-very-first.md"]), "title: The very first") {
		t.Error("the title inside the page did not follow its name")
	}
}

func TestRenamingToTheSameNameChangesNothing(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	if _, err := service(t, src, &syncer{}).Rename(context.Background(), anEditor(project), project,
		"news/one", "One"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("renaming to the same name wrote something")
	}
}

func TestAPageIsDuplicatedAndTheFirstStays(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}

	page, err := service(t, src, &syncer{}).Duplicate(context.Background(), anEditor(project), project,
		"news/one", "One again")
	if err != nil {
		t.Fatalf("Duplicate() error = %v", err)
	}

	if page.Key != "news/one-again" {
		t.Errorf("Key = %q", page.Key)
	}
	if len(src.site.saved[0].Deletes) != 0 {
		t.Error("duplicating removed the page it copied")
	}
	if _, still := src.site.files["content/news/one.md"]; !still {
		t.Error("the page that was copied is gone")
	}
}

func TestAPageIsDeleted(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	sync := &syncer{}

	if err := service(t, src, sync).Delete(context.Background(), anEditor(project), project, "news/one"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	save := src.site.saved[0]
	if len(save.Deletes) != 1 || save.Deletes[0] != "content/news/one.md" {
		t.Errorf("deleted %v", save.Deletes)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times, want once", sync.asked)
	}
}

func TestAPageTheModelNamesCannotBeRenamedOrDeleted(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	made := service(t, src, &syncer{})

	if _, err := made.Rename(context.Background(), anEditor(project), project, "home", "Welcome"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("Rename() of a single error = %v, want it refused", err)
	}
	if err := made.Delete(context.Background(), anEditor(project), project, "home"); !errors.Is(err, httpapi.ErrRefused) {
		t.Errorf("Delete() of a single error = %v, want it refused", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("the site was changed anyway")
	}
}

func TestAViewerMayBrowseButNotChangeAnything(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	made := service(t, src, &syncer{})

	if _, err := made.List(context.Background(), aViewer(project), project); err != nil {
		t.Errorf("List() as a viewer error = %v, want it allowed", err)
	}

	for name, call := range map[string]func() error{
		"create": func() error {
			_, err := made.Create(context.Background(), aViewer(project), project, content.NewPage{Section: "news", Title: "x"})
			return err
		},
		"rename": func() error {
			_, err := made.Rename(context.Background(), aViewer(project), project, "news/one", "x")
			return err
		},
		"duplicate": func() error {
			_, err := made.Duplicate(context.Background(), aViewer(project), project, "news/one", "x")
			return err
		},
		"delete": func() error {
			return made.Delete(context.Background(), aViewer(project), project, "news/one")
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, httpapi.ErrNotAllowed) {
				t.Errorf("error = %v, want ErrNotAllowed", err)
			}
		})
	}
	if len(src.site.saved) != 0 {
		t.Error("a viewer changed the site")
	}
}

func TestANameCannotReachOutOfTheCollection(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite()}
	made := service(t, src, &syncer{})

	// A title that would become a path is slugged into something harmless
	// rather than refused, because a person typing a slash means a slash in
	// words, not a folder.
	page, err := made.Create(context.Background(), anEditor(project), project,
		content.NewPage{Section: "news", Title: "../../etc/passwd"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if strings.Contains(page.Key, "..") {
		t.Errorf("Key = %q, want nothing that climbs", page.Key)
	}
	for name := range src.site.saved[0].Writes {
		if !strings.HasPrefix(name, "content/news/") {
			t.Errorf("wrote %q, want it inside the collection", name)
		}
	}
}
