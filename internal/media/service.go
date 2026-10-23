package media

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
)

// Site is what the media library needs from a project's working copy.
type Site interface {
	FS() fs.FS
	Save(ctx context.Context, save gitsync.Save) (gitsync.Version, error)
	Changes(ctx context.Context) ([]gitsync.Change, error)
	Model() (*model.Model, error)
}

// Sources opens a project's site, says what a person may do, and who they are.
type Sources interface {
	Site(ctx context.Context, projectID uuid.UUID) (Site, func() error, error)
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
	Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error)
}

// Syncer is told the site changed, so a picture reaches the remote without
// anybody waiting.
type Syncer interface {
	SyncSoon(ctx context.Context, projectID uuid.UUID) error
}

var ErrIncomplete = errors.New("media: the library is missing a part it needs")

type Library struct {
	sources    Sources
	sync       Syncer
	thumbnails *Thumbnails
}

func NewLibrary(sources Sources, sync Syncer, thumbnails *Thumbnails) (*Library, error) {
	if sources == nil {
		return nil, ErrIncomplete
	}
	return &Library{sources: sources, sync: sync, thumbnails: thumbnails}, nil
}

// List is every picture in the site, with what is known about each.
func (l *Library) List(ctx context.Context, actor policy.Actor, projectID uuid.UUID) ([]Image, error) {
	if err := l.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return nil, err
	}

	site, done, err := l.sources.Site(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer done()

	return l.pictures(ctx, site), nil
}

func (l *Library) pictures(ctx context.Context, site Site) []Image {
	fsys := site.FS()
	folder := FolderIn(fsys)
	said := ReadDescriptions(fsys)
	m, _ := site.Model()

	changed := map[string]gitsync.Change{}
	if changes, err := site.Changes(ctx); err == nil {
		for _, change := range changes {
			changed[change.Path] = change
		}
	}

	found := []Image{}
	for _, name := range Pictures(fsys, folder) {
		where := path.Join(folder, name)
		picture := Image{
			Path:   where,
			Name:   name,
			Alt:    said[name],
			UsedOn: UsedOn(fsys, m, folder, name),
		}
		if info, err := fs.Stat(fsys, where); err == nil {
			picture.Size = info.Size()
		}
		if change, waiting := changed[where]; waiting {
			picture.By = change.By.Name
			picture.At = change.At
		}
		found = append(found, picture)
	}
	return found
}

// Add puts a picture in the site. What it is is decided by its own bytes, not
// by what it was called, because a file named .png that is not one would show
// as broken on the live site.
func (l *Library) Add(ctx context.Context, actor policy.Actor, projectID uuid.UUID, called string, content []byte) (Image, error) {
	if err := l.allowed(ctx, actor, policy.ActionEditMedia, projectID); err != nil {
		return Image{}, err
	}
	if len(content) == 0 {
		return Image{}, &httpapi.RefusedError{Reason: "that file is empty"}
	}
	if len(content) > MaxBytes {
		return Image{}, &httpapi.RefusedError{
			Reason: fmt.Sprintf("that picture is %s, and the most a site will take is %s",
				inWords(int64(len(content))), inWords(MaxBytes)),
		}
	}

	_, extension, err := KindOf(content)
	if err != nil {
		return Image{}, &httpapi.RefusedError{
			Reason: "that is not a picture. ChatCMS takes PNG, JPEG, GIF, WebP and SVG.",
		}
	}

	name, err := Name(called, extension)
	if err != nil {
		return Image{}, &httpapi.RefusedError{
			Reason: "that file needs a name with some letters or numbers in it",
		}
	}

	site, done, err := l.sources.Site(ctx, projectID)
	if err != nil {
		return Image{}, err
	}
	defer done()

	folder := FolderIn(site.FS())
	name = freeName(site.FS(), folder, name)
	where := path.Join(folder, name)

	who, err := l.who(ctx, actor)
	if err != nil {
		return Image{}, err
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Add the picture " + name,
		Author:  who,
		Writes:  map[string][]byte{where: content},
	}); err != nil {
		return Image{}, err
	}

	l.syncSoon(ctx, projectID)
	return Image{Path: where, Name: name, Size: int64(len(content)), By: who.Name}, nil
}

// freeName leaves a picture that is already there alone, putting the new one
// beside it. Replacing silently would change every page that uses the old one.
func freeName(fsys fs.FS, folder, name string) string {
	if _, err := fs.Stat(fsys, path.Join(folder, name)); err != nil {
		return name
	}

	extension := path.Ext(name)
	stem := strings.TrimSuffix(name, extension)
	for number := 2; number < 1000; number++ {
		tried := fmt.Sprintf("%s-%d%s", stem, number, extension)
		if _, err := fs.Stat(fsys, path.Join(folder, tried)); err != nil {
			return tried
		}
	}
	return name
}

