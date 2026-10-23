//go:build integration

package testsupport

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	clusterOnce sync.Once
	clusterURL  string
	clusterErr  error
	databaseNum struct {
		sync.Mutex
		n int
	}
)

// Postgres returns a connection string for a database that is empty, unique to
// the calling test, and dropped when the test finishes. It reuses a cluster
// named by CHATCMS_TEST_DATABASE_URL when there is one, and otherwise starts a
// throwaway cluster on a unix socket under TMPDIR.
func Postgres(t *testing.T) string {
	t.Helper()

	clusterOnce.Do(startCluster)
	if clusterErr != nil {
		t.Fatalf("no PostgreSQL available for integration tests: %v", clusterErr)
	}

	name := nextDatabaseName()
	ctx := t.Context()

	admin, err := pgx.Connect(ctx, clusterURL)
	if err != nil {
		t.Fatalf("connecting to the test cluster failed: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		admin.Close(ctx)
		t.Fatalf("creating the test database failed: %v", err)
	}
	admin.Close(ctx)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		conn, err := pgx.Connect(cleanupCtx, clusterURL)
		if err != nil {
			return
		}
		defer conn.Close(cleanupCtx)
		conn.Exec(cleanupCtx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})

	return replaceDatabase(clusterURL, name)
}

func nextDatabaseName() string {
	databaseNum.Lock()
	defer databaseNum.Unlock()
	databaseNum.n++
	return fmt.Sprintf("chatcms_test_%d_%d", os.Getpid(), databaseNum.n)
}

func replaceDatabase(url, name string) string {
	cut := strings.SplitN(url, "?", 2)
	base := cut[0]
	slash := strings.LastIndex(base, "/")
	base = base[:slash+1] + name
	if len(cut) == 2 {
		return base + "?" + cut[1]
	}
	return base
}

func startCluster() {
	if url := os.Getenv("CHATCMS_TEST_DATABASE_URL"); url != "" {
		clusterURL = url
		return
	}

	for _, tool := range []string{"initdb", "pg_ctl"} {
		if _, err := exec.LookPath(tool); err != nil {
			clusterErr = fmt.Errorf("%s is not on PATH and CHATCMS_TEST_DATABASE_URL is not set", tool)
			return
		}
	}

	root, err := os.MkdirTemp("", "chatcms-pg-")
	if err != nil {
		clusterErr = err
		return
	}

	data := filepath.Join(root, "data")
	socket := filepath.Join(root, "sock")
	if err := os.MkdirAll(socket, 0o700); err != nil {
		clusterErr = err
		return
	}

	if out, err := run("initdb", "-D", data, "-U", "postgres", "--auth=trust",
		"--encoding=UTF8", "--no-locale", "--nosync"); err != nil {
		clusterErr = fmt.Errorf("initdb failed: %v: %s", err, out)
		return
	}

	options := fmt.Sprintf("-k %s -h '' -c fsync=off -c full_page_writes=off "+
		"-c synchronous_commit=off -c max_connections=100", socket)
	if out, err := run("pg_ctl", "-D", data, "-o", options,
		"-l", filepath.Join(root, "postgres.log"), "-w", "start"); err != nil {
		log, _ := os.ReadFile(filepath.Join(root, "postgres.log"))
		clusterErr = fmt.Errorf("starting PostgreSQL failed: %v: %s: %s", err, out, log)
		return
	}

	stopWhenTheTestsAreOver(data, root)

	clusterURL = fmt.Sprintf("postgres://postgres@/postgres?host=%s", socket)
}

// stopWhenTheTestsAreOver ties the cluster's life to this test binary's. A
// cluster started with pg_ctl outlives whoever started it, so `go test
// -tags=integration ./...` left one PostgreSQL per package running until the
// machine was rebooted. The sitter holds a pipe to us: when we go, its read
// ends, and it stops the cluster and takes the directory with it.
func stopWhenTheTestsAreOver(data, root string) {
	sitter := exec.Command("sh", "-c",
		`cat >/dev/null; pg_ctl -D "$1" stop -m immediate >/dev/null 2>&1; rm -rf "$2"`,
		"sh", data, root)

	held, err := sitter.StdinPipe()
	if err != nil {
		return
	}
	if err := sitter.Start(); err != nil {
		return
	}

	// Kept so that the pipe is not closed by the garbage collector while the
	// tests are still running, which would stop the cluster under them.
	keptOpen = append(keptOpen, held)
}

var keptOpen []io.Closer

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "PGUSER=postgres", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	return string(out), err
}
