//go:build integration

package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms/internal/app"
	"github.com/quiqr/chatcms/internal/httpapi"
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
