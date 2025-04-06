package integration

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/s-hammon/volta/internal/testing/testdb"
)

const (
	migrationsDir = "../../../sql/schema"
)

func TestPostgres(t *testing.T) {
	t.Parallel()

	db, cleanup, err := testdb.NewPostgres()
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	t.Cleanup(cleanup)
	if err = db.Ping(); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}

	testDatabase(t, database.DialectPostgres, db)
}

type collected struct {
	fullpath string
	version  int64
}

func collectMigrations(t *testing.T, dir string) []collected {
	t.Helper()

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	all := make([]collected, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			t.Fatalf("unexpected file in migrations dir: %s", f.Name())
		}
		v, err := goose.NumericComponent(f.Name())
		if err != nil {
			t.Fatalf("failed to parse version from dir: %v", err)
		}
		all = append(all, collected{
			fullpath: filepath.Base(f.Name()),
			version:  v,
		})
	}

	return all
}

func testDatabase(t *testing.T, dialect database.Dialect, db *sql.DB) {
	t.Helper()

	ctx := context.Background()
	wantFiles := collectMigrations(t, migrationsDir)
	p, err := goose.NewProvider(dialect, db, os.DirFS(migrationsDir))
	if err != nil {
		t.Fatalf("failed to create goose provider: %v", err)
	}
	if len(wantFiles) != len(p.ListSources()) {
		t.Fatalf("unexpected number of migrations: got %d, want %d", len(p.ListSources()), len(wantFiles))
	}
	res, err := p.Up(ctx)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	if len(res) != len(wantFiles) {
		t.Fatalf("unexpected number of migrations run: got %d, want %d", len(res), len(wantFiles))
	}

	for i, r := range res {
		if wantFiles[i].fullpath != r.Source.Path {
			t.Fatalf("unexpected migration run: got %s, want %s", r.Source.Path, wantFiles[i].fullpath)
		}
		if wantFiles[i].version != r.Source.Version {
			t.Fatalf("unexpected migration version: got %d, want %d", r.Source.Version, wantFiles[i].version)
		}
	}
	currentVersion, err := p.GetDBVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}
	if len(wantFiles) != int(currentVersion) {
		t.Fatalf("unexpected current version: got %d, want %d", currentVersion, len(wantFiles))
	}

	res, err = p.DownTo(ctx, 0)
	if err != nil {
		t.Fatalf("failed to run down migrations: %v", err)
	}
	if len(res) != len(wantFiles) {
		t.Fatalf("unexpected number of down migrations run: got %d, want %d", len(res), len(wantFiles))
	}
	currentVersion, err = p.GetDBVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}
	if int(currentVersion) != 0 {
		t.Fatalf("unexpected current version after down migrations: got %d, want 0", currentVersion)
	}

	for i := range len(wantFiles) {
		result, err := p.UpByOne(ctx)
		if err != nil {
			t.Fatalf("failed to run up by one migration: %v", err)
		}
		if errors.Is(err, goose.ErrNoNextVersion) {
			break
		}
		if wantFiles[i].fullpath != result.Source.Path {
			t.Fatalf("unexpected migration run: got %s, want %s", result.Source.Path, wantFiles[i].fullpath)
		}
	}

	currentVersion, err = p.GetDBVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}
	if len(wantFiles) != int(currentVersion) {
		t.Fatalf("unexpected current version: got %d, want %d", currentVersion, len(wantFiles))
	}
}
