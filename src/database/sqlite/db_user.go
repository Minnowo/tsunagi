package sqlite

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/database"

	"github.com/vinovest/sqlx"
)

func (db *SqliteDatabase) AddUser(ctx context.Context, tx sqlx.Queryable, user *database.UserTable) error {

	id, err := database.InsertGetLastID(ctx, tx,
		`INSERT INTO TSU_USER (IDENTIFIER, PUB_KEY, ENC_PRI_KEY) VALUES (?, ?, ?)`,
		user.Identifier, user.PubKey, user.EncryptedPriKey,
	)

	if err != nil {
		return err
	}

	user.ID = id

	return nil
}

func (db *SqliteDatabase) LoadUserByIdentifier(ctx context.Context, tx sqlx.Queryable, id data.Identifier, user *database.UserTable) error {

	return tx.GetContext(ctx, user, `SELECT * FROM TSU_USER WHERE IDENTIFIER = ?`, id)
}
