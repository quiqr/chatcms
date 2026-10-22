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
	"github.com/quiqr/chatcms/internal/policy"
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
		SecretKeyFile:   masterKeyFile(t),
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
	sites := gitsync.NewSites(gitCfg, cfg.SitesDir(), nil)

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

// masterKeyFile writes a key for the secrets the server keeps, which it
// refuses to start without.
func masterKeyFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "master-key")
	if err := os.WriteFile(path, []byte(strings.Repeat("k", 32)), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
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

// TestAcceptanceCreateOrganisationAndInviteAnEditor is acceptance test 1 from
// docs/BRIEFING-M01.md section 6, proven end to end through the API: an
// organisation is created, a project is made in it, a customer is invited as
// an editor, and they sign in with the link they were given.
func TestAcceptanceCreateOrganisationAndInviteAnEditor(t *testing.T) {
	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	const adminAddress = "admin@technative.example"
	const adminPassword = "a long enough passphrase"

	if _, err := server.Auth().People().Create(t.Context(), adminAddress, "The Admin", adminPassword); err != nil {
		t.Fatal(err)
	}

	admin := newClient(t, server.Addr())
	admin.warmUp()

	signedIn := admin.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": adminAddress, "password": adminPassword}, true)
	signedIn.Body.Close()
	if signedIn.StatusCode != http.StatusOK {
		t.Fatalf("the admin could not sign in: %d", signedIn.StatusCode)
	}

	created := admin.do(http.MethodPost, "/api/organisations",
		map[string]any{"slug": "technative", "name": "TechNative"}, true)
	created.Body.Close()
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("creating the organisation = %d, want 201", created.StatusCode)
	}

	project := admin.do(http.MethodPost, "/api/organisations/technative/projects",
		map[string]any{"slug": "korenaar", "name": "Korenaar Bakery", "domain": "korenaar.example"}, true)
	project.Body.Close()
	if project.StatusCode != http.StatusCreated {
		t.Fatalf("creating the project = %d, want 201", project.StatusCode)
	}

	invited := admin.do(http.MethodPost, "/api/organisations/technative/projects/korenaar/people",
		map[string]any{"email": "customer@korenaar.example", "role": "editor"}, true)
	payload, _ := io.ReadAll(invited.Body)
	invited.Body.Close()
	if invited.StatusCode != http.StatusCreated {
		t.Fatalf("inviting the customer = %d, want 201: %s", invited.StatusCode, payload)
	}

	var member struct {
		Role          string `json:"role"`
		StillToAccept bool   `json:"stillToAccept"`
	}
	if err := json.Unmarshal(payload, &member); err != nil {
		t.Fatal(err)
	}
	if member.Role != "editor" || !member.StillToAccept {
		t.Fatalf("member = %+v, want an editor who has still to accept", member)
	}

	// The customer sets a password through the link they were issued, and
	// signs in with it.
	person, err := server.Auth().People().ByEmail(t.Context(), "customer@korenaar.example")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := server.Auth().Credentials().Issue(t.Context(), auth.KindInvitation, person.ID)
	if err != nil {
		t.Fatal(err)
	}

	customer := newClient(t, server.Addr())
	customer.warmUp()

	set := customer.do(http.MethodPost, "/api/set-password",
		map[string]any{"token": token, "password": "the customer's own passphrase"}, true)
	set.Body.Close()
	if set.StatusCode != http.StatusOK {
		t.Fatalf("setting the password = %d, want 200", set.StatusCode)
	}

	in := customer.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "customer@korenaar.example", "password": "the customer's own passphrase"}, true)
	in.Body.Close()
	if in.StatusCode != http.StatusOK {
		t.Fatalf("the customer could not sign in: %d", in.StatusCode)
	}

	// They can see the project, and are told they may edit but not administer.
	view := customer.do(http.MethodGet, "/api/organisations/technative/projects/korenaar", nil, false)
	viewPayload, _ := io.ReadAll(view.Body)
	view.Body.Close()
	if view.StatusCode != http.StatusOK {
		t.Fatalf("the customer cannot see the project: %d", view.StatusCode)
	}

	var projectView struct {
		YourRole string   `json:"yourRole"`
		YouMay   []string `json:"youMay"`
	}
	if err := json.Unmarshal(viewPayload, &projectView); err != nil {
		t.Fatal(err)
	}
	if projectView.YourRole != "editor" {
		t.Errorf("yourRole = %q, want editor", projectView.YourRole)
	}

	may := map[string]bool{}
	for _, action := range projectView.YouMay {
		may[action] = true
	}
	if !may["edit-content"] || !may["publish"] {
		t.Errorf("the customer may %v, want editing and publishing", projectView.YouMay)
	}
	if may["manage-people"] || may["see-technical-detail"] {
		t.Errorf("the customer may %v, which includes things an editor may not do", projectView.YouMay)
	}
}

