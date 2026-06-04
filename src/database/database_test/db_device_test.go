package database_test

import (
	"context"
	"testing"
	"tsunagi/src/database"

	"github.com/stretchr/testify/assert"
	"github.com/vinovest/sqlx"
)

func runDeviceSuite(t *testing.T, newDB NewDB) {
	t.Run("AddDevice_SetsID", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		user := mustAddUser(t, ctx, db)
		device := mustAddDevice(t, ctx, db, user.ID)

		assert.NotZero(t, device.ID)
	})


	t.Run("AddDevice_UnknownUser_Errors", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		device := &database.DeviceTable{
			UserID:          999999,
			Identifier:      newIdentifier(t),
			PubKey:          []byte("pub"),
			EncryptedPriKey: []byte("enc"),
		}

		err := db.Transact(ctx, func(ctx context.Context, tx sqlx.Queryable) error {
			return db.AddDevice(ctx, tx, device)
		})

		assert.Error(t, err)
	})

	t.Run("AddMultipleDevices_UniqueIDs", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		user := mustAddUser(t, ctx, db)
		a := mustAddDevice(t, ctx, db, user.ID)
		b := mustAddDevice(t, ctx, db, user.ID)

		assert.NotEqual(t, a.ID, b.ID)
	})

	t.Run("AddDevice_MultipleUsers", func(t *testing.T) {
		db := newDB(t)
		ctx := context.Background()

		u1 := mustAddUser(t, ctx, db)
		u2 := mustAddUser(t, ctx, db)

		d1 := mustAddDevice(t, ctx, db, u1.ID)
		d2 := mustAddDevice(t, ctx, db, u2.ID)

		assert.NotEqual(t, d1.ID, d2.ID)
	})
}
