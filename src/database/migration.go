package database

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// MigrationFunc runs a database migration.
// It must set the database version using the given toVersion value if it returns a nil error.
type MigrationFunc func(ctx context.Context, db DB, toVersion Version, migrateUp bool) error

type Migration struct {
	FromVersion   Version
	ToVersion     Version
	MigrationFunc MigrationFunc
}

func NewMigration(from, to Version, fn MigrationFunc) Migration {
	return Migration{
		FromVersion:   from,
		ToVersion:     to,
		MigrationFunc: fn,
	}
}

// RunMigrations runs the migrations from the current version until the max version is reached.
// Stops at the first error encountered, and returns the new database version and the error.
func RunMigrations(ctx context.Context, db DB, curVersion Version, migrations []Migration) (Version, error) {

	for _, migration := range migrations {

		if curVersion != migration.FromVersion {
			continue
		}

		log.Info().
			Int("from", int(migration.FromVersion)).
			Int("to", int(migration.ToVersion)).
			Msg("Migrating database")

		if err := migration.MigrationFunc(ctx, db, migration.ToVersion, true); err != nil {
			return curVersion, fmt.Errorf(
				"failed to run migration from %s to %s: %w",
				migration.FromVersion,
				migration.ToVersion,
				err,
			)
		}

		curVersion = migration.ToVersion
	}

	return curVersion, nil

}
