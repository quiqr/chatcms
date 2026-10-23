package httpapi

import (
	"context"
	"net/http"
)

// BrowseService is what the HTTP layer needs in order to show what a site
// holds and to make, rename, copy and remove pages.
type BrowseService interface {
	ListContent(ctx context.Context, actor Actor, organisation, project string) ([]SectionView, error)
	CreatePage(ctx context.Context, actor Actor, organisation, project, section, title string) (PageView, error)
	RenamePage(ctx context.Context, actor Actor, organisation, project, page, title string) (PageView, error)
	DuplicatePage(ctx context.Context, actor Actor, organisation, project, page, title string) (PageView, error)
	DeletePage(ctx context.Context, actor Actor, organisation, project, page string) error
}

// SectionView is one group in the content browser.
type SectionView struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	// ItemName is the singular word for one of its items, used in "New
	// article". It is empty for the pages group, which cannot grow.
	ItemName string     `json:"itemName,omitempty"`
	Count    int        `json:"count"`
	Items    []ItemView `json:"items"`
}

// ItemView is one page in the browser.
type ItemView struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	// Waiting says this page has a change nobody has published yet.
	Waiting bool   `json:"waiting,omitempty"`
	By      string `json:"by,omitempty"`
	At      string `json:"at,omitempty"`
}

type newPageRequest struct {
	Section string `json:"section"`
	Title   string `json:"title"`
}

type renameRequest struct {
	Title string `json:"title"`
	// Copy says to leave the page where it is and make another.
	Copy bool `json:"copy"`
}

func listContentHandler(service BrowseService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sections, err := service.ListContent(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"sections": sections})
	})
}

func createPageHandler(service BrowseService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body newPageRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		page, err := service.CreatePage(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Section, body.Title)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, page)
	})
}

func renamePageHandler(service BrowseService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body renameRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		who := ActorFrom(r.Context())
		organisation, project := r.PathValue("organisation"), r.PathValue("project")

		var page PageView
		var err error
		if body.Copy {
			page, err = service.DuplicatePage(r.Context(), who, organisation, project, r.PathValue("page"), body.Title)
		} else {
			page, err = service.RenamePage(r.Context(), who, organisation, project, r.PathValue("page"), body.Title)
		}
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, page)
	})
}

func deletePageHandler(service BrowseService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := service.DeletePage(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("page"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
