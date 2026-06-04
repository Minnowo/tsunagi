// Package database_test contains a reusable test suite for the database.DB interface.
// Each DB implementation wires it up by calling RunSuite with a factory that returns
// a freshly migrated, empty database.
package database_test

import (
	"context"
	"testing"
	"tsunagi/src/data"
	"tsunagi/src/database"

	"github.com/stretchr/testify/require"
	"github.com/vinovest/sqlx"
)

// NewDB is a factory that returns a freshly migrated, empty DB ready for testing.
type NewDB func(t *testing.T) database.DB

// RunSuite runs the full DB interface test suite against the given implementation.
func RunSuite(t *testing.T, newDB NewDB) {
	t.Helper()
	t.Run("User", func(t *testing.T) { runUserSuite(t, newDB) })
	t.Run("Device", func(t *testing.T) { runDeviceSuite(t, newDB) })
	// t.Run("Inbox", func(t *testing.T) { runInboxSuite(t, newDB) })
	// t.Run("Transact", func(t *testing.T) { runTransactSuite(t, newDB) })
}

// --- shared helpers ---

func newIdentifier(t *testing.T) data.Identifier {
	t.Helper()
	var id data.Identifier
	id.GenNew()
	return id
}

func mustAddUser(t *testing.T, ctx context.Context, db database.DB) *database.UserTable {
	t.Helper()
	user := &database.UserTable{
		Identifier:      newIdentifier(t),
		PubKey:          []byte("pubkey"),
		EncryptedPriKey: []byte("encprikey"),
	}
	require.NoError(t, db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
		return db.AddUser(ctx, tx, user)
	}))
	return user
}

func mustAddDevice(t *testing.T, ctx context.Context, db database.DB, userID int64) *database.DeviceTable {
	t.Helper()
	device := &database.DeviceTable{
		UserID:          userID,
		Identifier:      newIdentifier(t),
		PubKey:          []byte("devpubkey"),
		EncryptedPriKey: []byte("devencprikey"),
	}
	require.NoError(t, db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
		return db.AddDevice(ctx, tx, device)
	}))
	return device
}
