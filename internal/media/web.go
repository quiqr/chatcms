package media

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
	library *Library
	tenancy *tenancy.Store
}

func NewWeb(library *Library, tenancyStore *tenancy.Store) (*Web, error) {
	if library == nil || tenancyStore == nil {
		return nil, ErrIncomplete
	}
	return &Web{library: library, tenancy: tenancyStore}, nil
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

func (w *Web) ListMedia(ctx context.Context, who httpapi.Actor, organisation, project string) ([]httpapi.ImageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return nil, err
	}

	pictures, err := w.library.List(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.ImageView, 0, len(pictures))
	for _, picture := range pictures {
		views = append(views, imageView(picture))
	}
	return views, nil
}

func (w *Web) AddMedia(ctx context.Context, who httpapi.Actor, organisation, project, name string, content []byte) (httpapi.ImageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ImageView{}, err
	}

	added, err := w.library.Add(ctx, actor, id, name, content)
	if err != nil {
		return httpapi.ImageView{}, err
	}
	return imageView(added), nil
}

func (w *Web) DescribeMedia(ctx context.Context, who httpapi.Actor, organisation, project, name, said string) (httpapi.ImageView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ImageView{}, err
	}

	described, err := w.library.Describe(ctx, actor, id, name, said)
	if err != nil {
		return httpapi.ImageView{}, err
	}
	return imageView(described), nil
}

func (w *Web) RemoveMedia(ctx context.Context, who httpapi.Actor, organisation, project, name string, evenIfUsed bool) error {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return w.library.Remove(ctx, actor, id, name, evenIfUsed)
}

func (w *Web) ReadMedia(ctx context.Context, who httpapi.Actor, organisation, project, name string, small bool) ([]byte, string, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return nil, "", err
	}
	return w.library.Read(ctx, actor, id, name, small)
}

func imageView(picture Image) httpapi.ImageView {
	view := httpapi.ImageView{
		Name:   picture.Name,
		Path:   picture.Path,
		Size:   picture.Size,
		By:     picture.By,
		Alt:    picture.Alt,
		UsedOn: picture.UsedOn,
	}
	if !picture.At.IsZero() {
		view.At = picture.At.Format(time.RFC3339)
	}
	return view
}

var _ httpapi.MediaService = (*Web)(nil)
