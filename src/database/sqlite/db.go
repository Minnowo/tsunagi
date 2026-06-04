package sqlite

import (
	"tsunagi/src/database"
)

type SqliteDatabase struct {
	database.BaseDB
	version database.Version
}
