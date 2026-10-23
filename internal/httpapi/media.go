package httpapi

import (
	"context"
	"net/http"
	"strconv"
)

// MediaService is what the HTTP layer needs in order to show, add, describe
// and remove the pictures a site carries.
type MediaService interface {
	ListMedia(ctx context.Context, actor Actor, organisation, project string) ([]ImageView, error)
	AddMedia(ctx context.Context, actor Actor, organisation, project, name string, content []byte) (ImageView, error)
	DescribeMedia(ctx context.Context, actor Actor, organisation, project, name, said string) (ImageView, error)
	RemoveMedia(ctx context.Context, actor Actor, organisation, project, name string, evenIfUsed bool) error
	ReadMedia(ctx context.Context, actor Actor, organisation, project, name string, small bool) ([]byte, string, error)
}

// ImageView is one picture as a person sees it.
type ImageView struct {
	Name string `json:"name"`
	// Path is where it lives in the site, which is what a page refers to.
	Path string `json:"path"`
	Size int64  `json:"size"`
	By   string `json:"by,omitempty"`
	At   string `json:"at,omitempty"`
	// Alt is what somebody wrote about what it shows.
	Alt string `json:"alt,omitempty"`
	// UsedOn is the pages that refer to it.
	UsedOn []string `json:"usedOn,omitempty"`
}

type describeRequest struct {
	Alt string `json:"alt"`
}

// MaxUpload is the largest request the pictures endpoint will read. It is a
// little over what a picture may be, to leave room for the form around it.
const MaxUpload = 26 << 20

func listMediaHandler(service MediaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pictures, err := service.ListMedia(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"images": pictures})
	})
}

func addMediaHandler(service MediaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxUpload)

		if err := r.ParseMultipartForm(8 << 20); err != nil {
			WriteError(w, http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
				"That picture is too large to send.")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			WriteError(w, http.StatusBadRequest, CodeBadRequest, "No picture was sent.")
			return
		}
		defer file.Close()

		content := make([]byte, header.Size)
		if _, err := readFully(file, content); err != nil {
			WriteError(w, http.StatusBadRequest, CodeBadRequest, "That picture could not be read.")
			return
		}

		picture, err := service.AddMedia(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), header.Filename, content)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, picture)
	})
}

func readFully(from interface{ Read([]byte) (int, error) }, into []byte) (int, error) {
	read := 0
	for read < len(into) {
		count, err := from.Read(into[read:])
		read += count
		if err != nil {
			if read == len(into) {
				return read, nil
			}
			return read, err
		}
	}
	return read, nil
}

func describeMediaHandler(service MediaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body describeRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}

		picture, err := service.DescribeMedia(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("image"), body.Alt)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, picture)
	})
}

func removeMediaHandler(service MediaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		evenIfUsed, _ := strconv.ParseBool(r.URL.Query().Get("evenIfUsed"))

		err := service.RemoveMedia(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("image"), evenIfUsed)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func readMediaHandler(service MediaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		small, _ := strconv.ParseBool(r.URL.Query().Get("small"))

		content, kind, err := service.ReadMedia(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("image"), small)
		if err != nil {
			writeTenancyError(w, err)
			return
		}

		w.Header().Set("Content-Type", kind)
		w.Header().Set("Cache-Control", "private, max-age=60")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(content)
	})
}
