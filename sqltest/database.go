package sqltest

import (
	"log/slog"
	"testing"

	sqlitevec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"maragu.dev/sqlh/sqltest"

	"app/sql"
)

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.t.Log(string(p))
	return len(p), nil
}

// NewDatabase for testing.
func NewDatabase(t *testing.T) *sql.Database {
	t.Helper()

	sqlitevec.Auto()

	// Load the helper from sqlh so we get cleanup etc.
	h := sqltest.NewHelper(t)
	db := sql.NewDatabase(sql.NewDatabaseOptions{
		Log: slog.New(slog.NewTextHandler(&testWriter{t: t}, nil)),
	})
	db.H = h

	if err := db.Connect(); err != nil {
		t.Fatal(err)
	}

	if err := db.MigrateUp(t.Context()); err != nil {
		t.Fatal(err)
	}

	return db
}
