package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/tenancy"
)

// Web answers the HTTP layer's questions, turning slugs into a project and a
// signed-in person into an actor.
type Web struct {
	service *Service
	tenancy *tenancy.Store
}

func NewWeb(service *Service, tenancyStore *tenancy.Store) (*Web, error) {
	if service == nil || tenancyStore == nil {
		return nil, ErrIncomplete
	}
	return &Web{service: service, tenancy: tenancyStore}, nil
}

func (w *Web) reach(ctx context.Context, who httpapi.Actor, organisation, project string) (policy.Actor, uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}

	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}

	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return policy.Actor{}, uuid.Nil, httpapi.ErrMissing
	}
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}
	return actor, found.ID, nil
}

func (w *Web) ReadPage(ctx context.Context, who httpapi.Actor, organisation, project, page string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	editable, err := w.service.Read(ctx, actor, id, page)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(editable), nil
}

func (w *Web) SavePage(ctx context.Context, who httpapi.Actor, organisation, project, page string, edit httpapi.EditRequest) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	saved, err := w.service.Save(ctx, actor, id, page, Edit{Values: edit.Values, Body: edit.Body})
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(saved), nil
}

// noModelIsRefusal turns a site with no model into something worth showing
// rather than a failure nobody can act on.
func noModelIsRefusal(err error) error {
	if errors.Is(err, ErrNoModel) {
		return &httpapi.RefusedError{
			Reason: "this site has no Quiqr model yet, so there are no forms to fill in",
		}
	}
	return err
}

func pageView(editable Editable) httpapi.PageView {
	return httpapi.PageView{
		Key:     editable.Key,
		Title:   editable.Title,
		Section: editable.Section,
		Fields:  controlViews(editable.Controls),
		Body:    editable.Body,
		HasBody: editable.HasBody,
	}
}

func controlViews(controls []Control) []httpapi.ControlView {
	views := make([]httpapi.ControlView, 0, len(controls))
	for _, control := range controls {
		view := httpapi.ControlView{
			Path:    control.Path,
			Type:    control.Type,
			Title:   control.Title,
			Tip:     control.Tip,
			Value:   control.Value,
			Config:  control.Config,
			Unknown: control.Unknown,
		}
		for _, option := range control.Options {
			view.Options = append(view.Options, httpapi.OptionView{Value: option.Value, Label: option.Label})
		}
		if len(control.Controls) > 0 {
			view.Fields = controlViews(control.Controls)
		}
		views = append(views, view)
	}
	return views
}

var _ httpapi.ContentService = (*Web)(nil)

func (w *Web) ListContent(ctx context.Context, who httpapi.Actor, organisation, project string) ([]httpapi.SectionView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return nil, err
	}

	sections, err := w.service.List(ctx, actor, id)
	if err != nil {
		return nil, noModelIsRefusal(err)
	}

	views := make([]httpapi.SectionView, 0, len(sections))
	for _, section := range sections {
		view := httpapi.SectionView{
			Key:      section.Key,
			Title:    section.Title,
			ItemName: section.ItemName,
			Count:    section.Count,
			Items:    make([]httpapi.ItemView, 0, len(section.Items)),
		}
		for _, item := range section.Items {
			shown := httpapi.ItemView{
				Key:     item.Key,
				Title:   item.Title,
				Waiting: item.Waiting,
				By:      item.By,
			}
			if !item.At.IsZero() {
				shown.At = item.At.Format(time.RFC3339)
			}
			view.Items = append(view.Items, shown)
		}
		views = append(views, view)
	}
	return views, nil
}

func (w *Web) CreatePage(ctx context.Context, who httpapi.Actor, organisation, project, section, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	made, err := w.service.Create(ctx, actor, id, NewPage{Section: section, Title: title})
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(made), nil
}

func (w *Web) RenamePage(ctx context.Context, who httpapi.Actor, organisation, project, page, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	renamed, err := w.service.Rename(ctx, actor, id, page, title)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(renamed), nil
}

func (w *Web) DuplicatePage(ctx context.Context, who httpapi.Actor, organisation, project, page, title string) (httpapi.PageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PageView{}, err
	}

	copied, err := w.service.Duplicate(ctx, actor, id, page, title)
	if err != nil {
		return httpapi.PageView{}, noModelIsRefusal(err)
	}
	return pageView(copied), nil
}

func (w *Web) DeletePage(ctx context.Context, who httpapi.Actor, organisation, project, page string) error {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return noModelIsRefusal(w.service.Delete(ctx, actor, id, page))
}

var _ httpapi.BrowseService = (*Web)(nil)
