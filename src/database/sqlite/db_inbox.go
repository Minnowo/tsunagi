package sqlite

import (
	"context"
	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
)

func (db *SqliteDatabase) AddInboxMessage(ctx context.Context, tx sqlx.Queryable, msg *database.InboxTable) error {

	id, err := database.InsertGetLastID(ctx, tx,
		`INSERT INTO TSU_INBOX (USER_ID, DEVICE_ID, CIPHER_TEXT) VALUES (?, ?, ?)`,
		msg.UserID, msg.DeviceID, msg.CipherText,
	)

	if err != nil {
		return err
	}

	msg.ID = id

	return nil
}
