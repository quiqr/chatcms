package httpapi

import (
	"context"
	"net/http"
)

// HistoryService is what the HTTP layer needs in order to show the history of
// a site and to go back to an earlier version.
type HistoryService interface {
	Timeline(ctx context.Context, actor Actor, organisation, project string) (TimelineView, error)
	VersionContents(ctx context.Context, actor Actor, organisation, project, version string) (ReviewView, error)
	BeforeGoingBack(ctx context.Context, actor Actor, organisation, project, version string) (WarningView, error)
	GoBackTo(ctx context.Context, actor Actor, organisation, project, version string, publishRightAway bool) (WentBackView, error)
}

// TimelineView is the history of a site, grouped into stretches of time.
type TimelineView struct {
	Unpublished int         `json:"unpublished"`
	Groups      []GroupView `json:"groups"`
}

type GroupView struct {
	Period  string        `json:"period"`
	Entries []VersionView `json:"entries"`
}

// VersionView is one version as a person sees it. The identifier is what a
// request names, never what a person reads.
type VersionView struct {
	Version string `json:"version"`
	Title   string `json:"title"`
	By      string `json:"by,omitempty"`
	At      string `json:"at,omitempty"`
	Changes int    `json:"changes"`
	Live    bool   `json:"live"`
}

// WarningView is what somebody is told before they go back.
type WarningView struct {
	Title string `json:"title"`
	When  string `json:"when,omitempty"`
	// SetAside is how much unpublished work will stop being in the site. It
	// is not lost, which is what Reassurance says in words.
	SetAside    int    `json:"setAside"`
	Reassurance string `json:"reassurance"`
}

// WentBackView is what happened when somebody went back.
type WentBackView struct {
	Title     string       `json:"title"`
	Published *PublishView `json:"published,omitempty"`
}

type goBackRequest struct {
	PublishRightAway *bool `json:"publishRightAway"`
}

func timelineHandler(service HistoryService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeline, err := service.Timeline(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, timeline)
	})
}

func versionContentsHandler(service HistoryService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contents, err := service.VersionContents(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("version"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, contents)
	})
}

func beforeGoingBackHandler(service HistoryService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		warning, err := service.BeforeGoingBack(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("version"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, warning)
	})
}

func goBackHandler(service HistoryService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body goBackRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		// Publishing right away is on unless somebody turns it off, which is
		// what the confirmation dialog draws.
		publishRightAway := body.PublishRightAway == nil || *body.PublishRightAway

		result, err := service.GoBackTo(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("version"), publishRightAway)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, result)
	})
}
