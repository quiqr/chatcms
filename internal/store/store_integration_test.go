//go:build integration

package store_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/quiqr/chatcms/internal/store"
	"github.com/quiqr/chatcms/internal/testsupport"
)

func open(t *testing.T) *store.DB {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	t.Cleanup(db.Close)
	return db
}

func TestOpenPingsDatabase(t *testing.T) {
	db := open(t)

	if err := db.Ping(t.Context()); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestOpenRejectsAnUnparseableURL(t *testing.T) {
	_, err := store.Open(t.Context(), "this is not a connection string")
	if err == nil {
		t.Fatal("Open() error = nil, want an error")
	}
	if got := err.Error(); got == "" {
		t.Error("the error carries no message")
	}
}

func TestMigrateFromEmptyDatabase(t *testing.T) {
	db := open(t)
	ctx := t.Context()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}

	applied, err := db.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion() error = %v, want nil", err)
	}
	known, err := db.KnownSchemaVersion()
	if err != nil {
		t.Fatalf("KnownSchemaVersion() error = %v, want nil", err)
	}
	if applied != known {
		t.Errorf("applied version = %d, want the latest known version %d", applied, known)
	}
	if known == 0 {
		t.Error("the binary knows no migrations")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := open(t)
	ctx := t.Context()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate() error = %v, want nil", err)
	}
	first, _ := db.SchemaVersion(ctx)

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate() error = %v, want nil", err)
	}
	second, _ := db.SchemaVersion(ctx)

	if first != second {
		t.Errorf("version moved from %d to %d on a second run", first, second)
	}
}

func TestMigrateUpDown(t *testing.T) {
	db := open(t)
	ctx := t.Context()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}
	if _, err := db.Queries.CountJobsByState(ctx, "queued"); err != nil {
		t.Fatalf("the jobs table is not usable after migrating: %v", err)
	}

	if err := db.MigrateDownTo(ctx, 0); err != nil {
		t.Fatalf("MigrateDownTo() error = %v, want nil", err)
	}
	if _, err := db.Queries.CountJobsByState(ctx, "queued"); err == nil {
		t.Error("the jobs table still exists after rolling the migration back")
	}
}

func TestConcurrentMigrateIsSafe(t *testing.T) {
	url := testsupport.Postgres(t)
	ctx := t.Context()

	const instances = 4
	errs := make([]error, instances)
	var wait sync.WaitGroup

	for i := range instances {
		wait.Add(1)
		go func() {
			defer wait.Done()

			db, err := store.Open(ctx, url)
			if err != nil {
				errs[i] = err
				return
			}
			defer db.Close()
			errs[i] = db.Migrate(ctx)
		}()
	}
	wait.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("instance %d: Migrate() error = %v, want nil", i, err)
		}
	}

	db, err := store.Open(ctx, url)
	if err != nil {
		t.Fatalf("store.Open() error = %v, want nil", err)
	}
	defer db.Close()

	rows, err := db.Pool().Query(ctx, "SELECT version_id, count(*) FROM goose_db_version GROUP BY version_id HAVING count(*) > 1")
	if err != nil {
		t.Fatalf("reading the migration log failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version, count int64
		rows.Scan(&version, &count)
		t.Errorf("migration %d was applied %d times, want once", version, count)
	}
}

func TestMigrateRefusesNewerDatabase(t *testing.T) {
	db := open(t)
	ctx := t.Context()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}

	known, _ := db.KnownSchemaVersion()
	_, err := db.Pool().Exec(ctx,
		"INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true)", known+1)
	if err != nil {
		t.Fatalf("faking a newer schema failed: %v", err)
	}

	err = db.Migrate(ctx)
	if !errors.Is(err, store.ErrDatabaseAhead) {
		t.Errorf("Migrate() error = %v, want ErrDatabaseAhead", err)
	}
}

func TestJobsRoundTrip(t *testing.T) {
	db := open(t)
	ctx := t.Context()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v, want nil", err)
	}

	queued, err := db.Queries.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:    "sync",
		Subject: "korenaar",
		Payload: []byte(`{"project":"korenaar"}`),
	})
	if err != nil {
		t.Fatalf("EnqueueJob() error = %v, want nil", err)
	}
	if queued.State != "queued" {
		t.Errorf("state = %q, want queued", queued.State)
	}

	claimed, err := db.Queries.ClaimNextJob(ctx)
	if err != nil {
		t.Fatalf("ClaimNextJob() error = %v, want nil", err)
	}
	if claimed.ID != queued.ID {
		t.Errorf("claimed job %d, want %d", claimed.ID, queued.ID)
	}
	if claimed.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", claimed.Attempts)
	}

	if _, err := db.Queries.ClaimNextJob(ctx); err == nil {
		t.Error("a second claim returned a job, want none left to claim")
	}

	if err := db.Queries.CompleteJob(ctx, claimed.ID); err != nil {
		t.Fatalf("CompleteJob() error = %v, want nil", err)
	}

	done, err := db.Queries.CountJobsByState(ctx, "done")
	if err != nil || done != 1 {
		t.Errorf("CountJobsByState(done) = %d, %v, want 1, nil", done, err)
	}
}
