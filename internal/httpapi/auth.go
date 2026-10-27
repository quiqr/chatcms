package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// AuthService is what the HTTP layer needs from whatever answers who someone
// is. The HTTP layer owns the cookies, the status codes and the wording; this
// interface owns none of them.
type AuthService interface {
	SignIn(ctx context.Context, email, password, caller string, remember bool) (Actor, string, time.Time, error)
	SignOut(ctx context.Context, token string) error
	Lookup(ctx context.Context, token string) (Actor, error)
	RequestPasswordReset(ctx context.Context, email, caller string) error
	SetPasswordWithToken(ctx context.Context, token, password string) (Actor, string, time.Time, error)
	SignOutEverywhere(ctx context.Context, who Actor) error
}

// These are the failures the HTTP layer tells apart. Everything else is an
// internal error.
var (
	ErrSignInRefused    = errors.New("that address and password do not match")
	ErrTooManyAttempts  = errors.New("too many attempts")
	ErrLinkUnusable     = errors.New("this link cannot be used any more")
	ErrPasswordRejected = errors.New("this password cannot be used")
)

// PasswordRejectedError carries why a password was refused, so the message
// the person reads says what to do rather than only that something is wrong.
type PasswordRejectedError struct {
	Reason string
}

func (e *PasswordRejectedError) Error() string {
	return e.Reason
}

func (e *PasswordRejectedError) Is(target error) bool {
	return target == ErrPasswordRejected
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type personResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type setPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type resetRequest struct {
	Email string `json:"email"`
}

func actorResponse(actor Actor) personResponse {
	return personResponse{ID: actor.ID, Email: actor.Email, DisplayName: actor.DisplayName}
}

func signInHandler(auth AuthService, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body signInRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		// A sign-in never continues a session it arrived with: handing someone
		// a session value and waiting for them to sign in with it is an attack
		// this closes.
		if cookie, err := r.Cookie(SessionCookie); err == nil && cookie.Value != "" {
			if err := auth.SignOut(r.Context(), cookie.Value); err != nil {
				WriteError(w, http.StatusInternalServerError, CodeInternal,
					"Something went wrong. Nothing was lost; please try again.")
				return
			}
		}

		actor, token, expires, err := auth.SignIn(
			r.Context(), body.Email, body.Password, callerOf(r), body.Remember)

		switch {
		case errors.Is(err, ErrTooManyAttempts):
			WriteError(w, http.StatusTooManyRequests, CodeTooManyAttempts,
				"Too many attempts. Wait a few minutes and try again.")
			return
		case errors.Is(err, ErrSignInRefused):
			WriteError(w, http.StatusUnauthorized, CodeUnauthorized,
				"That email address and password do not match.")
			return
		case err != nil:
			WriteError(w, http.StatusInternalServerError, CodeInternal,
				"Something went wrong. Nothing was lost; please try again.")
			return
		}

		SetSessionCookie(w, token, expires, secure)
		WriteJSON(w, http.StatusOK, actorResponse(actor))
	})
}

func signOutHandler(auth AuthService, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(SessionCookie); err == nil && cookie.Value != "" {
			if err := auth.SignOut(r.Context(), cookie.Value); err != nil {
				WriteError(w, http.StatusInternalServerError, CodeInternal,
					"Something went wrong. Nothing was lost; please try again.")
				return
			}
		}

		ClearSessionCookie(w, secure)
		WriteJSON(w, http.StatusNoContent, nil)
	})
}

func signOutEverywhereHandler(auth AuthService, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth.SignOutEverywhere(r.Context(), ActorFrom(r.Context())); err != nil {
			WriteError(w, http.StatusInternalServerError, CodeInternal,
				"Something went wrong. Nothing was lost; please try again.")
			return
		}

		// The session asking is one of the ones that just ended, so the
		// cookie goes with them.
		ClearSessionCookie(w, secure)
		w.WriteHeader(http.StatusNoContent)
	})
}

func currentPersonHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, actorResponse(ActorFrom(r.Context())))
	})
}

func requestResetHandler(auth AuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body resetRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		err := auth.RequestPasswordReset(r.Context(), body.Email, callerOf(r))
		if errors.Is(err, ErrTooManyAttempts) {
			WriteError(w, http.StatusTooManyRequests, CodeTooManyAttempts,
				"Too many attempts. Wait a few minutes and try again.")
			return
		}
		if err != nil {
			WriteError(w, http.StatusInternalServerError, CodeInternal,
				"Something went wrong. Nothing was lost; please try again.")
			return
		}

		// The same answer whether or not anyone has that address.
		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If that email address belongs to an account, a link is on its way to it.",
		})
	})
}

func setPasswordHandler(auth AuthService, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body setPasswordRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		actor, token, expires, err := auth.SetPasswordWithToken(r.Context(), body.Token, body.Password)
		switch {
		case errors.Is(err, ErrLinkUnusable):
			WriteError(w, http.StatusBadRequest, CodeBadRequest,
				"This link cannot be used any more. Ask for a new one.")
			return
		case errors.Is(err, ErrPasswordRejected):
			WriteError(w, http.StatusBadRequest, CodeBadRequest, passwordMessage(err))
			return
		case err != nil:
			WriteError(w, http.StatusInternalServerError, CodeInternal,
				"Something went wrong. Nothing was lost; please try again.")
			return
		}

		// Signed in straight away: somebody who has proved they hold the
		// address and chosen a password has done everything signing in would
		// ask for.
		SetSessionCookie(w, token, expires, secure)
		WriteJSON(w, http.StatusOK, actorResponse(actor))
	})
}

func passwordMessage(err error) string {
	var rejected *PasswordRejectedError
	if errors.As(err, &rejected) && rejected.Reason != "" {
		return capitalise(rejected.Reason) + "."
	}
	return "This password cannot be used."
}

func capitalise(text string) string {
	if text == "" {
		return text
	}
	runes := []rune(text)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] -= 32
	}
	return string(runes)
}

// callerOf names who is asking, for the purpose of limiting guessing. It is
// the connection's address; a proxy in front would have to supply it instead.
func callerOf(r *http.Request) string {
	host, _, found := splitHostPort(r.RemoteAddr)
	if !found {
		return r.RemoteAddr
	}
	return host
}

func splitHostPort(address string) (string, string, bool) {
	for i := len(address) - 1; i >= 0; i-- {
		if address[i] == ':' {
			return address[:i], address[i+1:], true
		}
	}
	return address, "", false
}
