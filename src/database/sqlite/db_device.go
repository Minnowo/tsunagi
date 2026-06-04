package sqlite

import (
	"context"
	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
)

func (db *SqliteDatabase) AddDevice(ctx context.Context, tx sqlx.Queryable, device *database.DeviceTable) error {

	id, err := database.InsertGetLastID(ctx, tx,
		`INSERT INTO TSU_DEVICE (USER_ID, IDENTIFIER, PUB_KEY, ENC_PRI_KEY) VALUES (?, ?, ?, ?)`,
		device.UserID, device.Identifier, device.PubKey, device.EncryptedPriKey,
	)

	if err != nil {
		return err
	}

	device.ID = id

	return nil
}
