package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"github.com/quiqr/chatcms/db"
)

var ErrDatabaseAhead = errors.New("the database schema is newer than this binary")

func newProvider(sqlDB *sql.DB) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(1, 120))
	if err != nil {
		return nil, fmt.Errorf("creating the migration lock failed: %w", err)
	}
	return goose.NewProvider(goose.DialectPostgres, sqlDB, db.Migrations(),
		goose.WithSessionLocker(locker))
}

func (d *DB) withProvider(fn func(*goose.Provider) error) error {
	sqlDB := stdlib.OpenDBFromPool(d.pool)
	defer sqlDB.Close()

	provider, err := newProvider(sqlDB)
	if err != nil {
		return err
	}
	return fn(provider)
}

func (d *DB) Migrate(ctx context.Context) error {
	return d.withProvider(func(provider *goose.Provider) error {
		applied, err := provider.GetDBVersion(ctx)
		if err != nil {
			return fmt.Errorf("reading the applied schema version failed: %w", err)
		}

		known := latestVersion(provider)
		if applied > known {
			return fmt.Errorf("%w: database is at version %d, this binary knows %d",
				ErrDatabaseAhead, applied, known)
		}

		if _, err := provider.Up(ctx); err != nil {
			return fmt.Errorf("applying migrations failed: %w", err)
		}
		return nil
	})
}

// MigrateDown rolls back the most recently applied migration.
func (d *DB) MigrateDown(ctx context.Context) error {
	return d.withProvider(func(provider *goose.Provider) error {
		_, err := provider.Down(ctx)
		return err
	})
}

// MigrateDownTo rolls back to a version, with zero meaning an empty schema.
// It exists for tests that prove a migration is reversible.
func (d *DB) MigrateDownTo(ctx context.Context, version int64) error {
	return d.withProvider(func(provider *goose.Provider) error {
		_, err := provider.DownTo(ctx, version)
		return err
	})
}

func (d *DB) SchemaVersion(ctx context.Context) (int64, error) {
	var version int64
	err := d.withProvider(func(provider *goose.Provider) error {
		applied, err := provider.GetDBVersion(ctx)
		version = applied
		return err
	})
	return version, err
}

func (d *DB) KnownSchemaVersion() (int64, error) {
	var version int64
	err := d.withProvider(func(provider *goose.Provider) error {
		version = latestVersion(provider)
		return nil
	})
	return version, err
}

func latestVersion(provider *goose.Provider) int64 {
	var latest int64
	for _, source := range provider.ListSources() {
		if source.Version > latest {
			latest = source.Version
		}
	}
	return latest
}
