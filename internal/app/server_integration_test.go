//go:build integration

package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms/internal/app"
	"github.com/quiqr/chatcms/internal/auth"
	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/jobs"
	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/testsupport"
)

func testConfig(t *testing.T) app.Config {
	t.Helper()

	return app.Config{
		Listen:          "127.0.0.1:0",
		BaseURL:         "http://localhost",
		StateDir:        t.TempDir(),
		DatabaseURL:     testsupport.Postgres(t),
		ShutdownTimeout: 5 * time.Second,
		SyncInterval:    time.Hour,
		LogFormat:       "json",
		GitBin:          "git",
		HugoBin:         "hugo",
	}
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func poison(t *testing.T, url string) {
	t.Helper()

	conn, err := pgx.Connect(t.Context(), url)
	if err != nil {
		t.Fatalf("connecting to the test database failed: %v", err)
	}
	defer conn.Close(t.Context())

	if _, err := conn.Exec(t.Context(), "CREATE TABLE jobs (blocking_the_migration int)"); err != nil {
		t.Fatalf("creating the conflicting table failed: %v", err)
	}
}

func startServer(t *testing.T, cfg app.Config) (*app.Server, func()) {
	t.Helper()

	server, err := app.NewServer(t.Context(), cfg, quietLogger())
	if err != nil {
		t.Fatalf("NewServer() error = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx) }()

	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Error("the server did not stop within ten seconds")
			}
		})
	}
	t.Cleanup(stop)

	waitForListener(t, server.Addr())
	return server, stop
}

func waitForListener(t *testing.T, address string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
	}
	t.Fatalf("nothing is listening on %s", address)
}

func TestHealthOKAgainstARealDatabase(t *testing.T) {
	server, _ := startServer(t, testConfig(t))

	response, err := http.Get("http://" + server.Addr() + "/api/health")
	if err != nil {
		t.Fatalf("requesting health failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want 200: %s", response.StatusCode, body)
	}

	var health httpapi.Health
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		t.Fatalf("health body is not a health report: %v", err)
	}
	if health.Database != httpapi.DatabaseReachable {
		t.Errorf("database = %q, want %q", health.Database, httpapi.DatabaseReachable)
	}
	if health.SchemaVersion == "" || health.SchemaVersion == "0" {
		t.Errorf("schemaVersion = %q, want the applied migration version", health.SchemaVersion)
	}
}

func TestServerServesTheFrontendNotice(t *testing.T) {
	server, _ := startServer(t, testConfig(t))

	response, err := http.Get("http://" + server.Addr() + "/projects")
	if err != nil {
		t.Fatalf("requesting the application failed: %v", err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if !strings.Contains(string(body), "<") {
		t.Errorf("body = %q, want an HTML document", body)
	}
}

func TestRunDoesNotListenWhenMigrationFails(t *testing.T) {
	cfg := testConfig(t)
	cfg.Listen = freeAddress(t)

	poison(t, cfg.DatabaseURL)

	_, err := app.NewServer(t.Context(), cfg, quietLogger())
	if err == nil {
		t.Fatal("NewServer() error = nil, want a migration failure")
	}

	conn, dialErr := net.DialTimeout("tcp", cfg.Listen, 200*time.Millisecond)
	if dialErr == nil {
		conn.Close()
		t.Errorf("something is listening on %s after a failed migration", cfg.Listen)
	}
}

func TestNewServerRefusesAnUnreachableDatabase(t *testing.T) {
	cfg := testConfig(t)
	cfg.DatabaseURL = "postgres://nobody@127.0.0.1:1/absent?connect_timeout=1"

	_, err := app.NewServer(t.Context(), cfg, quietLogger())
	if err == nil {
		t.Fatal("NewServer() error = nil, want an unreachable database error")
	}
	if !strings.Contains(err.Error(), "not reachable") {
		t.Errorf("error = %v, want it to say the database is not reachable", err)
	}
}

func TestShutdownDrainsInFlightRequest(t *testing.T) {
	cfg := testConfig(t)
	server, stop := startServer(t, cfg)

	release := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(1)

	var status int
	var requestErr error
	go func() {
		defer wait.Done()
		<-release
		response, err := http.Get("http://" + server.Addr() + "/api/health")
		if err != nil {
			requestErr = err
			return
		}
		defer response.Body.Close()
		io.Copy(io.Discard, response.Body)
		status = response.StatusCode
	}()

	close(release)
	wait.Wait()
	stop()

	if requestErr != nil {
		t.Fatalf("the in-flight request failed: %v", requestErr)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}

	if _, err := http.Get("http://" + server.Addr() + "/api/health"); err == nil {
		t.Error("the server still accepts connections after shutdown")
	}
}

func freeAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free port failed: %v", err)
	}
	address := listener.Addr().String()
	listener.Close()
	return address
}

