package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// SessionCookie carries the value that identifies a session. Page scripts
// cannot read it.
const SessionCookie = "chatcms_session"

// RequestTokenCookie carries the value a state-changing request has to repeat
// in a header. Page scripts can read this one, which is the point: another
// site cannot.
const RequestTokenCookie = "chatcms_request_token"

// RequestTokenHeader is where the repeated value goes.
const RequestTokenHeader = "X-Chatcms-Request-Token"

// Actor is who a request is from.
type Actor struct {
	ID          string
	Email       string
	DisplayName string
	Anonymous   bool
}

func Anonymous() Actor {
	return Actor{DisplayName: "anonymous", Anonymous: true}
}

type actorKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFrom(ctx context.Context) Actor {
	if actor, ok := ctx.Value(actorKey{}).(Actor); ok {
		return actor
	}
	return Anonymous()
}

// LookupSession turns the value a cookie carries into the person it
// identifies. It returns the anonymous actor when the value identifies
// nobody, which is not an error.
type LookupSession func(ctx context.Context, token string) (Actor, error)

// WithSession puts the person a request's session identifies into its
// context, so that no handler reads a cookie.
func WithSession(lookup LookupSession, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := Anonymous()

			if cookie, err := r.Cookie(SessionCookie); err == nil && lookup != nil {
				found, err := lookup(r.Context(), cookie.Value)
				switch {
				case err != nil:
					logger.LogAttrs(r.Context(), slog.LevelError, "reading the session failed",
						slog.String("request_id", RequestIDFrom(r.Context())),
						slog.Any("error", err))
				case !found.Anonymous:
					actor = found
				}
			}

			next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), actor)))
		})
	}
}

// RequireActor refuses a request that carries nobody.
func RequireActor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ActorFrom(r.Context()).Anonymous {
			WriteError(w, http.StatusUnauthorized, CodeUnauthorized, "Please sign in first.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetSessionCookie puts a session value in the browser. It is unreadable by
// page scripts, restricted to this site, and marked secure whenever the site
// is served over a secure connection.
func SetSessionCookie(w http.ResponseWriter, value string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie removes it.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
