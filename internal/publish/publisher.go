package publish

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
)

// Target is where and in what form a project publishes. It comes from the
// project's own settings; this package does not read them itself.
type Target struct {
	// Line is the name of the line to publish to. Empty means the line the
	// site is edited on.
	Line string
	// BuiltPages says that publishing sends the pages ChatCMS built rather
	// than the site's own files.
	BuiltPages bool
	// Author is who the publish is recorded as. It is the person who pressed
	// the button, so that the history of a built line reads like the history
	// of the site.
	Author gitsync.Identity
	// Message is what the publish is called in the history.
	Message string
}

var ErrNoBuiltPages = errors.New("publish: this project publishes built pages and none were built")

// Targets answers where a project publishes. The caller reads the project's
// settings; this package stays out of the database.
type Targets func(ctx context.Context, projectID uuid.UUID) (Target, error)

// Sites opens a project's working copy.
type Sites interface {
	Open(ctx context.Context, id string) (gitsync.Repo, error)
}

// Publisher carries out the publishing step of a workflow.
type Publisher struct {
	sites   Sites
	targets Targets
}

func NewPublisher(sites Sites, targets Targets) (*Publisher, error) {
	if sites == nil || targets == nil {
		return nil, errors.New("publish: the publisher is missing a part it needs")
	}
	return &Publisher{sites: sites, targets: targets}, nil
}

func (p *Publisher) Publish(ctx context.Context, projectID uuid.UUID, built fs.FS) error {
	target, err := p.targets(ctx, projectID)
	if err != nil {
		return err
	}
	if target.BuiltPages && built == nil {
		return ErrNoBuiltPages
	}

	repository, err := p.sites.Open(ctx, projectID.String())
	if err != nil {
		return err
	}
	defer repository.Close()

	if target.BuiltPages {
		_, err = repository.PublishPages(ctx, target.Line, built, target.Author, messageOr(target.Message))
		return err
	}
	if _, err := repository.PublishTo(ctx, target.Line); err != nil {
		if errors.Is(err, gitsync.ErrNothingChanged) {
			return nil
		}
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func messageOr(message string) string {
	if message == "" {
		return "Published from Quiqr ChatCMS"
	}
	return message
}