func TestSyncJobRunsASync(t *testing.T) {
	cfg := testConfig(t)
	logger := quietLogger()

	gitCfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(gitCfg, cfg.SitesDir())

	remote := bareRemoteWithAPage(t, gitCfg)
	repository, err := sites.Clone(t.Context(), "korenaar", remote)
	if err != nil {
		t.Fatalf("Clone() error = %v, want nil", err)
	}
	repository.Close()

	pushAnotherPage(t, gitCfg, remote)

	db, err := store.Open(t.Context(), cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	queue := jobs.NewQueue(db)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{Logger: logger})
	app.RegisterSyncJobForTest(worker, sites, logger)

	if _, err := queue.Enqueue(t.Context(), gitsync.SyncJobKind, "korenaar", nil); err != nil {
		t.Fatal(err)
	}
	if did, err := worker.RunOnce(t.Context()); err != nil || !did {
		t.Fatalf("RunOnce() = %v, %v, want the sync to have run", did, err)
	}

	done, _ := queue.CountByState(t.Context(), "done")
	if done != 1 {
		failed, _ := queue.CountByState(t.Context(), "queued")
		t.Fatalf("done = %d, queued = %d, want the sync to have succeeded", done, failed)
	}

	reopened, err := sites.Open(t.Context(), "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	if _, err := fs.ReadFile(reopened.FS(), "content/news.md"); err != nil {
		t.Errorf("the page pushed elsewhere did not arrive: %v", err)
	}
}

func TestShutdownDrainsTheWorkerToo(t *testing.T) {
	cfg := testConfig(t)
	cfg.SyncInterval = 50 * time.Millisecond

	server, stop := startServer(t, cfg)

	db, err := store.Open(t.Context(), cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	queue := jobs.NewQueue(db)
	if _, err := queue.Enqueue(t.Context(), "a-kind-with-no-handler", "korenaar", nil); err != nil {
		t.Fatal(err)
	}

	// The job has no handler, so the worker keeps it rather than losing it.
	// What matters here is that shutting the server down does not strand it.
	stop()

	queued, err := queue.CountByState(context.Background(), "queued")
	if err != nil {
		t.Fatal(err)
	}
	if queued != 1 {
		t.Errorf("queued = %d after shutdown, want the job still there", queued)
	}

	_ = server
}

func gitBinary(t *testing.T) string {
	t.Helper()

	if bin := os.Getenv("CHATCMS_GIT_BIN"); bin != "" {
		return bin
	}
	found, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}
	return found
}

func bareRemoteWithAPage(t *testing.T, cfg gitsync.Config) string {
	t.Helper()

	remote := filepath.Join(t.TempDir(), "remote.git")
	gitRun(t, cfg, "", "init", "--bare", "--initial-branch=main", remote)

	seed := filepath.Join(t.TempDir(), "seed")
	gitRun(t, cfg, "", "clone", "--", remote, seed)
	writeSeedFile(t, filepath.Join(seed, "content", "_index.md"), "---\ntitle: Home\n---\n")
	gitRun(t, cfg, seed, "add", "--all")
	gitRun(t, cfg, seed, "commit", "--message=First version")
	gitRun(t, cfg, seed, "push", "origin", "HEAD:refs/heads/main")

	return remote
}

func pushAnotherPage(t *testing.T, cfg gitsync.Config, remote string) {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "elsewhere")
	gitRun(t, cfg, "", "clone", "--", remote, dir)
	writeSeedFile(t, filepath.Join(dir, "content", "news.md"), "---\ntitle: News\n---\n")
	gitRun(t, cfg, dir, "add", "--all")
	gitRun(t, cfg, dir, "commit", "--message=Add news")
	gitRun(t, cfg, dir, "push", "origin", "HEAD:refs/heads/main")
}

