package history

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/publish"
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

func (w *Web) Timeline(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.TimelineView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.TimelineView{}, err
	}

	timeline, err := w.service.Timeline(ctx, actor, id)
	if err != nil {
		return httpapi.TimelineView{}, err
	}
	return timelineView(timeline), nil
}

func (w *Web) VersionContents(ctx context.Context, who httpapi.Actor, organisation, project, version string) (httpapi.ReviewView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ReviewView{}, err
	}

	summary, err := w.service.Contents(ctx, actor, id, version)
	if err != nil {
		return httpapi.ReviewView{}, err
	}
	return publish.ReviewViewOf(summary), nil
}

func (w *Web) BeforeGoingBack(ctx context.Context, who httpapi.Actor, organisation, project, version string) (httpapi.WarningView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.WarningView{}, err
	}

	warning, err := w.service.Before(ctx, actor, id, version)
	if err != nil {
		return httpapi.WarningView{}, err
	}

	view := httpapi.WarningView{
		Title:       warning.Title,
		SetAside:    warning.SetAside,
		Reassurance: reassurance(warning.SetAside),
	}
	if !warning.When.IsZero() {
		view.When = warning.When.Format(time.RFC3339)
	}
	return view, nil
}

// reassurance is the sentence the confirmation dialog shows. Nothing is lost
// by going back, and the one thing that worries people is unpublished work,
// so it is named rather than left to be wondered about.
func reassurance(setAside int) string {
	nothingIsLost := "Nothing is lost. Everything stays in your history, and you can undo this afterwards."
	switch setAside {
	case 0:
		return nothingIsLost
	case 1:
		return "One change you have not published yet will be set aside. " + nothingIsLost
	default:
		return fmt.Sprintf("%d changes you have not published yet will be set aside. %s", setAside, nothingIsLost)
	}
}

func (w *Web) GoBackTo(ctx context.Context, who httpapi.Actor, organisation, project, version string, publishRightAway bool) (httpapi.WentBackView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.WentBackView{}, err
	}

	result, err := w.service.GoBackTo(ctx, actor, id, version, publishRightAway)
	if errors.Is(err, ErrNothingToUndo) {
		return httpapi.WentBackView{}, &httpapi.RefusedError{
			Key:    "refused.site-already-exactly-as-version",
			Reason: "your site is already exactly as that version left it",
		}
	}
	if err != nil {
		return httpapi.WentBackView{}, err
	}

	view := httpapi.WentBackView{Title: result.Title}
	if result.Published != nil {
		published := publish.PublishViewOf(*result.Published)
		view.Published = &published
	}
	return view, nil
}

func timelineView(timeline Timeline) httpapi.TimelineView {
	view := httpapi.TimelineView{Unpublished: timeline.Unpublished, Groups: []httpapi.GroupView{}}

	for _, group := range timeline.Groups {
		entries := make([]httpapi.VersionView, 0, len(group.Entries))
		for _, entry := range group.Entries {
			shown := httpapi.VersionView{
				Version: entry.Version,
				Title:   entry.Title,
				By:      entry.By,
				Changes: entry.Changes,
				Live:    entry.Live,
			}
			if !entry.At.IsZero() {
				shown.At = entry.At.Format(time.RFC3339)
			}
			entries = append(entries, shown)
		}
		view.Groups = append(view.Groups, httpapi.GroupView{Period: group.Period, Entries: entries})
	}
	return view
}

var _ httpapi.HistoryService = (*Web)(nil)
