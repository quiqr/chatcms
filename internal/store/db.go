package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool    *pgxpool.Pool
	url     string
	Queries *Queries
}

func Open(ctx context.Context, url string) (*DB, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("the database connection string could not be parsed")
	}
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connecting to the database failed: %w", err)
	}

	return &DB{pool: pool, url: url, Queries: New(pool)}, nil
}

// Within is this database with every query running inside a transaction. It
// is how a caller counts something and acts on the count without a second
// caller slipping between the two.
func (db *DB) Within(tx pgx.Tx) *DB {
	return &DB{pool: db.pool, url: db.url, Queries: db.Queries.WithTx(tx)}
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) Close() {
	db.pool.Close()
}