func gitRun(t *testing.T, cfg gitsync.Config, dir string, args ...string) {
	t.Helper()

	if err := gitsync.RunForTest(t.Context(), cfg, dir, t.TempDir(), args...); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

func writeSeedFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// client keeps cookies the way a browser does, and repeats the request token
// the way the application's own pages do.
type client struct {
	t       *testing.T
	base    string
	jar     *cookiejar.Jar
	http    *http.Client
	baseURL *url.URL
}

func newClient(t *testing.T, address string) *client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse("http://" + address)
	if err != nil {
		t.Fatal(err)
	}

	return &client{t: t, base: "http://" + address, jar: jar, http: &http.Client{Jar: jar}, baseURL: parsed}
}

func (c *client) requestToken() string {
	for _, cookie := range c.jar.Cookies(c.baseURL) {
		if cookie.Name == httpapi.RequestTokenCookie {
			return cookie.Value
		}
	}
	return ""
}

func (c *client) sessionValue() string {
	for _, cookie := range c.jar.Cookies(c.baseURL) {
		if cookie.Name == httpapi.SessionCookie {
			return cookie.Value
		}
	}
	return ""
}

func (c *client) do(method, path string, body any, withToken bool) *http.Response {
	c.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	if withToken {
		request.Header.Set(httpapi.RequestTokenHeader, c.requestToken())
	}

	response, err := c.http.Do(request)
	if err != nil {
		c.t.Fatalf("%s %s failed: %v", method, path, err)
	}
	return response
}

// warmUp fetches something so the browser is given a request token.
func (c *client) warmUp() {
	response := c.do(http.MethodGet, "/api/health", nil, false)
	response.Body.Close()
}

func signInHarness(t *testing.T) (*app.Server, *client, string) {
	t.Helper()

	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	const address = "joris@example.org"
	const password = "a long enough passphrase"

	if _, err := server.Auth().People().Create(t.Context(), address, "Joris Bakker", password); err != nil {
		t.Fatalf("creating a person failed: %v", err)
	}

	caller := newClient(t, server.Addr())
	caller.warmUp()
	return server, caller, password
}

func TestSignInThroughTheAPI(t *testing.T) {
	_, caller, password := signInHarness(t)

	response := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, true)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want 200: %s", response.StatusCode, body)
	}
	if caller.sessionValue() == "" {
		t.Fatal("no session cookie was set")
	}

	me := caller.do(http.MethodGet, "/api/me", nil, false)
	defer me.Body.Close()

	if me.StatusCode != http.StatusOK {
		t.Fatalf("/api/me status = %d, want 200", me.StatusCode)
	}

	var person struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(me.Body).Decode(&person); err != nil {
		t.Fatal(err)
	}
	if person.Email != "joris@example.org" {
		t.Errorf("email = %q", person.Email)
	}
}

func TestProtectedEndpointRefusesNobody(t *testing.T) {
	_, caller, _ := signInHarness(t)

	response := caller.do(http.MethodGet, "/api/me", nil, false)
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.StatusCode)
	}

	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != httpapi.CodeUnauthorized {
		t.Errorf("code = %q, want %q", envelope.Error.Code, httpapi.CodeUnauthorized)
	}
}

func TestSignInWithoutTheRequestTokenIsRefused(t *testing.T) {
	_, caller, password := signInHarness(t)

	response := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, false)
	defer response.Body.Close()

	if response.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", response.StatusCode)
	}
	if caller.sessionValue() != "" {
		t.Error("a session was issued although the request was refused")
	}
}

func TestSignInRevokesTheSessionItArrivedWith(t *testing.T) {
	server, caller, password := signInHarness(t)

	first := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, true)
	first.Body.Close()
	original := caller.sessionValue()

	second := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, true)
	second.Body.Close()
	replacement := caller.sessionValue()

	if original == "" || replacement == "" {
		t.Fatal("a sign-in did not set a session")
	}
	if original == replacement {
		t.Fatal("signing in again reused the session it arrived with")
	}

	actor, err := server.Auth().Lookup(t.Context(), original)
	if err != nil {
		t.Fatal(err)
	}
	if !actor.Anonymous {
		t.Error("the session the sign-in arrived with still works")
	}
}