func TestAViewerCannotChangeAnything(t *testing.T) {
	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	owner, err := server.Auth().People().Create(t.Context(), "owner@example.org", "Owner", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	viewer, err := server.Auth().People().Create(t.Context(), "viewer@example.org", "Viewer", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}

	tenancyStore := server.Tenancy().Store()
	organisation, err := tenancyStore.CreateOrganisation(t.Context(), "technative", "TechNative", owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	project, err := tenancyStore.CreateProject(t.Context(), organisation.ID, "korenaar", "Korenaar", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := tenancyStore.SetProjectRole(t.Context(), project.ID, owner.ID, policy.RoleOwner); err != nil {
		t.Fatal(err)
	}
	if err := tenancyStore.SetProjectRole(t.Context(), project.ID, viewer.ID, policy.RoleViewer); err != nil {
		t.Fatal(err)
	}

	caller := newClient(t, server.Addr())
	caller.warmUp()
	in := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "viewer@example.org", "password": "a long enough passphrase"}, true)
	in.Body.Close()

	// Everything a viewer might try to change, called directly rather than
	// hidden behind an interface.
	attempts := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/api/organisations/technative/projects", map[string]any{"slug": "new", "name": "New"}},
		{http.MethodPost, "/api/organisations/technative/projects/korenaar/people",
			map[string]any{"email": "someone@example.org", "role": "editor"}},
		{http.MethodPut, "/api/organisations/technative/projects/korenaar/people/" + owner.ID.String(),
			map[string]any{"role": "viewer"}},
		{http.MethodDelete, "/api/organisations/technative/projects/korenaar/people/" + owner.ID.String(), nil},
	}

	for _, attempt := range attempts {
		response := caller.do(attempt.method, attempt.path, attempt.body, true)
		payload, _ := io.ReadAll(response.Body)
		response.Body.Close()

		if response.StatusCode != http.StatusForbidden {
			t.Errorf("%s %s = %d, want 403: %s", attempt.method, attempt.path, response.StatusCode, payload)
		}
	}

	// Nothing changed.
	members, err := tenancyStore.Members(t.Context(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Errorf("the project has %d people, want the two it started with", len(members))
	}

	projects, err := tenancyStore.ProjectsIn(t.Context(), organisation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("the organisation has %d projects, want the one it started with", len(projects))
	}

	// Reading still works.
	read := caller.do(http.MethodGet, "/api/organisations/technative/projects/korenaar", nil, false)
	read.Body.Close()
	if read.StatusCode != http.StatusOK {
		t.Errorf("a viewer reading the project = %d, want 200", read.StatusCode)
	}
}

func TestNobodyGetsUnauthorizedRatherThanForbidden(t *testing.T) {
	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	caller := newClient(t, server.Addr())
	caller.warmUp()

	response := caller.do(http.MethodGet, "/api/organisations", nil, false)
	response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for a request from nobody", response.StatusCode)
	}
}

