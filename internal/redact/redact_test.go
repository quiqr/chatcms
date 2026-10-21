package redact

import (
	"strings"
	"testing"
)

func TestSecretsRemovesCredentials(t *testing.T) {
	cases := map[string]struct {
		in   string
		gone string
	}{
		"a token in a URL": {"could not read from https://user:ghp_secret123@github.com/a/b.git", "ghp_secret123"},
		"a user in a URL":  {"cloning ssh://deploykey@example.org/site.git", "deploykey"},
		"a named token":    {"remote: rejected; token=abc123def", "abc123def"},
		"a named password": {"fatal: password: hunter2", "hunter2"},
		"an api key":       {"error: api_key = sk-live-9999", "sk-live-9999"},
		"a bearer header":  {"Authorization: Bearer eyJhbGciOi", "eyJhbGciOi"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := Secrets(tc.in)
			if strings.Contains(got, tc.gone) {
				t.Errorf("Secrets(%q) = %q, want %q gone", tc.in, got, tc.gone)
			}
			if !strings.Contains(got, Placeholder) {
				t.Errorf("Secrets(%q) = %q, want it to say something was removed", tc.in, got)
			}
		})
	}
}

func TestSecretsKeepsOrdinaryText(t *testing.T) {
	for _, message := range []string{
		"fatal: couldn't find remote ref refs/heads/main",
		"the page About us could not be saved",
		"connection refused",
	} {
		if got := Secrets(message); got != message {
			t.Errorf("Secrets(%q) = %q, want it untouched", message, got)
		}
	}
}

func TestSecretsLeavesAPlainURL(t *testing.T) {
	message := "cloning https://github.com/quiqr/chatcms.git"
	if got := Secrets(message); got != message {
		t.Errorf("Secrets(%q) = %q, want a URL with no credential untouched", message, got)
	}
}

func TestSecretsHandlesAMatchWithNoSeparator(t *testing.T) {
	// A bearer token with neither a colon nor an equals sign still loses its
	// value, and keeps the word that labelled it.
	got := Secrets("sent bearer abc123xyz to the remote")

	if strings.Contains(got, "abc123xyz") {
		t.Errorf("Secrets() = %q, want the token gone", got)
	}
	if !strings.Contains(strings.ToLower(got), "bearer") {
		t.Errorf("Secrets() = %q, want the label kept so a reader knows what was removed", got)
	}
}

func TestSecretsLeavesAnEmptyStringAlone(t *testing.T) {
	if got := Secrets(""); got != "" {
		t.Errorf("Secrets(\"\") = %q, want an empty string", got)
	}
}
