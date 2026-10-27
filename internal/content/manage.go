package content

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
)

// Slug turns what somebody typed into a file name, the way the old Quiqr
// does: small letters, digits and hyphens, with accents folded away.
func Slug(said string) string {
	lowered := strings.ToLower(strings.TrimSpace(said))

	out := strings.Builder{}
	previousHyphen := false
	for _, letter := range lowered {
		switch {
		case unicode.IsLetter(letter) && letter < unicode.MaxASCII,
			unicode.IsDigit(letter) && letter < unicode.MaxASCII:
			out.WriteRune(letter)
			previousHyphen = false
		case folded[letter] != "":
			out.WriteString(folded[letter])
			previousHyphen = false
		case !previousHyphen && out.Len() > 0:
			out.WriteRune('-')
			previousHyphen = true
		}
	}
	return strings.Trim(out.String(), "-")
}

// folded is the handful of letters a Dutch or German site actually uses. A
// letter nothing here knows becomes a hyphen, which is what the old Quiqr
// does with it too.
var folded = map[rune]string{
	'á': "a", 'à': "a", 'ä': "a", 'â': "a", 'å': "a",
	'é': "e", 'è': "e", 'ë': "e", 'ê': "e",
	'í': "i", 'ì': "i", 'ï': "i", 'î': "i",
	'ó': "o", 'ò': "o", 'ö': "o", 'ô': "o",
	'ú': "u", 'ù': "u", 'ü': "u", 'û': "u",
	'ñ': "n", 'ç': "c", 'ß': "ss",
}

var usableSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewPage is what somebody asked to make.
type NewPage struct {
	// Section is the collection's key. Only a collection can grow.
	Section string
	// Title is what they typed. The file name comes from it.
	Title string
}

// Create makes a page in a collection, with the title filled in and nothing
// else, so that opening it is the next thing that happens.
func (s *Service) Create(ctx context.Context, actor policy.Actor, projectID uuid.UUID, asked NewPage) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	collection, err := collectionOf(site, asked.Section)
	if err != nil {
		return Editable{}, err
	}

	slug := Slug(asked.Title)
	if !usableSlug.MatchString(slug) {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-needs-name-with-some",
			Reason: "a page needs a name with some letters or numbers in it",
		}
	}

	file := path.Join(strings.TrimSuffix(collection.Folder, "/"), slug+"."+extensionOf(collection))
	if _, err := fs.Stat(site.FS(), file); err == nil {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-name-taken",
			Reason: fmt.Sprintf("there is already a page called %q here", asked.Title),
			Detail: map[string]any{"title": asked.Title},
		}
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return Editable{}, err
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Add " + asked.Title,
		Author:  who,
		Writes:  map[string][]byte{file: firstContent(asked.Title)},
	}); err != nil {
		return Editable{}, err
	}

	s.syncSoon(ctx, projectID)
	return s.after(ctx, site, collection.Key+"/"+slug)
}

// firstContent is what a new page holds: its title and nothing else, in the
// shape the reference implementation writes.
func firstContent(title string) []byte {
	return []byte("---\ntitle: " + title + "\n---\n\n")
}

func extensionOf(collection model.Collection) string {
	if collection.Extension != "" {
		return collection.Extension
	}
	return "md"
}

func collectionOf(site Site, key string) (model.Collection, error) {
	m, err := site.Model()
	if err != nil {
		return model.Collection{}, ErrNoModel
	}
	for _, collection := range m.Collections {
		if collection.Key == key && !collection.Hidden {
			return collection, nil
		}
	}
	return model.Collection{}, httpapi.ErrMissing
}

// Rename gives a page a new name, which means a new file. The old Quiqr names
// a file after its title, so renaming is what keeps the two in step.
func (s *Service) Rename(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key, title string) (Editable, error) {
	return s.moveOrCopy(ctx, actor, projectID, key, title, false)
}

