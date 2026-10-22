package publish_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/publish"
)

type fakeRepo struct {
	publishedTo    string
	publishedPages fs.FS
	pagesLine      string
	author         gitsync.Identity
	message        string
	closed         int
	err            error
	gitsync.Repo
}

func (r *fakeRepo) PublishTo(ctx context.Context, line string) (gitsync.PublishResult, error) {
	r.publishedTo = "the site's own files to " + line
	return gitsync.PublishResult{}, r.err
}

func (r *fakeRepo) PublishPages(ctx context.Context, line string, pages fs.FS, author gitsync.Identity, message string) (gitsync.PublishResult, error) {
	r.publishedPages, r.pagesLine, r.author, r.message = pages, line, author, message
	return gitsync.PublishResult{}, r.err
}

func (r *fakeRepo) Close() error { r.closed++; return nil }

type fakeSites struct {
	repo *fakeRepo
	err  error
}

func (s *fakeSites) Open(ctx context.Context, id string) (gitsync.Repo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.repo, nil
}

func publisher(t *testing.T, sites *fakeSites, target publish.Target) *publish.Publisher {
	t.Helper()

	made, err := publish.NewPublisher(sites, func(ctx context.Context, projectID uuid.UUID) (publish.Target, error) {
		return target, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func TestAProjectThatPublishesItsSourceSendsTheSite(t *testing.T) {
	repo := &fakeRepo{}
	made := publisher(t, &fakeSites{repo: repo}, publish.Target{Line: "quiqr/publish"})

	if err := made.Publish(context.Background(), uuid.New(), nil); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if repo.publishedTo != "the site's own files to quiqr/publish" {
		t.Errorf("published %q", repo.publishedTo)
	}
	if repo.publishedPages != nil {
		t.Error("built pages were sent for a project that publishes its source")
	}
	if repo.closed != 1 {
		t.Errorf("the working copy was released %d times, want once", repo.closed)
	}
}

func TestAProjectThatPublishesBuiltPagesSendsThem(t *testing.T) {
	repo := &fakeRepo{}
	made := publisher(t, &fakeSites{repo: repo}, publish.Target{
		Line:       "live",
		BuiltPages: true,
		Author:     gitsync.Identity{Name: "Anna", Email: "anna@example.org"},
		Message:    "Spring prices",
	})

	pages := fstest.MapFS{"index.html": {Data: []byte("<h1>Home</h1>")}}
	if err := made.Publish(context.Background(), uuid.New(), pages); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if repo.pagesLine != "live" {
		t.Errorf("published to %q, want the line the project names", repo.pagesLine)
	}
	if repo.author.Name != "Anna" {
		t.Errorf("author = %q, want the person who pressed the button", repo.author.Name)
	}
	if repo.message != "Spring prices" {
		t.Errorf("message = %q, want the note", repo.message)
	}
	if repo.publishedTo != "" {
		t.Error("the site's own files were sent as well as the built pages")
	}
}

func TestBuiltPagesWithNothingBuiltIsRefused(t *testing.T) {
	repo := &fakeRepo{}
	made := publisher(t, &fakeSites{repo: repo}, publish.Target{Line: "live", BuiltPages: true})

	err := made.Publish(context.Background(), uuid.New(), nil)

	if !errors.Is(err, publish.ErrNoBuiltPages) {
		t.Fatalf("Publish() error = %v, want ErrNoBuiltPages", err)
	}
	if repo.publishedPages != nil || repo.publishedTo != "" {
		t.Error("something was published anyway")
	}
}

func TestAPublishWithNothingToSendIsNotAFailure(t *testing.T) {
	repo := &fakeRepo{err: gitsync.ErrNothingChanged}
	made := publisher(t, &fakeSites{repo: repo}, publish.Target{})

	if err := made.Publish(context.Background(), uuid.New(), nil); err != nil {
		t.Errorf("Publish() error = %v, want publishing nothing to be fine", err)
	}
}

func TestAPublishThatCannotOpenTheSiteFails(t *testing.T) {
	made := publisher(t, &fakeSites{err: errors.New("the working copy is gone")}, publish.Target{})

	if err := made.Publish(context.Background(), uuid.New(), nil); err == nil {
		t.Error("Publish() error = nil, want the failure to be reported")
	}
}

func TestAMessageIsAlwaysGivenToTheHistory(t *testing.T) {
	repo := &fakeRepo{}
	made := publisher(t, &fakeSites{repo: repo}, publish.Target{Line: "live", BuiltPages: true})

	if err := made.Publish(context.Background(), uuid.New(), fstest.MapFS{"index.html": {}}); err != nil {
		t.Fatal(err)
	}

	if repo.message == "" {
		t.Error("the published version has no message")
	}
}

func TestNewPublisherRefusesAMissingPart(t *testing.T) {
	if _, err := publish.NewPublisher(nil, func(context.Context, uuid.UUID) (publish.Target, error) {
		return publish.Target{}, nil
	}); err == nil {
		t.Error("NewPublisher() with no sites error = nil, want an error")
	}
	if _, err := publish.NewPublisher(&fakeSites{}, nil); err == nil {
		t.Error("NewPublisher() with no target error = nil, want an error")
	}
}
