package tcrypto

import (
	"errors"
	"hash"

	"golang.org/x/crypto/blake2s"
)

var (
	ErrInvalidMacSize = errors.New("invalid mac size")
	ErrMacMismatch    = errors.New("mac mismatch")
)

const (
	// MacSize is the number of bytes the mac hash output will be
	MacSize = blake2s.Size

	// MacKeySize is the number of bytes the mac key must be
	MacKeySize = blake2s.Size
)

func NewMAC(key []byte) (hash.Hash, error) {
	return blake2s.New256(key)
}