// Duplicate makes a copy of a page under a new name.
func (s *Service) Duplicate(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key, title string) (Editable, error) {
	return s.moveOrCopy(ctx, actor, projectID, key, title, true)
}

func (s *Service) moveOrCopy(
	ctx context.Context,
	actor policy.Actor,
	projectID uuid.UUID,
	key, title string,
	keepTheOldOne bool,
) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	section, _, nested := strings.Cut(key, "/")
	if !nested {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-model-names-cannot-renamed",
			Reason: "a page the model names cannot be renamed here; its name is in the model",
		}
	}

	collection, err := collectionOf(site, section)
	if err != nil {
		return Editable{}, err
	}

	page, form, err := load(ctx, site, key)
	if err != nil {
		return Editable{}, err
	}

	slug := Slug(title)
	if !usableSlug.MatchString(slug) {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-needs-name-with-some",
			Reason: "a page needs a name with some letters or numbers in it",
		}
	}

	to := path.Join(strings.TrimSuffix(collection.Folder, "/"), slug+"."+extensionOf(collection))
	if to == page.File {
		return s.after(ctx, site, key)
	}
	if _, err := fs.Stat(site.FS(), to); err == nil {
		return Editable{}, &httpapi.RefusedError{
			Key:    "refused.page-name-taken",
			Reason: fmt.Sprintf("there is already a page called %q here", title),
			Detail: map[string]any{"title": title},
		}
	}

	// The title inside the page follows its name, which is what the old
	// Quiqr does and what stops a renamed page still calling itself the
	// old thing.
	if _, has := form.Get("title"); has {
		if err := form.Set("title", title); err != nil {
			return Editable{}, &httpapi.RefusedError{Reason: err.Error()}
		}
	}
	written, err := form.Bytes()
	if err != nil {
		return Editable{}, &httpapi.RefusedError{Reason: unwritable(err)}
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return Editable{}, err
	}

	save := gitsync.Save{
		Author: who,
		Writes: map[string][]byte{to: written},
	}
	if keepTheOldOne {
		save.Message = "Copy " + page.Title + " to " + title
	} else {
		save.Message = "Rename " + page.Title + " to " + title
		save.Deletes = []string{page.File}
	}

	if _, err := site.Save(ctx, save); err != nil {
		return Editable{}, err
	}

	s.syncSoon(ctx, projectID)
	return s.after(ctx, site, collection.Key+"/"+slug)
}

// Delete removes a page. Only an item of a collection can go: a page the model
// names is part of the site's shape.
func (s *Service) Delete(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key string) error {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return err
	}
	defer done()

	section, _, nested := strings.Cut(key, "/")
	if !nested {
		return &httpapi.RefusedError{
			Key:    "refused.page-model-names-cannot-deleted",
			Reason: "a page the model names cannot be deleted here; it is part of how the site is built",
		}
	}
	if _, err := collectionOf(site, section); err != nil {
		return err
	}

	page, _, err := load(ctx, site, key)
	if err != nil {
		return err
	}

	who, err := s.who(ctx, actor)
	if err != nil {
		return err
	}

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Remove " + page.Title,
		Author:  who,
		Deletes: []string{page.File},
	}); err != nil {
		return err
	}

	s.syncSoon(ctx, projectID)
	return nil
}

// List is the content browser: every section with its items, and a mark on
// anything not yet published.
func (s *Service) List(ctx context.Context, actor policy.Actor, projectID uuid.UUID) ([]Section, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return nil, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer done()

	if _, err := site.Model(); err != nil {
		return nil, noModel(err)
	}

	changes, err := site.Changes(ctx)
	if err != nil {
		return nil, err
	}
	return Browse(ctx, site, waitingFrom(changes)), nil
}

func noModel(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), model.ErrNoModel.Error()) {
		return ErrNoModel
	}
	return &httpapi.RefusedError{
		Key:    "refused.model-cannot-be-read",
		Reason: "this site's model cannot be read: " + err.Error(),
		Detail: map[string]any{"why": err.Error()},
	}
}
