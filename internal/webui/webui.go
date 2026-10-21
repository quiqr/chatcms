package webui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distRoot embed.FS

const (
	documentName    = "index.html"
	assetsPrefix    = "assets/"
	immutableMaxAge = "public, max-age=31536000, immutable"
	documentCache   = "no-cache"
)

func FS() fs.FS {
	sub, err := fs.Sub(distRoot, "dist")
	if err != nil {
		panic(fmt.Sprintf("embedded frontend is unreadable: %v", err))
	}
	return sub
}

func Built() bool {
	return builtFS(FS())
}

func builtFS(assets fs.FS) bool {
	_, err := fs.Stat(assets, documentName)
	return err == nil
}

func Handler() http.Handler {
	return HandlerFor(FS())
}

func HandlerFor(assets fs.FS) http.Handler {
	built := builtFS(assets)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "This address only serves pages.", http.StatusMethodNotAllowed)
			return
		}

		if !built {
			serveNotice(w)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			serveDocument(w, r, assets)
			return
		}

		file, err := assets.Open(name)
		if err != nil {
			serveDocument(w, r, assets)
			return
		}
		file.Close()

		if strings.HasPrefix(name, assetsPrefix) {
			w.Header().Set("Cache-Control", immutableMaxAge)
		} else {
			w.Header().Set("Cache-Control", documentCache)
		}
		http.FileServerFS(assets).ServeHTTP(w, r)
	})
}

func serveDocument(w http.ResponseWriter, r *http.Request, assets fs.FS) {
	document, err := fs.ReadFile(assets, documentName)
	if err != nil {
		serveNotice(w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", documentCache)
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		w.Write(document)
	}
}

const notice = `<!doctype html>
<meta charset="utf-8">
<title>Quiqr ChatCMS</title>
<h1>The frontend was not built</h1>
<p>This server was built without the browser application, so there is nothing
to show here. The API is running and answering at <code>/api/health</code>.</p>
<p>Run <code>npm run build</code> in <code>web/</code> and copy
<code>web/dist</code> into <code>internal/webui/dist</code>, or build the
server with <code>nix build .#chatcmsd</code>, which does both.</p>
`

func serveNotice(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", documentCache)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(notice))
}
