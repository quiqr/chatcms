package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestStringCarriesVersionAndPlatform(t *testing.T) {
	got := String()

	for _, want := range []string{"chatcmsd", Version, runtime.GOOS, runtime.GOARCH} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}

func TestVersionDefaultsToDev(t *testing.T) {
	if Version == "" {
		t.Error("Version must never be empty; the build stamps it via ldflags")
	}
}
