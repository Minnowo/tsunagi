package data

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

var (
	ErrInvalidIdentifier = errors.New("invalid identifier")
)

// var strEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)
var strEncoding = base64.StdEncoding

type Identifier [tcrypto.NoiseKeySize]byte

func (id Identifier) String() string {
	return strEncoding.EncodeToString(id[:])
}

func (id *Identifier) FromString(s string) error {

	if len(id) == 0 {
		panic("Identifier.FromString: nil Identifier receiver")
	}

	decoded, err := strEncoding.DecodeString(s)

	if err != nil {
		return err
	}

	if len(decoded) != len(id) {
		return ErrInvalidIdentifier
	}

	copy(id[:], decoded)
	return nil
}

func (id *Identifier) FromBytes(buf []byte) error {

	if len(id) == 0 {
		panic("Identifier.FromBytes: nil Identifier receiver")
	}

	if len(buf) != len(id) {
		return ErrInvalidIdentifier
	}

	copy(id[:], buf)
	return nil
}

func (id Identifier) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *Identifier) Scan(value any) error {
	if id == nil {
		panic("Identifier.Scan: nil Identifier receiver")
	}

	switch v := value.(type) {
	case []byte:
		if len(v) != len(id) {
			return ErrInvalidIdentifier
		}
		copy(id[:], v)
		return nil

	default:
		return fmt.Errorf("Identifier.Scan: unsupported type for Identifier: %T", value)
	}
}

func (id Identifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *Identifier) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return id.FromString(s)
}
