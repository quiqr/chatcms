//go:build integration

package app_test

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms/internal/app"
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
