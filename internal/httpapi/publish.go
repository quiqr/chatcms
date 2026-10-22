package httpapi

import (
	"context"
	"net/http"
)

// PublishService is what the HTTP layer needs in order to publish a project:
// what changed, the act itself, and what happened the last few times.
type PublishService interface {
	Review(ctx context.Context, actor Actor, organisation, project string) (ReviewView, error)
	Publish(ctx context.Context, actor Actor, organisation, project, note string) (PublishView, error)
	ReadPublish(ctx context.Context, actor Actor, organisation, project, id string) (PublishView, error)
	PublishHistory(ctx context.Context, actor Actor, organisation, project string) ([]PublishView, error)
}

// ReviewView is what somebody sees before deciding to publish.
type ReviewView struct {
	Count         int          `json:"count"`
	LastPublished string       `json:"lastPublished,omitempty"`
	Changes       []ChangeView `json:"changes"`
}

// ChangeView is one line of the review list.
type ChangeView struct {
	Kind    string            `json:"kind"`
	Section string            `json:"section"`
	Page    string            `json:"page"`
	By      string            `json:"by,omitempty"`
	At      string            `json:"at,omitempty"`
	Fields  []FieldChangeView `json:"fields,omitempty"`
}

// FieldChangeView is one field of one page, with the words that changed.
type FieldChangeView struct {
	Name  string     `json:"name"`
	Words []WordView `json:"words,omitempty"`
}

// WordView is a run of text and what happened to it.
type WordView struct {
	Part string `json:"part"`
	Text string `json:"text"`
}

// PublishView is one publish, as a person sees it.
type PublishView struct {
	ID          string     `json:"id"`
	State       string     `json:"state"`
	Note        string     `json:"note,omitempty"`
	ChangeCount int        `json:"changeCount"`
	StartedAt   string     `json:"startedAt"`
	FinishedAt  string     `json:"finishedAt,omitempty"`
	Steps       []StepView `json:"steps,omitempty"`
	// Trouble is written for whoever pressed the button. It carries no
	// technical vocabulary and is safe to show to anybody.
	Trouble string `json:"trouble,omitempty"`
	// Detail is the raw output. It is only ever filled in for somebody the
	// policy function allows technical detail.
	Detail string `json:"detail,omitempty"`
}

// StepView is one step of the publish, for the progress list.
type StepView struct {
	Title string `json:"title"`
	State string `json:"state"`
}

type publishRequest struct {
	Note string `json:"note"`
}

func reviewHandler(service PublishService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		review, err := service.Review(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, review)
	})
}

func publishHandler(service PublishService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request publishRequest
		if err := Decode(w, r, &request); err != nil {
			WriteDecodeError(w, err)
			return
		}

		record, err := service.Publish(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), request.Note)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, record)
	})
}

func readPublishHandler(service PublishService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, err := service.ReadPublish(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("publish"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, record)
	})
}

func publishHistoryHandler(service PublishService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		records, err := service.PublishHistory(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"publishes": records})
	})
}
