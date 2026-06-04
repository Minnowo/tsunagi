package data

type User struct {
	ID      Identifier `json:"id"`
	PubKey  []byte     `json:"pubkey"`
	PriKey  []byte     `json:"-"`
	Devices []Device   `json:"devices"`
}

type Device struct {
	ID     Identifier `json:"id"`
	UserID Identifier `json:"user_id"`
	PubKey []byte     `json:"pubkey"`
	PriKey []byte     `json:"-"`
}
