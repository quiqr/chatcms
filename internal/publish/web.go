package publish

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/policy"
	"github.com/quiqr/chatcms/internal/tenancy"
)

// Web answers the HTTP layer's questions. It turns slugs into projects and a
// signed-in person into an actor, and leaves every decision about who may do
// what to the service and through it to internal/policy.
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

func (w *Web) Review(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ReviewView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ReviewView{}, err
	}

	summary, err := w.service.Review(ctx, actor, id)
	if err != nil {
		return httpapi.ReviewView{}, err
	}
	return reviewView(summary), nil
}

func (w *Web) Publish(ctx context.Context, who httpapi.Actor, organisation, project, note string) (httpapi.PublishView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PublishView{}, err
	}

	record, err := w.service.Run(ctx, actor, id, note)
	if errors.Is(err, ErrAlreadyPublishing) {
		return httpapi.PublishView{}, &httpapi.RefusedError{
			Key:    "refused.site-already-being-published-wait",
			Reason: "this site is already being published; wait for that to finish",
		}
	}
	if err != nil {
		return httpapi.PublishView{}, err
	}
	return publishView(record), nil
}

func (w *Web) ReadPublish(ctx context.Context, who httpapi.Actor, organisation, project, id string) (httpapi.PublishView, error) {
	actor, projectID, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.PublishView{}, err
	}

	publishID, err := uuid.Parse(id)
	if err != nil {
		return httpapi.PublishView{}, httpapi.ErrMissing
	}

	record, err := w.service.Get(ctx, actor, projectID, publishID)
	if errors.Is(err, ErrNoSuchPublish) {
		return httpapi.PublishView{}, httpapi.ErrMissing
	}
	if err != nil {
		return httpapi.PublishView{}, err
	}
	return publishView(record), nil
}

func (w *Web) PublishHistory(ctx context.Context, who httpapi.Actor, organisation, project string) ([]httpapi.PublishView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return nil, err
	}

	records, err := w.service.History(ctx, actor, id, 0)
	if err != nil {
		return nil, err
	}

	views := make([]httpapi.PublishView, 0, len(records))
	for _, record := range records {
		views = append(views, publishView(record))
	}
	return views, nil
}

// ReviewViewOf is how a summary of changes reaches the interface. It is
// exported because the history shows what one version changed in exactly the
// same shape as the review shows what is waiting.
func ReviewViewOf(summary Summary) httpapi.ReviewView {
	return reviewView(summary)
}

// PublishViewOf is how one publish reaches the interface.
func PublishViewOf(record Record) httpapi.PublishView {
	return publishView(record)
}

func reviewView(summary Summary) httpapi.ReviewView {
	view := httpapi.ReviewView{Count: summary.Count(), Changes: []httpapi.ChangeView{}}
	if summary.LastPublished != nil {
		view.LastPublished = summary.LastPublished.Format(time.RFC3339)
	}

	for _, change := range summary.Changes {
		entry := httpapi.ChangeView{
			Kind:    string(change.Kind),
			Section: change.Section,
			Page:    change.Page,
			By:      change.By,
		}
		if !change.At.IsZero() {
			entry.At = change.At.Format(time.RFC3339)
		}
		for _, field := range change.Fields {
			entry.Fields = append(entry.Fields, fieldView(field))
		}
		view.Changes = append(view.Changes, entry)
	}
	return view
}

func fieldView(field FieldChange) httpapi.FieldChangeView {
	view := httpapi.FieldChangeView{Name: field.Name}
	for _, word := range field.Words {
		view.Words = append(view.Words, httpapi.WordView{Part: string(word.Part), Text: word.Text})
	}
	return view
}

func publishView(record Record) httpapi.PublishView {
	view := httpapi.PublishView{
		ID:          record.ID.String(),
		State:       string(record.State),
		Note:        record.Note,
		ChangeCount: record.ChangeCount,
		StartedAt:   record.StartedAt.Format(time.RFC3339),
		Trouble:     record.Trouble,
		Detail:      record.Detail,
	}
	if record.FinishedAt != nil {
		view.FinishedAt = record.FinishedAt.Format(time.RFC3339)
	}
	for _, step := range record.Steps {
		view.Steps = append(view.Steps, httpapi.StepView{Title: step.Title, State: string(step.State)})
	}
	return view
}

var _ httpapi.PublishService = (*Web)(nil)
