package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
)

func (db *SqliteDatabase) GetVersion(ctx context.Context) (database.Version, error) {

	var tbl database.VersionTable

	query := `SELECT VERSION FROM TSU_VERSION LIMIT 1`

	err := db.DB.GetContext(ctx, &tbl, query)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return database.VERSION_UNKNOWN, nil
		}

		if SQLiteErrorNoSuchTable.Is(err) {
			return database.VERSION_NONE, err
		}

		return database.VERSION_UNKNOWN, nil
	}

	return tbl.Version, nil
}

func SetVersion(ctx context.Context, tx sqlx.Queryable, version database.Version) error {

	_, err := tx.ExecContext(ctx, "UPDATE TSU_VERSION SET VERSION = $1", int(version))

	return err
}
