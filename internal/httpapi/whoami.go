package httpapi

import (
	"context"
	"net/http"
)

// WhoAmIView is everything a screen needs before it can draw its shell: who
// is looking, what they may do over the installation, and which
// organisations they may switch to. One answer, because three requests per
// navigation would be three chances to half-load.
type WhoAmIView struct {
	Person        PersonView         `json:"person"`
	Administers   bool               `json:"administers"`
	Organisations []OrganisationView `json:"organisations"`
	// Appearance is what this person chose, or empty when they have chosen
	// nothing and the machine decides. It rides along here because the
	// interface needs it before it draws anything.
	Appearance string `json:"appearance"`
}

// PersonView is somebody as a screen shows them.
type PersonView struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// WhoAmIService answers that one question.
type WhoAmIService interface {
	WhoAmI(ctx context.Context, who Actor) (WhoAmIView, error)
}

func whoAmIHandler(service WhoAmIService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		answer, err := service.WhoAmI(r.Context(), ActorFrom(r.Context()))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, answer)
	})
}
