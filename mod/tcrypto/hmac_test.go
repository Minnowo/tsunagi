package tcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMACReturnsHash(t *testing.T) {
	key := make([]byte, MacKeySize)
	h, err := NewMAC(key)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

func TestNewMACProducesCorrectOutputSize(t *testing.T) {
	key := make([]byte, MacKeySize)
	h, err := NewMAC(key)
	require.NoError(t, err)
	assert.Equal(t, MacSize, h.Size())
}

func TestNewMACSameKeyAndInputProducesSameTag(t *testing.T) {
	key := make([]byte, MacKeySize)
	msg := []byte("hello")

	h1, err := NewMAC(key)
	require.NoError(t, err)
	h1.Write(msg)
	tag1 := h1.Sum(nil)

	h2, err := NewMAC(key)
	require.NoError(t, err)
	h2.Write(msg)
	tag2 := h2.Sum(nil)

	assert.Equal(t, tag1, tag2)
}

func TestNewMACDifferentKeysProduceDifferentTags(t *testing.T) {
	key1 := make([]byte, MacKeySize)
	key2 := make([]byte, MacKeySize)
	key2[0] = 0xff

	msg := []byte("hello")

	h1, err := NewMAC(key1)
	require.NoError(t, err)
	h1.Write(msg)

	h2, err := NewMAC(key2)
	require.NoError(t, err)
	h2.Write(msg)

	assert.NotEqual(t, h1.Sum(nil), h2.Sum(nil))
}

func TestNewMACDifferentMessagesProduceDifferentTags(t *testing.T) {
	key := make([]byte, MacKeySize)

	h1, err := NewMAC(key)
	require.NoError(t, err)
	h1.Write([]byte("hello"))

	h2, err := NewMAC(key)
	require.NoError(t, err)
	h2.Write([]byte("world"))

	assert.NotEqual(t, h1.Sum(nil), h2.Sum(nil))
}