func TestSignInRefusesWrongCredentials(t *testing.T) {
	_, caller, _ := signInHarness(t)

	for _, body := range []map[string]any{
		{"email": "joris@example.org", "password": "not the password at all"},
		{"email": "nobody@example.org", "password": "not the password at all"},
	} {
		response := caller.do(http.MethodPost, "/api/sign-in", body, true)
		payload, _ := io.ReadAll(response.Body)
		response.Body.Close()

		if response.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d for %v, want 401", response.StatusCode, body["email"])
		}
		if strings.Contains(string(payload), "exist") || strings.Contains(string(payload), "unknown") {
			t.Errorf("the answer hints at whether the address is known: %s", payload)
		}
	}
}

func TestSignOutRevokes(t *testing.T) {
	server, caller, password := signInHarness(t)

	in := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, true)
	in.Body.Close()
	session := caller.sessionValue()

	out := caller.do(http.MethodPost, "/api/sign-out", nil, true)
	out.Body.Close()

	if out.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", out.StatusCode)
	}

	actor, err := server.Auth().Lookup(t.Context(), session)
	if err != nil {
		t.Fatal(err)
	}
	if !actor.Anonymous {
		t.Error("the session still works after signing out")
	}

	me := caller.do(http.MethodGet, "/api/me", nil, false)
	me.Body.Close()
	if me.StatusCode != http.StatusUnauthorized {
		t.Errorf("/api/me after signing out = %d, want 401", me.StatusCode)
	}
}

func TestSignOutWithoutASession(t *testing.T) {
	_, caller, _ := signInHarness(t)

	response := caller.do(http.MethodPost, "/api/sign-out", nil, true)
	response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204 even with nothing to sign out of", response.StatusCode)
	}
}

func TestResetRequestSaysTheSameThingEitherWay(t *testing.T) {
	_, caller, _ := signInHarness(t)

	var answers []string
	for _, address := range []string{"joris@example.org", "nobody@example.org"} {
		response := caller.do(http.MethodPost, "/api/password-reset",
			map[string]any{"email": address}, true)
		payload, _ := io.ReadAll(response.Body)
		response.Body.Close()

		if response.StatusCode != http.StatusOK {
			t.Fatalf("status = %d for %s, want 200", response.StatusCode, address)
		}
		answers = append(answers, string(payload))
	}

	if answers[0] != answers[1] {
		t.Errorf("the answers differ:\n%s\n%s", answers[0], answers[1])
	}
	for _, answer := range answers {
		if strings.Contains(answer, "token") {
			t.Errorf("the answer carries a link: %s", answer)
		}
	}
}

func TestSetPasswordThroughTheAPI(t *testing.T) {
	server, caller, _ := signInHarness(t)

	person, err := server.Auth().People().Create(t.Context(), "invited@example.org", "Invited", "")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := server.Auth().Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}

	response := caller.do(http.MethodPost, "/api/set-password",
		map[string]any{"token": token, "password": "a brand new long passphrase"}, true)
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}

	in := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "invited@example.org", "password": "a brand new long passphrase"}, true)
	in.Body.Close()
	if in.StatusCode != http.StatusOK {
		t.Errorf("signing in after setting a password = %d, want 200", in.StatusCode)
	}
}

func TestSetPasswordRefusesAWeakOne(t *testing.T) {
	server, caller, _ := signInHarness(t)

	person, err := server.Auth().People().Create(t.Context(), "invited2@example.org", "Invited", "")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := server.Auth().Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}

	response := caller.do(http.MethodPost, "/api/set-password",
		map[string]any{"token": token, "password": "short"}, true)
	payload, _ := io.ReadAll(response.Body)
	response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
	if !strings.Contains(string(payload), "12") {
		t.Errorf("the message does not say what the minimum is: %s", payload)
	}
}

func TestSetPasswordRefusesAnUnusableLink(t *testing.T) {
	_, caller, _ := signInHarness(t)

	response := caller.do(http.MethodPost, "/api/set-password",
		map[string]any{"token": "nothing-like-a-token", "password": "a long enough passphrase"}, true)
	payload, _ := io.ReadAll(response.Body)
	response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
	if !strings.Contains(strings.ToLower(string(payload)), "new one") {
		t.Errorf("the message does not say what to do next: %s", payload)
	}
}
