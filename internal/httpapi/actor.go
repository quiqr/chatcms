package httpapi

import (
	"context"
	"net/http"
)

type Actor struct {
	ID          string
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

func AnonymousActor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), Anonymous())))
	})
}
