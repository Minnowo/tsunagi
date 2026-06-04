package tcrypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plaintext := []byte("hello, world")
	password := []byte("supersecret")

	enc, err := EncryptWithPassword(plaintext, password)
	require.NoError(t, err)

	got, err := DecryptWithPassword(enc, password)
	require.NoError(t, err)

	assert.Equal(t, plaintext, got)
}

func TestDecryptWrongPassword(t *testing.T) {
	enc, err := EncryptWithPassword([]byte("secret data"), []byte("correctpassword"))
	require.NoError(t, err)

	_, err = DecryptWithPassword(enc, []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestDecryptCorruptedCiphertext(t *testing.T) {
	enc, err := EncryptWithPassword([]byte("data"), []byte("password"))
	require.NoError(t, err)

	enc.Ciphertext[0] ^= 0xff

	_, err = DecryptWithPassword(enc, []byte("password"))
	assert.Error(t, err)
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	enc, err := EncryptWithPassword([]byte{}, []byte("password"))
	require.NoError(t, err)

	got, err := DecryptWithPassword(enc, []byte("password"))
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestEncryptProducesUniqueNonceAndSalt(t *testing.T) {
	password := []byte("password")
	data := []byte("data")

	enc1, err := EncryptWithPassword(data, password)
	require.NoError(t, err)

	enc2, err := EncryptWithPassword(data, password)
	require.NoError(t, err)

	assert.NotEqual(t, enc1.Salt, enc2.Salt, "salts should differ across encryptions")
	assert.NotEqual(t, enc1.Nonce, enc2.Nonce, "nonces should differ across encryptions")
}

func TestEncryptLargePlaintext(t *testing.T) {
	plaintext := bytes.Repeat([]byte("A"), 1<<20) // 1 MiB
	password := []byte("password")

	enc, err := EncryptWithPassword(plaintext, password)
	require.NoError(t, err)

	got, err := DecryptWithPassword(enc, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}
