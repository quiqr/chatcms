package preview_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quiqr/chatcms/internal/preview"
)

func hugoBin(t *testing.T) string {
	t.Helper()

	if named := os.Getenv("CHATCMS_HUGO_BIN"); named != "" {
		return named
	}
	found, err := exec.LookPath("hugo")
	if err != nil {
		t.Skipf("no Hugo on this machine: %v", err)
	}
	return found
}

// aSite writes a Hugo site that renders, so that what is tested is a preview
// of a site rather than a process that happens to listen.
func aSite(t *testing.T, title string) string {
	t.Helper()

	dir := t.TempDir()
	files := map[string]string{
		"hugo.toml":          "baseURL = \"https://example.org/\"\ntitle = \"" + title + "\"\n",
		"content/_index.md":  "---\ntitle: " + title + "\n---\n\nThe body of " + title + ".\n",
		"layouts/index.html": "<!doctype html><html><body><h1>{{ .Title }}</h1>{{ .Content }}</body></html>\n",
	}
	for name, content := range files {
		where := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(where), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(where, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func servers(t *testing.T, settings preview.Settings) *preview.Servers {
	t.Helper()

	settings.HugoBin = hugoBin(t)
	if settings.Ready == 0 {
		settings.Ready = 60 * time.Second
	}
	made, err := preview.NewServers(settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(made.Close)
	return made
}

// asked is what a browser asking for the site's front page sees.
func asked(t *testing.T, handler http.Handler, path string) (int, string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	body, _ := io.ReadAll(recorder.Body)
	return recorder.Code, string(body)
}

func TestAPreviewShowsTheSite(t *testing.T) {
	made := servers(t, preview.Settings{})

	handler, err := made.For(t.Context(), "one", aSite(t, "Korenaar Bakery"), "http://example.org/preview/an/org/")
	if err != nil {
		t.Fatalf("For() error = %v", err)
	}

	status, body := asked(t, handler, "http://example.org/preview/an/org/")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", status, body)
	}
	if !strings.Contains(body, "Korenaar Bakery") {
		t.Errorf("the preview shows %q, want the site", body)
	}
}

func TestAChangeShowsInThePreviewWithoutRestartingIt(t *testing.T) {
	dir := aSite(t, "Korenaar Bakery")
	made := servers(t, preview.Settings{})

	handler, err := made.For(t.Context(), "one", dir, "http://example.org/preview/an/org/")
	if err != nil {
		t.Fatal(err)
	}
	if _, body := asked(t, handler, "http://example.org/preview/an/org/"); !strings.Contains(body, "The body of Korenaar Bakery") {
		t.Fatalf("the preview starts with %q", body)
	}

	if err := os.WriteFile(filepath.Join(dir, "content/_index.md"),
		[]byte("---\ntitle: Korenaar Bakery\n---\n\nNow with plums.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Hugo watches the files. The change is waited for by asking until it
	// shows, with a deadline, rather than by sleeping for a guess.
	deadline := time.Now().Add(20 * time.Second)
	for {
		_, body := asked(t, handler, "http://example.org/preview/an/org/")
		if strings.Contains(body, "Now with plums") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the change did not show within 20 seconds; the preview says %q", body)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if made.Running() != 1 {
		t.Errorf("%d previews are running, want the same one", made.Running())
	}
}

func TestAskingTwiceUsesTheSamePreview(t *testing.T) {
	dir := aSite(t, "Korenaar Bakery")
	made := servers(t, preview.Settings{})

	if _, err := made.For(t.Context(), "one", dir, "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}
	if _, err := made.For(t.Context(), "one", dir, "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}

	if made.Running() != 1 {
		t.Errorf("%d previews are running, want one", made.Running())
	}
}

func TestAPreviewNobodyHasLookedAtIsStopped(t *testing.T) {
	made := servers(t, preview.Settings{Idle: 15 * time.Minute})

	if _, err := made.For(t.Context(), "one", aSite(t, "Korenaar Bakery"), "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}
	if made.Running() != 1 {
		t.Fatalf("%d previews are running, want one", made.Running())
	}

	// A quarter of an hour later, as far as the idle check is concerned.
	made.PretendTimeIs(func() time.Time { return time.Now().Add(20 * time.Minute) })
	made.SweepNow()

	if made.Running() != 0 {
		t.Errorf("%d previews are still running, want none", made.Running())
	}
}

func TestAPreviewSomebodyIsLookingAtIsKept(t *testing.T) {
	dir := aSite(t, "Korenaar Bakery")
	made := servers(t, preview.Settings{Idle: 15 * time.Minute})

	handler, err := made.For(t.Context(), "one", dir, "http://example.org/preview/an/org/")
	if err != nil {
		t.Fatal(err)
	}

	made.PretendTimeIs(func() time.Time { return time.Now().Add(20 * time.Minute) })
	// Asking for it again is what looking at it does.
	if _, err := made.For(t.Context(), "one", dir, "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}
	made.SweepNow()

	if made.Running() != 1 {
		t.Errorf("%d previews are running, want the one somebody is looking at", made.Running())
	}
	if status, _ := asked(t, handler, "http://example.org/preview/an/org/"); status != http.StatusOK {
		t.Errorf("the kept preview answers %d", status)
	}
}

func TestOnlySoManyPreviewsRunAtOnce(t *testing.T) {
	made := servers(t, preview.Settings{Most: 2})

	for _, id := range []string{"one", "two"} {
		if _, err := made.For(t.Context(), id, aSite(t, id), "http://example.org/preview/an/org/"); err != nil {
			t.Fatalf("For(%q) error = %v", id, err)
		}
	}

	_, err := made.For(t.Context(), "three", aSite(t, "three"), "http://example.org/preview/an/org/")
	if !errors.Is(err, preview.ErrTooMany) {
		t.Fatalf("For() error = %v, want ErrTooMany", err)
	}
	if made.Running() != 2 {
		t.Errorf("%d previews are running, want the two allowed", made.Running())
	}
}

func TestAPreviewIsStoppedWhenItsSiteMoves(t *testing.T) {
	made := servers(t, preview.Settings{})

	if _, err := made.For(t.Context(), "one", aSite(t, "First"), "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}

	handler, err := made.For(t.Context(), "one", aSite(t, "Second"), "http://example.org/preview/an/org/")
	if err != nil {
		t.Fatalf("For() with a new working copy error = %v", err)
	}

	if made.Running() != 1 {
		t.Errorf("%d previews are running, want one", made.Running())
	}
	_, body := asked(t, handler, "http://example.org/preview/an/org/")
	if !strings.Contains(body, "Second") {
		t.Errorf("the preview shows %q, want the site it was pointed at", body)
	}
}

func TestStoppingEverythingLeavesNothingRunning(t *testing.T) {
	made := servers(t, preview.Settings{})

	if _, err := made.For(t.Context(), "one", aSite(t, "Korenaar Bakery"), "http://example.org/preview/an/org/"); err != nil {
		t.Fatal(err)
	}

	made.Close()

	if made.Running() != 0 {
		t.Errorf("%d previews survived", made.Running())
	}
	// Closing twice is what a shutdown that was already shutting down does.
	made.Close()
}

func TestAPreviewOfSomethingThatWillNotBuildSaysSo(t *testing.T) {
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, "hugo.toml"), []byte("this is not = [valid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	made := servers(t, preview.Settings{Ready: 3 * time.Second})

	_, err := made.For(t.Context(), "one", broken, "http://example.org/preview/an/org/")
	if !errors.Is(err, preview.ErrNotReady) {
		t.Fatalf("For() error = %v, want ErrNotReady", err)
	}
	if made.Running() != 0 {
		t.Errorf("%d previews are running after one that could not start", made.Running())
	}
}

func TestServersNeedAHugo(t *testing.T) {
	if _, err := preview.NewServers(preview.Settings{}); !errors.Is(err, preview.ErrNoHugo) {
		t.Errorf("NewServers() error = %v, want ErrNoHugo", err)
	}
}
