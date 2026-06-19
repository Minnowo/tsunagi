package tcrypto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTokenKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, MacKeySize)
	key[0] = 42
	return key
}

func makePubKey(t *testing.T) []byte {
	t.Helper()
	kp, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	return kp.Public
}

func TestBuildAuthTokenReturnsToken(t *testing.T) {
	token, err := BuildAuthToken(makePubKey(t), time.Minute, makeTokenKey(t))
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, MacSize+TokenTimestampSize+NoiseKeySize, len(token))
}

func TestParseAuthTokenReturnsOriginalPubKey(t *testing.T) {
	pub := makePubKey(t)
	key := makeTokenKey(t)

	token, err := BuildAuthToken(pub, time.Minute, key)
	require.NoError(t, err)

	got, err := ParseAuthToken(token, key)
	require.NoError(t, err)
	assert.Equal(t, pub, got)
}

func TestParseAuthTokenWrongKeyFails(t *testing.T) {
	pub := makePubKey(t)
	key := makeTokenKey(t)

	token, err := BuildAuthToken(pub, time.Minute, key)
	require.NoError(t, err)

	wrongKey := make([]byte, MacKeySize)
	wrongKey[0] = 0xff

	_, err = ParseAuthToken(token, wrongKey)
	assert.ErrorIs(t, err, ErrMacMismatch)
}

func TestParseAuthTokenExpiredFails(t *testing.T) {
	pub := makePubKey(t)
	key := makeTokenKey(t)

	// Build a token that expired 1 second ago.
	token, err := BuildAuthToken(pub, -time.Second, key)
	require.NoError(t, err)

	_, err = ParseAuthToken(token, key)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestParseAuthTokenTamperedTagFails(t *testing.T) {
	pub := makePubKey(t)
	key := makeTokenKey(t)

	token, err := BuildAuthToken(pub, time.Minute, key)
	require.NoError(t, err)

	token[0] ^= 0xff // flip bits in the MAC tag

	_, err = ParseAuthToken(token, key)
	assert.ErrorIs(t, err, ErrMacMismatch)
}

func TestParseAuthTokenTooShortFails(t *testing.T) {
	key := makeTokenKey(t)
	_, err := ParseAuthToken([]byte("short"), key)
	assert.ErrorIs(t, err, ErrInvalidTokenLength)
}

func TestBuildAuthTokenWrongPubKeySizeFails(t *testing.T) {
	key := makeTokenKey(t)
	_, err := BuildAuthToken([]byte("tooshort"), time.Minute, key)
	assert.ErrorIs(t, err, ErrInvalidKeySize)
}
