package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrVersionNotFound = errors.New("version not found")
)

type RecordRow struct {
	ID        int
	Version   int
	Data      map[string]string
	CreatedAt time.Time
}

type VersionInfo struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// RecordRepository defines storage operations for versioned records
type RecordRepository interface {
	GetLatestVersion(ctx context.Context, id int) (*RecordRow, error)
	GetVersion(ctx context.Context, id, version int) (*RecordRow, error)
	GetVersionAtTime(ctx context.Context, id int, at time.Time) (*RecordRow, error)
	ListVersions(ctx context.Context, id int) ([]VersionInfo, error)
	CreateVersion(ctx context.Context, id int, data map[string]string) (*RecordRow, error)
	GetNextVersion(ctx context.Context, id int) (int, error)
}

// Querier abstracts sql.DB and sql.Tx so we can use either
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type SQLiteRepository struct {
	db Querier
}

func NewSQLiteRepository(db Querier) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) GetLatestVersion(ctx context.Context, id int) (*RecordRow, error) {
	query := `
		SELECT id, version, data, created_at 
		FROM records WHERE id = ? 
		ORDER BY version DESC LIMIT 1`

	return r.scanRow(r.db.QueryRowContext(ctx, query, id))
}

func (r *SQLiteRepository) GetVersion(ctx context.Context, id, version int) (*RecordRow, error) {
	query := `SELECT id, version, data, created_at FROM records WHERE id = ? AND version = ?`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id, version))
}

// GetVersionAtTime finds what the record looked like at a given point in time
// really useful for auditing - "what did we know on date X?"
func (r *SQLiteRepository) GetVersionAtTime(ctx context.Context, id int, at time.Time) (*RecordRow, error) {
	query := `
		SELECT id, version, data, created_at 
		FROM records 
		WHERE id = ? AND created_at <= ? 
		ORDER BY version DESC LIMIT 1`

	return r.scanRow(r.db.QueryRowContext(ctx, query, id, at))
}

func (r *SQLiteRepository) ListVersions(ctx context.Context, id int) ([]VersionInfo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT version, created_at FROM records WHERE id = ? ORDER BY version ASC`, id)
	if err != nil {
		return nil, fmt.Errorf("querying versions: %w", err)
	}
	defer rows.Close()

	var versions []VersionInfo
	for rows.Next() {
		var v VersionInfo
		if err := rows.Scan(&v.Version, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning version: %w", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, ErrRecordNotFound
	}
	return versions, nil
}

func (r *SQLiteRepository) CreateVersion(ctx context.Context, id int, data map[string]string) (*RecordRow, error) {
	nextVer, err := r.GetNextVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling data: %w", err)
	}

	now := time.Now().UTC()
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO records (id, version, data, created_at) VALUES (?, ?, ?, ?)`,
		id, nextVer, string(jsonData), now)
	if err != nil {
		return nil, fmt.Errorf("inserting record: %w", err)
	}

	return &RecordRow{ID: id, Version: nextVer, Data: data, CreatedAt: now}, nil
}

func (r *SQLiteRepository) GetNextVersion(ctx context.Context, id int) (int, error) {
	var next int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM records WHERE id = ?`, id).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("getting next version: %w", err)
	}
	return next, nil
}

func (r *SQLiteRepository) scanRow(row *sql.Row) (*RecordRow, error) {
	var rec RecordRow
	var jsonData string

	err := row.Scan(&rec.ID, &rec.Version, &jsonData, &rec.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}

	if err := json.Unmarshal([]byte(jsonData), &rec.Data); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}
	return &rec, nil
}
