package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE records (
			id INTEGER NOT NULL,
			version INTEGER NOT NULL,
			data TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id, version)
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return db
}

func TestCreateVersion_NewRecord(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	data := map[string]string{"name": "Test Business", "status": "active"}

	rec, err := repo.CreateVersion(context.Background(), 1, data)
	if err != nil {
		t.Fatalf("create version: %v", err)
	}

	if rec.ID != 1 || rec.Version != 1 {
		t.Errorf("got id=%d version=%d, want 1 and 1", rec.ID, rec.Version)
	}
	if rec.Data["name"] != "Test Business" {
		t.Errorf("name = %q, want Test Business", rec.Data["name"])
	}
}

func TestCreateVersion_MultipleVersions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	v1, _ := repo.CreateVersion(ctx, 1, map[string]string{"status": "pending"})
	if v1.Version != 1 {
		t.Errorf("first version = %d, want 1", v1.Version)
	}

	v2, _ := repo.CreateVersion(ctx, 1, map[string]string{"status": "active"})
	if v2.Version != 2 {
		t.Errorf("second version = %d, want 2", v2.Version)
	}

	versions, _ := repo.ListVersions(ctx, 1)
	if len(versions) != 2 {
		t.Errorf("got %d versions, want 2", len(versions))
	}
}

func TestGetLatestVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	repo.CreateVersion(ctx, 1, map[string]string{"val": "first"})
	repo.CreateVersion(ctx, 1, map[string]string{"val": "second"})
	repo.CreateVersion(ctx, 1, map[string]string{"val": "third"})

	latest, err := repo.GetLatestVersion(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Version != 3 || latest.Data["val"] != "third" {
		t.Errorf("got version %d val=%q, want 3 and third", latest.Version, latest.Data["val"])
	}
}

func TestGetVersion_Specific(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	repo.CreateVersion(ctx, 1, map[string]string{"stage": "alpha"})
	repo.CreateVersion(ctx, 1, map[string]string{"stage": "beta"})
	repo.CreateVersion(ctx, 1, map[string]string{"stage": "prod"})

	v2, _ := repo.GetVersion(ctx, 1, 2)
	if v2.Data["stage"] != "beta" {
		t.Errorf("stage = %q, want beta", v2.Data["stage"])
	}
}

func TestGetVersion_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	_, err := repo.GetLatestVersion(context.Background(), 999)
	if err != ErrRecordNotFound {
		t.Errorf("err = %v, want ErrRecordNotFound", err)
	}
}

func TestGetVersionAtTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// insert with specific timestamps
	db.Exec("INSERT INTO records (id, version, data, created_at) VALUES (1, 1, '{\"state\":\"v1\"}', '2026-01-01T10:00:00Z')")
	db.Exec("INSERT INTO records (id, version, data, created_at) VALUES (1, 2, '{\"state\":\"v2\"}', '2026-01-15T10:00:00Z')")
	db.Exec("INSERT INTO records (id, version, data, created_at) VALUES (1, 3, '{\"state\":\"v3\"}', '2026-02-01T10:00:00Z')")

	// query between v1 and v2
	ts, _ := time.Parse(time.RFC3339, "2026-01-10T10:00:00Z")
	rec, _ := repo.GetVersionAtTime(ctx, 1, ts)
	if rec.Version != 1 {
		t.Errorf("at Jan 10 got version %d, want 1", rec.Version)
	}

	// query between v2 and v3
	ts2, _ := time.Parse(time.RFC3339, "2026-01-20T10:00:00Z")
	rec2, _ := repo.GetVersionAtTime(ctx, 1, ts2)
	if rec2.Version != 2 {
		t.Errorf("at Jan 20 got version %d, want 2", rec2.Version)
	}
}

func TestListVersions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	repo.CreateVersion(ctx, 1, map[string]string{"a": "1"})
	repo.CreateVersion(ctx, 1, map[string]string{"a": "2"})
	repo.CreateVersion(ctx, 2, map[string]string{"b": "1"})

	v, _ := repo.ListVersions(ctx, 1)
	if len(v) != 2 {
		t.Errorf("record 1 has %d versions, want 2", len(v))
	}
	// should be ascending order
	if v[0].Version != 1 || v[1].Version != 2 {
		t.Errorf("versions not in order: %+v", v)
	}
}

func TestListVersions_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	_, err := repo.ListVersions(context.Background(), 999)
	if err != ErrRecordNotFound {
		t.Errorf("err = %v, want ErrRecordNotFound", err)
	}
}

func TestMultipleRecords_Isolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	repo.CreateVersion(ctx, 1, map[string]string{"rec": "one"})
	repo.CreateVersion(ctx, 2, map[string]string{"rec": "two"})
	repo.CreateVersion(ctx, 1, map[string]string{"rec": "one-v2"})

	v1, _ := repo.ListVersions(ctx, 1)
	v2, _ := repo.ListVersions(ctx, 2)

	if len(v1) != 2 {
		t.Errorf("record 1: %d versions, want 2", len(v1))
	}
	if len(v2) != 1 {
		t.Errorf("record 2: %d versions, want 1", len(v2))
	}
}
