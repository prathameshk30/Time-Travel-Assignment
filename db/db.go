package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func DefaultConfig() Config {
	return Config{
		Path:            "timetravel.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

type Database struct {
	db   *sql.DB
	once sync.Once
}

// New opens a database connection and configures it for our use case
func New(cfg Config) (*Database, error) {
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// quick sanity check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	// sqlite specific pragmas - enabling WAL gives us better read concurrency
	// which matters when we're doing lots of version queries
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("pragma fk: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("pragma wal: %w", err)
	}

	return &Database{db: db}, nil
}

// Initialize sets up tables. Safe to call multiple times.
func (d *Database) Initialize(ctx context.Context) error {
	var initErr error
	d.once.Do(func() {
		initErr = d.createTables(ctx)
	})
	return initErr
}

func (d *Database) createTables(ctx context.Context) error {
	// Using composite PK on (id, version) - this way we can efficiently
	// query all versions for a given record while ensuring uniqueness
	schema := `
		CREATE TABLE IF NOT EXISTS records (
			id          INTEGER NOT NULL,
			version     INTEGER NOT NULL,
			data        TEXT NOT NULL,
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id, version)
		);
		CREATE INDEX IF NOT EXISTS idx_records_id ON records(id);
		CREATE INDEX IF NOT EXISTS idx_records_created_at ON records(created_at);
	`

	_, err := d.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	return nil
}

func (d *Database) DB() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}
