package tcrypto

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestToBytesFromBytesRoundtrip(t *testing.T) {
	enc, err := EncryptWithPassword([]byte("roundtrip test"), []byte("pass"))
	require.NoError(t, err)

	b := enc.ToBytes()

	var dec PasswordEncryptedData
	err = dec.FromBytes(b)
	require.NoError(t, err)

	assert.Equal(t, enc.Salt, dec.Salt)
	assert.Equal(t, enc.Nonce, dec.Nonce)
	assert.Equal(t, enc.Ciphertext, dec.Ciphertext)
}

func TestFromBytesDecryptRoundtrip(t *testing.T) {
	plaintext := []byte("serialized then decrypted")
	password := []byte("pass")

	enc, err := EncryptWithPassword(plaintext, password)
	require.NoError(t, err)

	var dec PasswordEncryptedData
	err = dec.FromBytes(enc.ToBytes())
	require.NoError(t, err)

	got, err := DecryptWithPassword(&dec, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestToBytesExpectedLength(t *testing.T) {
	var p PasswordEncryptedData
	p.Ciphertext = []byte("ct")

	b := p.ToBytes()
	expected := 4 + len(p.Salt) + 4 + len(p.Nonce) + 4 + len(p.Ciphertext)
	assert.Len(t, b, expected)
}

func TestFromBytesTruncatedInputs(t *testing.T) {
	enc, err := EncryptWithPassword([]byte("data"), []byte("pass"))
	require.NoError(t, err)
	b := enc.ToBytes()

	cases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"partial salt length field", b[:2]},
		{"partial salt body", b[:5]},
		{"partial nonce length field", b[:4+32+2]},
		{"partial nonce body", b[:4+32+4+1]},
		{"partial ciphertext length field", b[:4+32+4+chacha20poly1305.NonceSizeX+2]},
		{"partial ciphertext body", b[:len(b)-1]},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p PasswordEncryptedData
			assert.Error(t, p.FromBytes(tc.data))
		})
	}
}

func TestFromBytesInvalidSaltSize(t *testing.T) {
	// salt length field set to 64, which exceeds the [32]byte Salt field
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 64)

	var p PasswordEncryptedData
	err := p.FromBytes(buf)
	assert.Error(t, err)
}

func TestFromBytesInvalidNonceSize(t *testing.T) {
	// valid salt section, then nonce length field set to 255 (> NonceSizeX)
	buf := make([]byte, 4+32+4)
	binary.LittleEndian.PutUint32(buf[0:], 32)     // salt length = 32
	binary.LittleEndian.PutUint32(buf[4+32:], 255) // nonce length = 255

	var p PasswordEncryptedData
	err := p.FromBytes(buf)
	assert.Error(t, err)
}

func TestFromBytesEmptyCiphertext(t *testing.T) {
	var src PasswordEncryptedData
	// leave Salt and Nonce zeroed, Ciphertext empty
	b := src.ToBytes()

	var dst PasswordEncryptedData
	err := dst.FromBytes(b)
	require.NoError(t, err)
	assert.Equal(t, src.Salt, dst.Salt)
	assert.Equal(t, src.Nonce, dst.Nonce)
	assert.Empty(t, dst.Ciphertext)
}
