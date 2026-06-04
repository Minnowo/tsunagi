package database_test

import (
	"context"
	"testing"
	"tsunagi/src/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinovest/sqlx"
)

func runUserSuite(t *testing.T, newDB NewDB) {
	t.Run("AddUser_SetsID", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		user := mustAddUser(t, ctx, db)

		assert.NotZero(t, user.ID)
	})

	t.Run("AddUser_StoresFields", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		user := mustAddUser(t, ctx, db)

		var loaded database.UserTable
		require.NoError(t, db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
			return db.LoadUserByIdentifier(ctx, tx, user.Identifier, &loaded)
		}))

		assert.Equal(t, user.Identifier, loaded.Identifier)
		assert.Equal(t, user.PubKey, loaded.PubKey)
		assert.Equal(t, user.EncryptedPriKey, loaded.EncryptedPriKey)
	})

	t.Run("AddUser_DuplicateIdentifier_Errors", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		user := mustAddUser(t, ctx, db)

		dup := &database.UserTable{
			Identifier:      user.Identifier,
			PubKey:          []byte("other"),
			EncryptedPriKey: []byte("other"),
		}

		err := db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
			return db.AddUser(ctx, tx, dup)
		})

		assert.Error(t, err)
	})

	t.Run("LoadUserByIdentifier_NotFound_Errors", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		var loaded database.UserTable
		err := db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
			return db.LoadUserByIdentifier(ctx, tx, newIdentifier(t), &loaded)
		})

		assert.Error(t, err)
	})

	t.Run("AddMultipleUsers_UniqueIDs", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		a := mustAddUser(t, ctx, db)
		b := mustAddUser(t, ctx, db)

		assert.NotEqual(t, a.ID, b.ID)
	})
}
