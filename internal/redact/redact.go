// Package redact removes anything credential-shaped from text that is about
// to be stored, logged or shown. A message from an external tool can carry a
// URL with a token in it, and that message ends up in a job's recorded reason
// and in an error an implementer reads.
package redact

import (
	"regexp"
	"strings"
)

// Placeholder is what replaces a secret, chosen so a reader can see that
// something was removed rather than wondering what is missing.
const Placeholder = "[redacted]"

var credentialPattern = regexp.MustCompile(
	// A userinfo section of a URL, a labelled credential with its value, and a
	// bare bearer token. The labelled case swallows an authentication scheme
	// word so that "Authorization: Bearer <token>" loses the token and not
	// just the word Bearer.
	`(?i)([a-z][a-z0-9+.-]*://)[^/\s@]+(:[^/\s@]*)?@` +
		`|\b(token|password|passwd|secret|api[_-]?key|access[_-]?key|authorization)\b` +
		`\s*[:=]?\s*(?:bearer\s+|basic\s+)?\S+` +
		`|\bbearer\s+\S+`)

// Secrets returns the text with every credential-shaped run replaced.
func Secrets(text string) string {
	return credentialPattern.ReplaceAllStringFunc(text, func(match string) string {
		if index := strings.Index(match, "://"); index >= 0 {
			return match[:index+3] + Placeholder + "@"
		}
		if separator := strings.IndexAny(match, ":="); separator >= 0 {
			return match[:separator+1] + " " + Placeholder
		}
		word := strings.Fields(match)
		if len(word) == 0 {
			return Placeholder
		}
		return word[0] + " " + Placeholder
	})
}
