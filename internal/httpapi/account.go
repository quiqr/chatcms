package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// AccountService is what a person may change about themselves, and what
// somebody managing people may do to an account.
type AccountService interface {
	ChangeOwnName(ctx context.Context, who Actor, displayName string) (Actor, error)
	ChangeOwnPassword(ctx context.Context, who Actor, current, next string) error
	TurnAccountOff(ctx context.Context, who Actor, personID uuid.UUID) error
	TurnAccountOn(ctx context.Context, personID uuid.UUID) error
}

type changeNameRequest struct {
	DisplayName string `json:"displayName"`
}

type changePasswordRequest struct {
	Current string `json:"current"`
	Next    string `json:"next"`
}

func accountHandlers(mux *http.ServeMux, service AccountService) {
	mux.Handle("PUT /api/account/name", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body changeNameRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}

			changed, err := service.ChangeOwnName(r.Context(), ActorFrom(r.Context()), body.DisplayName)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, actorResponse(changed))
		})))

	mux.HandleFunc("/api/account/name", MethodNotAllowed)

	mux.Handle("PUT /api/account/password", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body changePasswordRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}

			err := service.ChangeOwnPassword(r.Context(), ActorFrom(r.Context()), body.Current, body.Next)
			switch {
			case err == nil:
				w.WriteHeader(http.StatusNoContent)
			case isPasswordRejected(err):
				WriteError(w, http.StatusBadRequest, CodeBadRequest, passwordMessage(err))
			default:
				writeTenancyError(w, err)
			}
		})))

	mux.HandleFunc("/api/account/password", MethodNotAllowed)
}

func isPasswordRejected(err error) bool {
	var rejected *PasswordRejectedError
	return errors.As(err, &rejected)
}
