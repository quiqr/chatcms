package publish

import (
	"context"
	"errors"
	"io/fs"
	"sort"
	"time"

	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/model"
)

// Change is one thing a person will see in the review list.
type Change struct {
	Kind    Kind          `json:"kind"`
	Section string        `json:"section"`
	Page    string        `json:"page"`
	By      string        `json:"by"`
	At      time.Time     `json:"at"`
	Fields  []FieldChange `json:"fields,omitempty"`
	Path    string        `json:"-"`
}

// Summary is what the review screen shows before anybody presses publish.
type Summary struct {
	Changes       []Change   `json:"changes"`
	LastPublished *time.Time `json:"lastPublished,omitempty"`
}

func (s Summary) Count() int { return len(s.Changes) }

// Reader is the part of a site a summary needs: what differs from the
// published site, what the files say now, and what they said when they were
// published.
type Reader interface {
	Changes(ctx context.Context) ([]gitsync.Change, error)
	ReadPublished(ctx context.Context, name string) ([]byte, error)
	FS() fs.FS
}

// Describe turns what differs into pages and fields. A file it cannot read
// as content is still listed, named by its path, because a change nobody is
// shown is worse than one shown plainly.
func Describe(ctx context.Context, site Reader, m *model.Model) (Summary, error) {
	differences, err := site.Changes(ctx)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{Changes: []Change{}}
	for _, difference := range differences {
		summary.Changes = append(summary.Changes, describeOne(ctx, site, m, difference))
	}

	sort.SliceStable(summary.Changes, func(a, b int) bool {
		if summary.Changes[a].Section != summary.Changes[b].Section {
			return summary.Changes[a].Section < summary.Changes[b].Section
		}
		return summary.Changes[a].Page < summary.Changes[b].Page
	})
	return summary, nil
}

func describeOne(ctx context.Context, site Reader, m *model.Model, difference gitsync.Change) Change {
	place := Locate(m, difference.Path)
	change := Change{
		Kind:    kindOf(difference, place),
		Section: place.Section,
		Page:    place.Page,
		By:      difference.By.Name,
		At:      difference.At,
		Path:    difference.Path,
	}

	if change.Kind != ChangedPage {
		return change
	}
	change.Fields = fieldsOf(ctx, site, place, difference.Path)
	return change
}

func kindOf(difference gitsync.Change, place Place) Kind {
	image := isImage(difference.Path)
	content := place.Known || isContent(difference.Path)

	switch {
	case difference.Kind == gitsync.Added && image:
		return NewImage
	case difference.Kind == gitsync.Removed && image:
		return RemovedImage
	case difference.Kind == gitsync.Added && content:
		return NewPage
	case difference.Kind == gitsync.Removed && content:
		return RemovedPage
	case difference.Kind == gitsync.Edited && content:
		return ChangedPage
	default:
		return OtherFile
	}
}

func fieldsOf(ctx context.Context, site Reader, place Place, name string) []FieldChange {
	was, err := site.ReadPublished(ctx, name)
	if err != nil {
		if !errors.Is(err, gitsync.ErrNotPublished) {
			return nil
		}
		was = nil
	}
	now, err := fs.ReadFile(site.FS(), name)
	if err != nil {
		return nil
	}

	before, beforeErr := frontmatter.Parse(name, was)
	after, afterErr := frontmatter.Parse(name, now)
	if afterErr != nil {
		return nil
	}
	if beforeErr != nil {
		before = nil
	}
	return Compare(place.Fields, before, after)
}