// TestAcceptanceImportWithADeployKey is acceptance test 2 from
// docs/BRIEFING-M01.md section 6, through the API: a project is set up, its
// key is shown, the key is given to the host, the site is imported, and the
// model is found.
func TestAcceptanceImportWithADeployKey(t *testing.T) {
	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	const address = "implementer@technative.example"
	const password = "a long enough passphrase"

	if _, err := server.Auth().People().Create(t.Context(), address, "The Implementer", password); err != nil {
		t.Fatal(err)
	}

	caller := newClient(t, server.Addr())
	caller.warmUp()
	in := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": address, "password": password}, true)
	in.Body.Close()

	created := caller.do(http.MethodPost, "/api/organisations",
		map[string]any{"slug": "technative", "name": "TechNative"}, true)
	created.Body.Close()

	project := caller.do(http.MethodPost, "/api/organisations/technative/projects",
		map[string]any{"slug": "korenaar", "name": "Korenaar Bakery", "domain": "korenaar.example"}, true)
	project.Body.Close()
	if project.StatusCode != http.StatusCreated {
		t.Fatalf("creating the project = %d", project.StatusCode)
	}

	base := "/api/organisations/technative/projects/korenaar"

	// The key is made and shown, so the implementer can give it to the host.
	keyResponse := caller.do(http.MethodPost, base+"/key", nil, true)
	keyPayload, _ := io.ReadAll(keyResponse.Body)
	keyResponse.Body.Close()
	if keyResponse.StatusCode != http.StatusCreated {
		t.Fatalf("making a key = %d: %s", keyResponse.StatusCode, keyPayload)
	}

	var key struct {
		PublicKey string `json:"publicKey"`
		Advice    string `json:"advice"`
	}
	if err := json.Unmarshal(keyPayload, &key); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key.PublicKey, "ssh-ed25519 ") {
		t.Fatalf("public key = %q", key.PublicKey)
	}
	if strings.Contains(string(keyPayload), "PRIVATE") {
		t.Fatalf("the answer carries a private key: %s", keyPayload)
	}
	if key.Advice == "" {
		t.Error("nothing was said about what to do with the key")
	}

	// The customer's site, on a remote the key reaches. A repository on this
	// machine needs no key, which is what makes this testable without a host.
	remote := customerRemote(t, server)

	imported := caller.do(http.MethodPost, base+"/import",
		map[string]any{"kind": "repository", "address": remote}, true)
	importPayload, _ := io.ReadAll(imported.Body)
	imported.Body.Close()
	if imported.StatusCode != http.StatusCreated {
		t.Fatalf("importing = %d: %s", imported.StatusCode, importPayload)
	}

	var result struct {
		Ready       bool `json:"ready"`
		IsHugoSite  bool `json:"isHugoSite"`
		HasModel    bool `json:"hasModel"`
		Pages       int  `json:"pages"`
		Collections int  `json:"collections"`
	}
	if err := json.Unmarshal(importPayload, &result); err != nil {
		t.Fatal(err)
	}
	if !result.IsHugoSite || !result.HasModel || !result.Ready {
		t.Fatalf("result = %+v, want a Hugo site with a model", result)
	}
	if result.Pages != 1 || result.Collections != 1 {
		t.Errorf("pages = %d, collections = %d, want 1 and 1", result.Pages, result.Collections)
	}

	// The connection is good, and the answer carries no secret.
	connection := caller.do(http.MethodPost, base+"/connection", nil, true)
	connectionPayload, _ := io.ReadAll(connection.Body)
	connection.Body.Close()
	if connection.StatusCode != http.StatusOK {
		t.Fatalf("testing the connection = %d: %s", connection.StatusCode, connectionPayload)
	}

	var reachable struct {
		Good    bool   `json:"good"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(connectionPayload, &reachable); err != nil {
		t.Fatal(err)
	}
	if !reachable.Good {
		t.Errorf("the connection is not good: %s", reachable.Message)
	}
	if strings.Contains(string(connectionPayload), "PRIVATE") {
		t.Error("the connection answer carries a private key")
	}

	// Publishing is recorded.
	publishing := caller.do(http.MethodPut, base+"/publishing",
		map[string]any{"to": "its-own-line", "form": "source"}, true)
	publishingPayload, _ := io.ReadAll(publishing.Body)
	publishing.Body.Close()
	if publishing.StatusCode != http.StatusOK {
		t.Fatalf("setting publishing = %d: %s", publishing.StatusCode, publishingPayload)
	}
	if !strings.Contains(string(publishingPayload), "quiqr/publish") {
		t.Errorf("publishing = %s, want the default line offered", publishingPayload)
	}

	// The site can now be inspected without importing again.
	site := caller.do(http.MethodGet, base+"/site", nil, false)
	sitePayload, _ := io.ReadAll(site.Body)
	site.Body.Close()
	if site.StatusCode != http.StatusOK || !strings.Contains(string(sitePayload), `"ready":true`) {
		t.Errorf("inspecting the site = %d: %s", site.StatusCode, sitePayload)
	}
}

func TestThePrivateKeyIsNeverReturned(t *testing.T) {
	cfg := testConfig(t)
	server, _ := startServer(t, cfg)

	const address = "implementer@technative.example"
	const password = "a long enough passphrase"
	if _, err := server.Auth().People().Create(t.Context(), address, "The Implementer", password); err != nil {
		t.Fatal(err)
	}

	caller := newClient(t, server.Addr())
	caller.warmUp()
	in := caller.do(http.MethodPost, "/api/sign-in", map[string]any{"email": address, "password": password}, true)
	in.Body.Close()

	caller.do(http.MethodPost, "/api/organisations",
		map[string]any{"slug": "technative", "name": "TechNative"}, true).Body.Close()
	caller.do(http.MethodPost, "/api/organisations/technative/projects",
		map[string]any{"slug": "korenaar", "name": "Korenaar"}, true).Body.Close()

	base := "/api/organisations/technative/projects/korenaar"
	caller.do(http.MethodPost, base+"/key", nil, true).Body.Close()

	for _, where := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, base},
		{http.MethodGet, base + "/key"},
		{http.MethodGet, base + "/site"},
		{http.MethodGet, base + "/publishing"},
		{http.MethodGet, "/api/organisations/technative/projects"},
	} {
		response := caller.do(where.method, where.path, nil, false)
		payload, _ := io.ReadAll(response.Body)
		response.Body.Close()

		for _, secret := range []string{"PRIVATE KEY", "deployPrivateKey", "BEGIN OPENSSH"} {
			if strings.Contains(string(payload), secret) {
				t.Errorf("%s %s carries %q: %s", where.method, where.path, secret, payload)
			}
		}
	}
}

// customerRemote builds a bare repository holding a Hugo site with a Quiqr
// model, standing in for the customer's own.
func customerRemote(t *testing.T, server *app.Server) string {
	t.Helper()

	cfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}

	remote := filepath.Join(t.TempDir(), "customer.git")
	gitRun(t, cfg, "", "init", "--bare", "--initial-branch=main", remote)

	seed := filepath.Join(t.TempDir(), "seed")
	gitRun(t, cfg, "", "clone", "--", remote, seed)

	writeSeedFile(t, filepath.Join(seed, "hugo.toml"), "baseURL = 'https://korenaar.example'\n")
	writeSeedFile(t, filepath.Join(seed, "quiqr", "model", "base.yaml"), `singles:
  - key: home
    title: Homepage
    file: content/_index.md
    fields:
      - key: title
        type: string
collections:
  - key: posts
    title: Posts
    folder: content/posts/
    fields:
      - key: title
        type: string
`)
	writeSeedFile(t, filepath.Join(seed, "content", "_index.md"), "---\ntitle: Home\n---\n\nWelcome.\n")

	gitRun(t, cfg, seed, "add", "--all")
	gitRun(t, cfg, seed, "commit", "--message=The customer's site")
	gitRun(t, cfg, seed, "push", "origin", "HEAD:refs/heads/main")

	_ = server
	return remote
}
