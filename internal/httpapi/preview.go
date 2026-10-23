package httpapi

import (
	"context"
	"net/http"
	"strings"
)

// PreviewService is what the HTTP layer needs in order to show a site as a
// reader will see it.
type PreviewService interface {
	// PreviewHandler gives what answers requests for one site's preview, or
	// says why it cannot.
	PreviewHandler(ctx context.Context, actor Actor, organisation, project string) (http.Handler, error)
}

// PreviewPrefix is where a preview is reached. It is under the app's own
// address on purpose: the app is what asks who somebody is, so a preview
// cannot be seen by anybody who may not see the project.
const PreviewPrefix = "/preview/"

func previewHandler(service PreviewService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organisation, project, rest, ok := splitPreviewPath(r.URL.Path)
		if !ok {
			NotFound(w, r)
			return
		}

		handler, err := service.PreviewHandler(r.Context(), ActorFrom(r.Context()), organisation, project)
		if err != nil {
			writeTenancyError(w, err)
			return
		}

		// The preview is handed the path within the site, so that a site
		// built for its own root is served as it was built.
		forward := r.Clone(r.Context())
		forward.URL.Path = rest
		forward.RequestURI = ""
		handler.ServeHTTP(w, forward)
	})
}

// splitPreviewPath reads /preview/<organisation>/<project>/<the rest>.
func splitPreviewPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, PreviewPrefix)
	if rest == path {
		return "", "", "", false
	}

	organisation, rest, found := strings.Cut(rest, "/")
	if !found || organisation == "" {
		return "", "", "", false
	}

	project, within, found := strings.Cut(rest, "/")
	if project == "" {
		return "", "", "", false
	}
	if !found {
		return organisation, project, "/", true
	}
	return organisation, project, "/" + within, true
}
