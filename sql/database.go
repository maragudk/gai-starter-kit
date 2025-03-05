package sql

import (
	"context"
	"log/slog"

	sqlitevec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"maragu.dev/errors"
	"maragu.dev/sqlh/sql"
)

type Database struct {
	H   *sql.Helper
	log *slog.Logger
}

type NewDatabaseOptions struct {
	Log  *slog.Logger
	Path string
}

// NewDatabase with the given options.
// If no logger is provided, logs are discarded.
func NewDatabase(opts NewDatabaseOptions) *Database {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}

	return &Database{
		H: sql.NewHelper(sql.NewHelperOptions{
			Log:  opts.Log,
			Path: opts.Path,
		}),
		log: opts.Log,
	}
}

func (d *Database) Connect() error {
	sqlitevec.Auto()

	if err := d.H.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	var vecVersion string
	if err := d.H.Get(context.Background(), &vecVersion, "select vec_version()"); err != nil {
		return errors.Wrap(err, "error getting vec version")
	}
	d.log.Info("Loaded SQLite vector search extension", "version", vecVersion)

	return nil
}

func (d *Database) MigrateUp(ctx context.Context) error {
	if err := d.H.MigrateUp(ctx); err != nil {
		return err
	}

	return nil
}

func (d *Database) MigrateDown(ctx context.Context) error {
	if err := d.H.MigrateDown(ctx); err != nil {
		return err
	}

	return nil
}
