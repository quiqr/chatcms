package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AccountService is what a person may change about themselves, and what
// somebody managing people may do to an account.
type AccountService interface {
	ChangeOwnName(ctx context.Context, who Actor, displayName string) (Actor, error)
	ChangeOwnPassword(ctx context.Context, who Actor, current, next string) error
	TurnAccountOff(ctx context.Context, who Actor, personID uuid.UUID) error
	TurnAccountOn(ctx context.Context, personID uuid.UUID) error

	SetPicture(ctx context.Context, who Actor, content []byte) error
	RemovePicture(ctx context.Context, who Actor) error
	ReadPicture(ctx context.Context, personID string) ([]byte, string, bool, error)

	AskForANewAddress(ctx context.Context, who Actor, address string) error

	SessionViews(ctx context.Context, who Actor, thisToken string) ([]SessionView, error)
	EndSession(ctx context.Context, who Actor, sessionID string) error
}

// SessionView is one place somebody is signed in.
type SessionView struct {
	ID       string    `json:"id"`
	LastSeen time.Time `json:"lastSeen"`
	Started  time.Time `json:"started"`
	ThisOne  bool      `json:"thisOne"`
}

type newAddressRequest struct {
	Email string `json:"email"`
}

// LargestPicture is the most a request may carry, refused before it is read
// rather than after.
const LargestPicture = 5 << 20

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

	mux.Handle("PUT /api/account/picture", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, LargestPicture)

			file, _, err := r.FormFile("picture")
			if err != nil {
				WriteError(w, http.StatusBadRequest, CodeBadRequest,
					"That picture could not be read, or it is too large.")
				return
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				WriteError(w, http.StatusBadRequest, CodeBadRequest,
					"That picture could not be read, or it is too large.")
				return
			}

			if err := service.SetPicture(r.Context(), ActorFrom(r.Context()), content); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.Handle("DELETE /api/account/picture", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.RemovePicture(r.Context(), ActorFrom(r.Context())); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc("/api/account/picture", MethodNotAllowed)

	// Anybody signed in may see anybody's picture: it is shown beside their
	// name wherever they appear, and hiding it from colleagues would make
	// the screens useless without hiding anything worth hiding.
	mux.Handle("GET /api/people/{person}/picture", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			content, kind, has, err := service.ReadPicture(r.Context(), r.PathValue("person"))
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			if !has {
				WriteError(w, http.StatusNotFound, CodeNotFound, "There is nothing here.")
				return
			}

			w.Header().Set("Content-Type", kind)
			w.Header().Set("Cache-Control", "private, max-age=60")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
		})))

	mux.Handle("POST /api/account/address", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var body newAddressRequest
			if err := Decode(w, r, &body); err != nil {
				WriteDecodeError(w, err)
				return
			}

			if err := service.AskForANewAddress(r.Context(), ActorFrom(r.Context()), body.Email); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))

	mux.HandleFunc("/api/account/address", MethodNotAllowed)

	mux.Handle("GET /api/account/sessions", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			here := ""
			if cookie, err := r.Cookie(SessionCookie); err == nil {
				here = cookie.Value
			}

			sessions, err := service.SessionViews(r.Context(), ActorFrom(r.Context()), here)
			if err != nil {
				writeTenancyError(w, err)
				return
			}
			WriteJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
		})))

	mux.HandleFunc("/api/account/sessions", MethodNotAllowed)

	mux.Handle("DELETE /api/account/sessions/{session}", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := service.EndSession(r.Context(), ActorFrom(r.Context()), r.PathValue("session")); err != nil {
				writeTenancyError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})))
}

func isPasswordRejected(err error) bool {
	var rejected *PasswordRejectedError
	return errors.As(err, &rejected)
}
