package httpapi

import (
	"context"
	"net/http"
)

// SyncService is what the HTTP layer needs in order to say how a site stands
// and to settle what two people changed.
type SyncService interface {
	SiteStanding(ctx context.Context, actor Actor, organisation, project string) (StandingView, error)
	BringChanges(ctx context.Context, actor Actor, organisation, project string) (StandingView, error)
	AnswerChanges(ctx context.Context, actor Actor, organisation, project string, answers []AnswerRequest) (StandingView, error)
	AbandonChanges(ctx context.Context, actor Actor, organisation, project string) (StandingView, error)
}

// StandingView is how a site stands, for the indicator and the sync screen.
type StandingView struct {
	// State is one of saved, saving, offline, arrived or asking.
	State string `json:"state"`
	// Said is the same thing in words a content creator reads.
	Said        string `json:"said"`
	Unpublished int    `json:"unpublished"`
	// ArrivedFrom names whoever else changed the site.
	ArrivedFrom []string `json:"arrivedFrom,omitempty"`
	// Conflicts is what two people changed, as pages and fields.
	Conflicts []PageConflictView `json:"conflicts,omitempty"`
	// Technical is only ever filled in for somebody the policy function
	// allows technical detail. It is the one place a git word belongs.
	Technical *TechnicalView `json:"technical,omitempty"`
}

type TechnicalView struct {
	Branch string `json:"branch"`
	Remote string `json:"remote"`
	Head   string `json:"head"`
}

// PageConflictView is one page two people changed.
type PageConflictView struct {
	Page  string `json:"page,omitempty"`
	Title string `json:"title"`
	// File is what an answer names. It is never shown.
	File      string              `json:"file"`
	Who       string              `json:"who,omitempty"`
	WholeFile bool                `json:"wholeFile,omitempty"`
	Fields    []FieldConflictView `json:"fields,omitempty"`
}

// FieldConflictView is one field two people changed, with both answers.
type FieldConflictView struct {
	Path   string     `json:"path"`
	Name   string     `json:"name"`
	Mine   string     `json:"mine"`
	Theirs string     `json:"theirs"`
	Words  []WordView `json:"words,omitempty"`
}

// AnswerRequest is which version to keep for one page.
type AnswerRequest struct {
	File   string            `json:"file"`
	Fields map[string]string `json:"fields,omitempty"`
	Whole  string            `json:"whole,omitempty"`
}

type answersRequest struct {
	Answers []AnswerRequest `json:"answers"`
}

func siteStandingHandler(service SyncService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		standing, err := service.SiteStanding(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, standing)
	})
}

func bringChangesHandler(service SyncService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		standing, err := service.BringChanges(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, standing)
	})
}

func answerChangesHandler(service SyncService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body answersRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		standing, err := service.AnswerChanges(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Answers)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, standing)
	})
}

func abandonChangesHandler(service SyncService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		standing, err := service.AbandonChanges(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, standing)
	})
}
