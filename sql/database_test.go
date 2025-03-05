package sql_test

import (
	"testing"

	"maragu.dev/is"

	"app/sqltest"
)

func TestDatabase_Migrate(t *testing.T) {
	t.Run("can migrate down and back up", func(t *testing.T) {
		db := sqltest.NewDatabase(t)

		err := db.MigrateDown(t.Context())
		is.NotError(t, err)

		err = db.MigrateUp(t.Context())
		is.NotError(t, err)

		var version string
		err = db.H.Get(t.Context(), &version, "select version from migrations")
		is.NotError(t, err)
		is.Equal(t, "1741176647-documents", version)
	})
}
