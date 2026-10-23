package content

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/forms"
	"github.com/quiqr/chatcms/internal/frontmatter"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
)

// Site is what editing one page needs from a project's working copy.
type Site interface {
	FS() fs.FS
	Save(ctx context.Context, save gitsync.Save) (gitsync.Version, error)
	Changes(ctx context.Context) ([]gitsync.Change, error)
	Model() (*model.Model, error)
}

// Sources opens a project's site, says what a person may do with it, and who
// they are.
type Sources interface {
	Site(ctx context.Context, projectID uuid.UUID) (Site, func() error, error)
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
	Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error)
}

// Syncer is told that a site changed, so that the change reaches the remote
// without anybody waiting for it.
type Syncer interface {
	SyncSoon(ctx context.Context, projectID uuid.UUID) error
}

var (
	ErrIncomplete = errors.New("content: the service is missing a part it needs")
	// ErrNoModel says the site has no Quiqr model, so there are no forms to
	// render yet.
	ErrNoModel = errors.New("content: this site has no model yet")
)

type Service struct {
	sources Sources
	sync    Syncer
}

func NewService(sources Sources, sync Syncer) (*Service, error) {
	if sources == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources, sync: sync}, nil
}

// Editable is one page as a form: what to draw, and what is in it now.
type Editable struct {
	Key      string
	Title    string
	Section  string
	Controls []Control
	Body     string
	// HasBody says the model declares a field for the page's own text. Where
	// it does not, the text is left exactly as it is: what a site may edit is
	// the model's decision, not this package's.
	HasBody bool
}

// Read gives a page as a form.
func (s *Service) Read(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key string) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionReadContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	page, form, err := load(ctx, site, key)
	if err != nil {
		return Editable{}, err
	}

	return editable(page, form), nil
}

func editable(page Page, form *forms.Form) Editable {
	return Editable{
		Key:      page.Key,
		Title:    page.Title,
		Section:  page.Section,
		Controls: Describe(page.Fields, form),
		Body:     form.Body(),
		HasBody:  hasBody(form),
	}
}

// hasBody says the model declares a field for the page's own text. The
// reference implementation edits that text through a field of its own, so a
// model that does not declare one is a model whose pages have no editable
// text.
func hasBody(form *forms.Form) bool {
	for _, binding := range form.Bindings() {
		if binding.IsBody {
			return true
		}
	}
	return false
}

func load(ctx context.Context, site Site, key string) (Page, *forms.Form, error) {
	m, err := site.Model()
	if errors.Is(err, model.ErrNoModel) {
		return Page{}, nil, ErrNoModel
	}
	if err != nil {
		// A model with a mistake in it is not a site without one. What the
		// parser says names the file and the field, which is what somebody
		// fixing it needs.
		return Page{}, nil, &httpapi.RefusedError{Reason: "this site's model cannot be read: " + err.Error()}
	}

	page, found := Locate(m, key)
	if !found || page.File == "" {
		return Page{}, nil, httpapi.ErrMissing
	}

	content, err := fs.ReadFile(site.FS(), page.File)
	if err != nil {
		return Page{}, nil, httpapi.ErrMissing
	}

	form, err := forms.Load(page.Fields, page.File, content)
	if err != nil {
		return Page{}, nil, fmt.Errorf("content: reading %s: %w", page.Key, err)
	}
	return page, form, nil
}

// Edit is what somebody changed: values by the path of the control they
// changed, and the page's own text.
type Edit struct {
	Values map[string]any
	Body   *string
}

// Save writes an edit back and makes a version of it. What it writes is
// produced by internal/forms, which keeps the file byte for byte as it was
// apart from what changed.
func (s *Service) Save(ctx context.Context, actor policy.Actor, projectID uuid.UUID, key string, edit Edit) (Editable, error) {
	if err := s.allowed(ctx, actor, policy.ActionEditContent, projectID); err != nil {
		return Editable{}, err
	}

	site, done, err := s.sources.Site(ctx, projectID)
	if err != nil {
		return Editable{}, err
	}
	defer done()

	page, form, err := load(ctx, site, key)
	if err != nil {
		return Editable{}, err
	}

	known := map[string]bool{}
	for _, path := range Paths(Describe(page.Fields, form)) {
		known[path] = true
	}

	for path, value := range edit.Values {
		if !known[path] {
			return Editable{}, &httpapi.RefusedError{
				Reason: fmt.Sprintf("this page has no field called %q", path),
			}
		}
		if err := set(form, path, value); err != nil {
			return Editable{}, &httpapi.RefusedError{Reason: err.Error()}
		}
	}
	if edit.Body != nil {
		if !hasBody(form) {
			return Editable{}, &httpapi.RefusedError{
				Reason: "this page's model does not give its text a field, so there is nothing to write it to",
			}
		}
		if err := form.Set(forms.BodyKey, *edit.Body); err != nil {
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

	if _, err := site.Save(ctx, gitsync.Save{
		Message: "Edit " + page.Title,
		Author:  who,
		Writes:  map[string][]byte{page.File: written},
	}); err != nil && !errors.Is(err, gitsync.ErrNothingChanged) {
		return Editable{}, err
	}

	s.syncSoon(ctx, projectID)
	return s.after(ctx, site, key)
}

func (s *Service) after(ctx context.Context, site Site, key string) (Editable, error) {
	page, form, err := load(ctx, site, key)
	if err != nil {
		return Editable{}, err
	}
	return editable(page, form), nil
}

// set writes one value, clearing the field when nothing was given rather than
// writing an empty value the model never had.
func set(form *forms.Form, path string, value any) error {
	if value == nil {
		return form.Clear(path)
	}
	return form.Set(path, value)
}

func unwritable(err error) string {
	if errors.Is(err, frontmatter.ErrUnwritableFormat) {
		return "this page is in a format ChatCMS will not write, so it was left alone"
	}
	return err.Error()
}

func (s *Service) syncSoon(ctx context.Context, projectID uuid.UUID) {
	if s.sync == nil {
		return
	}
	// A failure to queue is not a failure to save. The edit is already a
	// version; the next sync will carry it.
	_ = s.sync.SyncSoon(ctx, projectID)
}

func (s *Service) who(ctx context.Context, actor policy.Actor) (gitsync.Identity, error) {
	personID, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return gitsync.Identity{}, httpapi.ErrNotAllowed
	}
	return s.sources.Who(ctx, personID)
}

func (s *Service) allowed(ctx context.Context, actor policy.Actor, action policy.Action, projectID uuid.UUID) error {
	resource, err := s.sources.Resource(ctx, projectID)
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