func inWords(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.0f kB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// Describe writes what a picture shows, for somebody using a screen reader.
func (l *Library) Describe(ctx context.Context, actor policy.Actor, projectID uuid.UUID, name, said string) (Image, error) {
	if err := l.allowed(ctx, actor, policy.ActionEditMedia, projectID); err != nil {
		return Image{}, err
	}
	if err := checkName(name); err != nil {
		return Image{}, err
	}

	site, done, err := l.sources.Site(ctx, projectID)
	if err != nil {
		return Image{}, err
	}
	defer done()

	folder := FolderIn(site.FS())
	if _, err := fs.Stat(site.FS(), path.Join(folder, name)); err != nil {
		return Image{}, httpapi.ErrMissing
	}

	held := ReadDescriptions(site.FS())
	held[name] = strings.TrimSpace(said)
	written, err := WriteDescriptions(held)
	if err != nil {
		return Image{}, err
	}

	who, err := l.who(ctx, actor)
	if err != nil {
		return Image{}, err
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Describe the picture " + name,
		Author:  who,
		Writes:  map[string][]byte{DescriptionsFile: written},
	}); err != nil && !errors.Is(err, gitsync.ErrNothingChanged) {
		return Image{}, err
	}

	l.syncSoon(ctx, projectID)
	for _, picture := range l.pictures(ctx, site) {
		if picture.Name == name {
			return picture, nil
		}
	}
	return Image{}, httpapi.ErrMissing
}

// Remove takes a picture out of the site. A picture something still uses is
// refused unless whoever asked says they know, because the pages that use it
// would show a gap.
func (l *Library) Remove(ctx context.Context, actor policy.Actor, projectID uuid.UUID, name string, evenIfUsed bool) error {
	if err := l.allowed(ctx, actor, policy.ActionEditMedia, projectID); err != nil {
		return err
	}
	if err := checkName(name); err != nil {
		return err
	}

	site, done, err := l.sources.Site(ctx, projectID)
	if err != nil {
		return err
	}
	defer done()

	folder := FolderIn(site.FS())
	where := path.Join(folder, name)
	if _, err := fs.Stat(site.FS(), where); err != nil {
		return httpapi.ErrMissing
	}

	m, _ := site.Model()
	if used := UsedOn(site.FS(), m, folder, name); len(used) > 0 && !evenIfUsed {
		return &httpapi.RefusedError{
			Reason: fmt.Sprintf("that picture is on %s", pagesInWords(len(used))),
		}
	}

	who, err := l.who(ctx, actor)
	if err != nil {
		return err
	}

	writes := map[string][]byte{}
	held := ReadDescriptions(site.FS())
	if _, described := held[name]; described {
		delete(held, name)
		written, err := WriteDescriptions(held)
		if err != nil {
			return err
		}
		writes[DescriptionsFile] = written
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Remove the picture " + name,
		Author:  who,
		Writes:  writes,
		Deletes: []string{where},
	}); err != nil {
		return err
	}

	l.syncSoon(ctx, projectID)
	return nil
}

func pagesInWords(count int) string {
	if count == 1 {
		return "a page that is still using it"
	}
	return fmt.Sprintf("%d pages that are still using it", count)
}

// Read gives a picture back, either whole or as a small copy for a grid.
func (l *Library) Read(ctx context.Context, actor policy.Actor, projectID uuid.UUID, name string, small bool) ([]byte, string, error) {
	if err := l.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return nil, "", err
	}
	if err := checkName(name); err != nil {
		return nil, "", err
	}

	site, done, err := l.sources.Site(ctx, projectID)
	if err != nil {
		return nil, "", err
	}
	defer done()

	folder := FolderIn(site.FS())
	content, err := fs.ReadFile(site.FS(), path.Join(folder, name))
	if err != nil {
		return nil, "", httpapi.ErrMissing
	}

	kind := KindByName(name)
	if !small || l.thumbnails == nil {
		return content, kind, nil
	}
	return l.thumbnails.Of(content, kind)
}

// checkName refuses anything that is not plainly one file in the pictures
// folder, so that a name from outside cannot reach another part of the site.
func checkName(name string) error {
	if name == "" || name == "." || name == ".." ||
		strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return httpapi.ErrMissing
	}
	return nil
}

func (l *Library) syncSoon(ctx context.Context, projectID uuid.UUID) {
	if l.sync == nil {
		return
	}
	_ = l.sync.SyncSoon(ctx, projectID)
}

func (l *Library) who(ctx context.Context, actor policy.Actor) (gitsync.Identity, error) {
	personID, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return gitsync.Identity{}, httpapi.ErrNotAllowed
	}
	return l.sources.Who(ctx, personID)
}

func (l *Library) allowed(ctx context.Context, actor policy.Actor, action policy.Action, projectID uuid.UUID) error {
	resource, err := l.sources.Resource(ctx, projectID)
	if err != nil {
		return err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return httpapi.ErrMissing
	}
	if !policy.Can(actor, action, resource) {
		return httpapi.ErrNotAllowed
	}
	return nil
}
