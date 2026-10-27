// Package imagekind says what a picture is from its own bytes. It is its own
// package because two things need the answer and neither should depend on
// the other: a site's media library, and a person's picture on their
// account.
package imagekind

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrNotAnImage says the bytes are not a kind of picture ChatCMS will take.
var ErrNotAnImage = errors.New("that is not a picture ChatCMS can use")

// Kinds are the pictures ChatCMS will take, by what the bytes say they are.
// The name is not trusted: a file called .png that is not one would be shown
// as a broken picture on the live site.
var Kinds = map[string]string{
	"image/png":     ".png",
	"image/jpeg":    ".jpg",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
}

// Of says what a picture is from its own bytes, and what it should be
// called. SVG is text, so it is recognised by its own opening rather than by
// sniffing, which reports it as plain text or XML.
func Of(content []byte) (string, string, error) {
	if looksLikeSVG(content) {
		return "image/svg+xml", ".svg", nil
	}

	kind, _, _ := strings.Cut(http.DetectContentType(content), ";")
	extension, known := Kinds[kind]
	if !known {
		return "", "", fmt.Errorf("%w: it looks like %s", ErrNotAnImage, kind)
	}
	return kind, extension, nil
}

func looksLikeSVG(content []byte) bool {
	head := strings.ToLower(string(content[:min(len(content), 1024)]))
	return strings.Contains(head, "<svg")
}
