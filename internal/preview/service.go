package preview

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
)

// Sources says where a project's working copy is and what a person may do
// with it.
type Sources interface {
	Dir(ctx context.Context, projectID uuid.UUID) (string, error)
	Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error)
}

var ErrIncomplete = errors.New("preview: the service is missing a part it needs")

// Service answers requests for a site's preview, through the app, so that a
// preview is only ever seen by somebody allowed to see the project.
type Service struct {
	sources Sources
	servers *Servers
}

func NewService(sources Sources, servers *Servers) (*Service, error) {
	if sources == nil || servers == nil {
		return nil, ErrIncomplete
	}
	return &Service{sources: sources, servers: servers}, nil
}

// Handler gives what answers requests for one site's preview. Anybody who may
// read the project may preview it: looking is not changing.
func (s *Service) Handler(ctx context.Context, actor policy.Actor, projectID uuid.UUID) (http.Handler, error) {
	resource, err := s.sources.Resource(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return nil, httpapi.ErrMissing
	}

	dir, err := s.sources.Dir(ctx, projectID)
	if err != nil {
		return nil, httpapi.ErrMissing
	}

	handler, err := s.servers.For(ctx, projectID.String(), dir)
	if errors.Is(err, ErrTooMany) {
		return nil, &httpapi.RefusedError{
			Reason: "too many sites are being previewed at once; try again in a few minutes",
		}
	}
	if errors.Is(err, ErrNotReady) {
		return nil, &httpapi.RefusedError{
			Reason: "the preview did not start. Your site may have something in it Hugo cannot build.",
		}
	}
	if err != nil {
		return nil, err
	}
	return handler, nil
}

// Stop ends a site's preview, for when the working copy changes underneath it.
func (s *Service) Stop(projectID uuid.UUID) {
	s.servers.Stop(projectID.String())
}
