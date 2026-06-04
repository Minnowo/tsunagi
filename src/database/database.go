package database

import (
	"context"
	"fmt"
	"strconv"
	"tsunagi/src/data"

	"github.com/vinovest/sqlx"
)

var (
	ErrInvalidDatabaseVersion = fmt.Errorf("database version is invalid")
)

type Version int

const (
	VERSION_UNKNOWN Version = -1
	VERSION_NONE    Version = 0
)

func (v Version) String() string {
	return strconv.Itoa(int(v))
}

// DB is interface for accessing and manipulating data in database.
type DB interface {

	// Sqlx returns the underlying *sqlx.DB
	Sqlx() *sqlx.DB

	// Base returns the *BaseDB
	Base() *BaseDB

	// Transact opens a transaction which is rollback if the function returns an error, otherwise is commited.
	Transact(ctx context.Context, txFunc func(context.Context, sqlx.Queryable) error) error

	// Migrate runs migrations for this database
	Migrate(ctx context.Context) error

	AddUser(ctx context.Context, tx sqlx.Queryable, user *UserTable) error

	LoadUserByIdentifier(ctx context.Context, tx sqlx.Queryable, id data.Identifier, user *UserTable) error

	AddDevice(ctx context.Context, tx sqlx.Queryable, device *DeviceTable) error

	AddInboxMessage(ctx context.Context, tx sqlx.Queryable, msg *InboxTable) error
}

type BaseDB struct {
	DB *sqlx.DB
}

func (db *BaseDB) Sqlx() *sqlx.DB {
	return db.DB
}

func (db *BaseDB) Base() *BaseDB {
	return db
}

func (db *BaseDB) Close() error {
	return db.DB.Close()
}

func (db *BaseDB) Transact(ctx context.Context, txFunc func(context.Context, sqlx.Queryable) error) error {
	// The whole point of this method is that if you don't pass in the db.DB (*sqlx.DB) type,
	// you'll get a nil pointer exception and it's unclear why until you realize this.
	//
	// This method makes using sqlx.TransactContext safe by always calling it properly.
	return sqlx.TransactContext(ctx, db.DB, txFunc)
}

// InsertGetLastID executes an INSERT query on the given queryable and returns the auto-generated row ID.
func InsertGetLastID(ctx context.Context, q sqlx.Queryable, query string, args ...any) (int64, error) {

	result, err := q.ExecContext(ctx, query, args...)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
