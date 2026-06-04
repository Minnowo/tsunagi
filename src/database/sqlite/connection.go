package sqlite

import (
	"context"
	"strings"
	"time"
	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
	"github.com/vinovest/sqlx/reflectx"
)

func openSqliteDatabase(ctx context.Context, driver, connString string) (db *SqliteDatabase, err error) {

	conn, err := sqlx.ConnectContext(ctx, driver, connString)

	if err != nil {
		return nil, err
	}

	conn.Mapper = reflectx.NewMapperTagColFunc("db", strings.ToUpper, strings.ToUpper, strings.ToUpper)
	conn.SetMaxOpenConns(5)
	conn.SetConnMaxLifetime(time.Minute * 10)

	_, err = conn.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	db = &SqliteDatabase{
		BaseDB: database.BaseDB{
			DB: conn,
		},
		version: database.VERSION_UNKNOWN,
	}

	return db, nil
}
