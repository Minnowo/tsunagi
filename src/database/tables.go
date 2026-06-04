package database

import (
	"time"
	"tsunagi/src/data"
)

// VersionTable is the TSU_VERSION table.
type VersionTable struct {
	Version Version `db:"version"`
}

// UserTable is the TSU_USER table.
type UserTable struct {
	ID              int64           `db:"id"`
	Identifier      data.Identifier `db:"identifier"`
	PubKey          []byte          `db:"pub_key"`
	EncryptedPriKey []byte          `db:"enc_pri_key"`
	Created         time.Time       `db:"created"`
}

// DeviceTable is the TSU_DEVICE table.
type DeviceTable struct {
	ID              int64           `db:"id"`
	UserID          int64           `db:"user_id"`
	Identifier      data.Identifier `db:"identifier"`
	PubKey          []byte          `db:"pub_key"`
	EncryptedPriKey []byte          `db:"enc_pri_key"`
	Created         time.Time       `db:"created"`
}

// InboxTable is the TSU_INBOX table.
// This table contains messages incoming for a specific user/device.
type InboxTable struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	DeviceID   int64     `db:"device_id"`
	CipherText []byte    `db:"cipher_text"`
	Created    time.Time `db:"created"`
}
