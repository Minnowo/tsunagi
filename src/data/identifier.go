package data

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Identifier [32]byte

func (id Identifier) String() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

func (id *Identifier) FromBytes(buf []byte) error {

	if len(id) == 0 {
		return fmt.Errorf("nil Identifier receiver")
	}

	if len(buf) != len(id) {
		return fmt.Errorf("invalid identifier length: must be 32 bytes")
	}

	copy(id[:], buf)
	return nil
}

func (id *Identifier) FromString(s string) error {

	if len(id) == 0 {
		return fmt.Errorf("nil Identifier receiver")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(s)

	if err != nil {
		return err
	}

	if len(decoded) != len(id) {
		return fmt.Errorf("invalid identifier length: must be 32 bytes")
	}

	copy(id[:], decoded)
	return nil
}

func (id Identifier) MarshalJSON() ([]byte, error) {
	encoded := base64.RawURLEncoding.EncodeToString(id[:])
	return json.Marshal(encoded)
}

func (id *Identifier) UnmarshalJSON(data []byte) error {
	var s string

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return id.FromString(s)
}
