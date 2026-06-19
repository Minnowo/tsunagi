package tcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doHandshakeIN(t *testing.T) (NoiseCiphers, NoiseCiphers) {
	t.Helper()

	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeIN(senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeIN()
	require.NoError(t, err)

	// Step 1: sender -> responder (-> e, s)
	msg1, err := SenderHandshakeINStep1(senderState)
	require.NoError(t, err)

	// Responder reads msg1 and replies (-> e, s then <- e, ee, se)
	msg2, responderCiphers, err := ResponderHandshakeINStep1(msg1, responderState)
	require.NoError(t, err)

	// Sender reads msg2 and derives ciphers
	senderCiphers, err := SenderHandshakeINStep2(msg2, senderState)
	require.NoError(t, err)

	return senderCiphers, responderCiphers
}

func TestINFullHandshake(t *testing.T) {
	senderCiphers, responderCiphers := doHandshakeIN(t)
	assert.NotNil(t, senderCiphers.Enc)
	assert.NotNil(t, senderCiphers.Dec)
	assert.NotNil(t, responderCiphers.Enc)
	assert.NotNil(t, responderCiphers.Dec)
}

func TestINHandshakeCiphersCanEncryptDecrypt(t *testing.T) {
	senderCiphers, responderCiphers := doHandshakeIN(t)

	plaintext := []byte("hello noise IN")
	ciphertext, err := senderCiphers.Enc.Encrypt(nil, nil, plaintext)
	require.NoError(t, err)

	got, err := responderCiphers.Dec.Decrypt(nil, nil, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestINResponderToSenderEncryptDecrypt(t *testing.T) {
	senderCiphers, responderCiphers := doHandshakeIN(t)

	plaintext := []byte("hello from responder")
	ciphertext, err := responderCiphers.Enc.Encrypt(nil, nil, plaintext)
	require.NoError(t, err)

	got, err := senderCiphers.Dec.Decrypt(nil, nil, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestINSenderStep1ReturnsMsgOnly(t *testing.T) {
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	state, err := NewSenderHandshakeIN(senderKP)
	require.NoError(t, err)

	msg, err := SenderHandshakeINStep1(state)
	require.NoError(t, err)
	assert.NotEmpty(t, msg)
}

func TestINResponderStep1BadMessage(t *testing.T) {
	responderState, err := NewResponderHandshakeIN()
	require.NoError(t, err)

	_, _, err = ResponderHandshakeINStep1([]byte("bad message"), responderState)
	assert.Error(t, err)
}

func TestINSenderStep2BadMessage(t *testing.T) {
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeIN(senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeIN()
	require.NoError(t, err)

	msg1, err := SenderHandshakeINStep1(senderState)
	require.NoError(t, err)
	_, _, err = ResponderHandshakeINStep1(msg1, responderState)
	require.NoError(t, err)

	_, err = SenderHandshakeINStep2([]byte("corrupt"), senderState)
	assert.Error(t, err)
}

func TestINCrossSessionDecryptFails(t *testing.T) {
	sessionA, _ := doHandshakeIN(t)
	_, sessionB := doHandshakeIN(t)

	ciphertext, err := sessionA.Enc.Encrypt(nil, nil, []byte("hello"))
	require.NoError(t, err)

	_, err = sessionB.Dec.Decrypt(nil, nil, ciphertext)
	assert.Error(t, err)
}
