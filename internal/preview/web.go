package preview

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/tenancy"
)

// Web answers the HTTP layer's questions.
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

func (w *Web) PreviewHandler(ctx context.Context, who httpapi.Actor, organisation, project string) (http.Handler, error) {
	if who.Anonymous || who.ID == "" {
		return nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return nil, httpapi.ErrNotAllowed
	}

	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return nil, err
	}

	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return nil, httpapi.ErrMissing
	}
	if err != nil {
		return nil, err
	}

	return w.service.Handler(ctx, actor, found.ID)
}

var _ httpapi.PreviewService = (*Web)(nil)
