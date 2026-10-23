package httpapi

import (
	"context"
	"net/http"
)

// ContentService is what the HTTP layer needs in order to show one page as a
// form and to write an edit back.
type ContentService interface {
	ReadPage(ctx context.Context, actor Actor, organisation, project, page string) (PageView, error)
	SavePage(ctx context.Context, actor Actor, organisation, project, page string, edit EditRequest) (PageView, error)
}

// PageView is one page as a form: what to draw, and what is in it now.
type PageView struct {
	Key     string        `json:"key"`
	Title   string        `json:"title"`
	Section string        `json:"section"`
	Fields  []ControlView `json:"fields"`
	Body    string        `json:"body"`
	// HasBody says the page's own text can be edited. Where it cannot, the
	// text is still shown, because leaving it out would make a page look
	// emptier than it is.
	HasBody bool `json:"hasBody"`
}

// ControlView is one thing a person can change.
type ControlView struct {
	Path    string         `json:"path"`
	Type    string         `json:"type"`
	Title   string         `json:"title"`
	Tip     string         `json:"tip,omitempty"`
	Value   any            `json:"value,omitempty"`
	Options []OptionView   `json:"options,omitempty"`
	Fields  []ControlView  `json:"fields,omitempty"`
	Config  map[string]any `json:"config,omitempty"`
	// Unknown says this version has no control for that kind of field, so it
	// is shown as plain text rather than left out.
	Unknown bool `json:"unknown,omitempty"`
}

type OptionView struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// EditRequest is what somebody changed.
type EditRequest struct {
	Values map[string]any `json:"values"`
	Body   *string        `json:"body"`
}

func readPageHandler(service ContentService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := service.ReadPage(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("page"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, page)
	})
}

func savePageHandler(service ContentService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body EditRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		page, err := service.SavePage(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("page"), body)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, page)
	})
}
