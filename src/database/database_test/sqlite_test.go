package database_test

import (
	"testing"
	"tsunagi/src/database"
	"tsunagi/src/database/sqlite"

	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {

	RunSuite(t, func(t *testing.T) database.DB {

		dir := t.TempDir()

		conn, err := sqlite.OpenSqliteDatabase(t.Context(), dir+"/sqlite_test.sqlite")
		require.Nil(t, err)

		err = conn.Migrate(t.Context())
		require.Nil(t, err)

		return conn
	})
}
