package media

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"
)

// Image is one picture in a site, as a person sees it.
type Image struct {
	// Path is where it lives in the site. It is what a page refers to and
	// what a request names.
	Path string
	Name string
	Size int64
	By   string
	At   time.Time
	// Alt is the description somebody wrote for a screen reader.
	Alt string
	// UsedOn is the pages that refer to it, by the key the content browser
	// uses, so that "used on" can be a link.
	UsedOn []string
}

// Used says whether anything refers to this picture.
func (i Image) Used() bool { return len(i.UsedOn) > 0 }

var (
	// ErrNotAnImage says the file is not a kind of picture ChatCMS will take.
	ErrNotAnImage = errors.New("media: that is not a picture ChatCMS can use")
	// ErrTooLarge says the file is bigger than a site should carry.
	ErrTooLarge = errors.New("media: that picture is too large")
	// ErrNoName says the file has no name anything could be made from.
	ErrNoName = errors.New("media: that file has no usable name")
)

// MaxBytes is the largest picture a site will take. A photograph from a phone
// is a few megabytes; anything much larger is a mistake rather than a choice,
// and git is a poor place to keep it.
const MaxBytes = 25 << 20

// kinds are the pictures ChatCMS will take, by what the bytes say they are.
// The name is not trusted: a file called .png that is not one would be shown
// as a broken picture on the live site.
var kinds = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// KindOf says what a picture is from its own bytes, and what it should be
// called. SVG is text, so it is recognised by its own opening rather than by
// sniffing, which reports it as plain text or XML.
func KindOf(content []byte) (string, string, error) {
	if looksLikeSVG(content) {
		return "image/svg+xml", ".svg", nil
	}

	kind, _, _ := strings.Cut(http.DetectContentType(content), ";")
	extension, known := kinds[kind]
	if !known {
		return "", "", fmt.Errorf("%w: it looks like %s", ErrNotAnImage, kind)
	}
	return kind, extension, nil
}

func looksLikeSVG(content []byte) bool {
	head := strings.ToLower(string(content[:min(len(content), 1024)]))
	return strings.Contains(head, "<svg")
}

// Name turns what a file was called into a name a site can carry: small
// letters, digits, hyphens and one extension, which is the shape the old
// Quiqr uses for everything else too.
func Name(said, extension string) (string, error) {
	stem := strings.TrimSuffix(said, path.Ext(said))

	out := strings.Builder{}
	previousHyphen := false
	for _, letter := range strings.ToLower(stem) {
		switch {
		case (unicode.IsLetter(letter) || unicode.IsDigit(letter)) && letter < unicode.MaxASCII:
			out.WriteRune(letter)
			previousHyphen = false
		case !previousHyphen && out.Len() > 0:
			out.WriteRune('-')
			previousHyphen = true
		}
	}

	cleaned := strings.Trim(out.String(), "-")
	if cleaned == "" {
		return "", ErrNoName
	}
	return cleaned + extension, nil
}

// Folder is where a site keeps its pictures. Hugo serves everything under
// static as it is, and the old Quiqr puts uploads in a folder of their own
// inside it; a site that already has one keeps it.
const Folder = "static/images"

var otherFolders = []string{"static/uploads", "static/img", "assets/images"}

// FolderIn is the folder a site already uses, or the usual one where it has
// none yet.
func FolderIn(fsys fs.FS) string {
	for _, folder := range append([]string{Folder}, otherFolders...) {
		if info, err := fs.Stat(fsys, folder); err == nil && info.IsDir() {
			return folder
		}
	}
	return Folder
}

// Pictures lists the pictures in a folder, by name so that the order is the
// same every time.
func Pictures(fsys fs.FS, folder string) []string {
	entries, err := fs.ReadDir(fsys, folder)
	if err != nil {
		return nil
	}

	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if _, known := kinds[kindByName(entry.Name())]; known {
			names = append(names, entry.Name())
		}
	}
	return names
}

func kindByName(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return ""
	}
}

// KindByName is what a picture's name says it is, for serving it back.
func KindByName(name string) string { return kindByName(name) }
