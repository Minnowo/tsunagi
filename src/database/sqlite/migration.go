package sqlite

import (
	"context"
	"fmt"
	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
)

var sqliteUpMigrations = []database.Migration{
	database.NewMigration(database.VERSION_NONE, 1, migrateNoneTo1),
}

func GetMigrationMaxVersion() database.Version {

	last := len(sqliteUpMigrations) - 1

	return sqliteUpMigrations[last].ToVersion
}

func (db *SqliteDatabase) Migrate(ctx context.Context) error {

	version, err := db.GetVersion(ctx)

	if err != nil {

		if SQLiteErrorNoSuchTable.Is(err) {
			version = database.VERSION_NONE
		} else {
			return err
		}
	}

	if version == database.VERSION_UNKNOWN {
		return fmt.Errorf(
			"the database version is unknown, an existing schema exists, but there is no version number: %w",
			database.ErrInvalidDatabaseVersion,
		)
	}

	if version > GetMigrationMaxVersion() {
		return fmt.Errorf(
			"the database version is larger than the maximum migration version: %w",
			database.ErrInvalidDatabaseVersion,
		)
	}

	db.version = version

	newVersion, err := database.RunMigrations(ctx, db, database.Version(version), sqliteUpMigrations)

	// should be used even if an error happens
	db.version = newVersion

	if err != nil {
		return err
	}

	return nil
}

func migrateNoneTo1(ctx context.Context, db database.DB, toVersion database.Version, migrateUp bool) error {

	return db.Transact(ctx, func(ctx context.Context, db sqlx.Queryable) error {

		_, err := db.Exec(`

			CREATE TABLE TSU_VERSION (
				VERSION INTEGER NOT NULL
			);

			CREATE TABLE TSU_USER (
				ID              INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
				IDENTIFIER      BLOB NOT NULL,
				PUB_KEY         BLOB NOT NULL,
				ENC_PRI_KEY     BLOB NOT NULL,
				CREATED         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

				UNIQUE(IDENTIFIER)
			);

			CREATE TABLE TSU_DEVICE (
				ID            INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
				USER_ID       INTEGER NOT NULL,
				IDENTIFIER    BLOB NOT NULL,
				PUB_KEY       BLOB NOT NULL,
				ENC_PRI_KEY   BLOB NOT NULL,
				CREATED       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

				UNIQUE(IDENTIFIER),
				UNIQUE (ID, USER_ID),
				FOREIGN KEY (USER_ID) REFERENCES TSU_USER(ID)
			);

			CREATE TABLE TSU_INBOX (
				ID            INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
				USER_ID       INTEGER NOT NULL,
				DEVICE_ID     INTEGER NOT NULL,
				CIPHER_TEXT   BLOB NOT NULL,
				CREATED       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

				FOREIGN KEY (DEVICE_ID, USER_ID) REFERENCES TSU_DEVICE(ID, USER_ID)
			);

			CREATE INDEX IDX_TSU_USER_IDENTIFIER ON TSU_USER(IDENTIFIER);
			CREATE INDEX IDX_TSU_DEVICE_IDENTIFIER ON TSU_DEVICE(IDENTIFIER);
		`)

		if err != nil {
			return err
		}

		_, err = db.Exec(
			`INSERT INTO TSU_VERSION (VERSION) VALUES (?)`,
			toVersion,
		)

		return err
	})
}
