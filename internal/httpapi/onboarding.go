package httpapi

import (
	"context"
	"net/http"
)

// OnboardingService is what the HTTP layer needs in order to set a project
// up: a key, a site, and where it publishes.
type OnboardingService interface {
	GenerateKey(ctx context.Context, actor Actor, organisation, project string) (KeyView, error)
	ReadKey(ctx context.Context, actor Actor, organisation, project string) (KeyView, error)
	ImportSite(ctx context.Context, actor Actor, organisation, project, kind, address, remote string) (ImportView, error)
	InspectSite(ctx context.Context, actor Actor, organisation, project string) (ImportView, error)
	TestConnection(ctx context.Context, actor Actor, organisation, project string) (ConnectionView, error)
	SetPublishing(ctx context.Context, actor Actor, organisation, project, to, branch, form string) (PublishingView, error)
	ReadPublishing(ctx context.Context, actor Actor, organisation, project string) (PublishingView, error)
}

// KeyView is the half of a deploy key that may be shown. The other half is
// never in any answer.
type KeyView struct {
	PublicKey string `json:"publicKey"`
	// Advice says what to do with it.
	Advice string `json:"advice"`
}

// ImportView says what arrived and what, if anything, is still missing.
type ImportView struct {
	Kind        string   `json:"kind"`
	Remote      string   `json:"remote,omitempty"`
	IsHugoSite  bool     `json:"isHugoSite"`
	HasModel    bool     `json:"hasModel"`
	Ready       bool     `json:"ready"`
	Pages       int      `json:"pages"`
	Collections int      `json:"collections"`
	Problem     string   `json:"problem,omitempty"`
	LookedFor   []string `json:"lookedFor,omitempty"`
}

// ConnectionView is the answer to "does this key reach the host".
type ConnectionView struct {
	Good      bool   `json:"good"`
	Message   string `json:"message"`
	PublicKey string `json:"publicKey,omitempty"`
}

// PublishingView is where and in what form a project publishes.
type PublishingView struct {
	To     string `json:"to"`
	Branch string `json:"branch,omitempty"`
	Form   string `json:"form"`
}

type importRequest struct {
	Kind    string `json:"kind"`
	Address string `json:"address"`
	Remote  string `json:"remote"`
}

type publishingRequest struct {
	To     string `json:"to"`
	Branch string `json:"branch"`
	Form   string `json:"form"`
}

func generateKeyHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, err := service.GenerateKey(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, key)
	})
}

func readKeyHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, err := service.ReadKey(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, key)
	})
}

func importSiteHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body importRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		result, err := service.ImportSite(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Kind, body.Address, body.Remote)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, result)
	})
}

func inspectSiteHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := service.InspectSite(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, result)
	})
}

func testConnectionHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := service.TestConnection(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, connection)
	})
}

func readPublishingHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publishing, err := service.ReadPublishing(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, publishing)
	})
}

func setPublishingHandler(service OnboardingService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body publishingRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		publishing, err := service.SetPublishing(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.To, body.Branch, body.Form)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, publishing)
	})
}
