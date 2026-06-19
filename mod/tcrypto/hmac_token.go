package tcrypto

import (
	"crypto/hmac"
	"encoding/binary"
	"errors"
	"time"
)

var (
	ErrInvalidTokenLength = errors.New("invalid token length")
	ErrTokenExpired       = errors.New("token has expired")
)

const (
	// how many bytes the timestamp takes in the auth token
	TokenTimestampSize = 8
)

// BuildAuthToken creates a mac signed auth token which is valid for the given duration.
func BuildAuthToken(pubKey []byte, validFor time.Duration, key []byte) ([]byte, error) {

	if len(pubKey) != NoiseKeySize {
		return nil, ErrInvalidKeySize
	}

	mac, err := NewMAC(key)

	if err != nil {
		return nil, err
	}

	var tsBuf [TokenTimestampSize]byte

	t := time.Now().Add(validFor).Unix()
	binary.BigEndian.PutUint64(tsBuf[:], uint64(t))

	mac.Write(tsBuf[:])
	mac.Write(pubKey)

	tag := mac.Sum(nil)

	// token layout: mac || ts || pubkey
	out := make([]byte, 0, len(tag)+len(tsBuf)+len(pubKey))
	out = append(out, tag...)
	out = append(out, tsBuf[:]...)
	out = append(out, pubKey...)

	return out, nil
}

// ParseAuthToken parses and validates the auth token.
// If the token is valid, the public key is returned with a nil error.
// Otherwise an error is returned.
func ParseAuthToken(token []byte, key []byte) ([]byte, error) {

	expectedLen := MacSize + TokenTimestampSize + NoiseKeySize

	if len(token) != expectedLen {
		return nil, ErrInvalidTokenLength
	}

	mac, err := NewMAC(key)

	if err != nil {
		return nil, err
	}

	tag := token[:MacSize]
	tsBuf := token[MacSize : MacSize+TokenTimestampSize]
	pubKey := token[MacSize+TokenTimestampSize:]

	mac.Write(tsBuf)
	mac.Write(pubKey)

	expectedTag := mac.Sum(nil)

	if !hmac.Equal(tag, expectedTag) {
		return nil, ErrMacMismatch
	}

	expires := binary.BigEndian.Uint64(tsBuf)
	now := time.Now().Unix()

	if now > int64(expires) {
		return nil, ErrTokenExpired
	}

	return pubKey, nil
}
