package tcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNoiseKeypair(t *testing.T) {
	kp, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	assert.NotEmpty(t, kp.Public)
	assert.NotEmpty(t, kp.Private)
}

func TestGenerateNoiseKeypairUnique(t *testing.T) {
	kp1, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	kp2, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	assert.NotEqual(t, kp1.Public, kp2.Public)
}

func TestNewSenderHandshakeState(t *testing.T) {
	responder, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	sender, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	state, err := NewSenderHandshakeState(responder.Public, sender)
	require.NoError(t, err)
	assert.NotNil(t, state)
}

func TestNewResponderHandshakeState(t *testing.T) {
	kp, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	state, err := NewResponderHandshakeState(kp)
	require.NoError(t, err)
	assert.NotNil(t, state)
}

func doHandshake(t *testing.T) (NoiseCiphers, NoiseCiphers) {
	t.Helper()

	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	// Step 1: sender -> responder
	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)
	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	// Step 2: responder -> sender
	msg2, err := ResponderHandshakeStep2(responderState)
	require.NoError(t, err)
	err = SenderHandshakeStep2(msg2, senderState)
	require.NoError(t, err)

	// Step 3: sender -> responder (completes handshake)
	msg3, senderCiphers, err := SenderHandshakeStep3(senderState)
	require.NoError(t, err)
	responderCiphers, err := ResponderHandshakeStep3(msg3, responderState)
	require.NoError(t, err)

	return senderCiphers, responderCiphers
}

func TestFullHandshake(t *testing.T) {
	senderCiphers, responderCiphers := doHandshake(t)
	assert.NotNil(t, senderCiphers.Enc)
	assert.NotNil(t, senderCiphers.Dec)
	assert.NotNil(t, responderCiphers.Enc)
	assert.NotNil(t, responderCiphers.Dec)
}

func TestHandshakeCiphersCanEncryptDecrypt(t *testing.T) {
	senderCiphers, responderCiphers := doHandshake(t)

	plaintext := []byte("hello noise")
	ciphertext, err := senderCiphers.Enc.Encrypt(nil, nil, plaintext)
	require.NoError(t, err)

	got, err := responderCiphers.Dec.Decrypt(nil, nil, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestHandshakeStep1BadMessage(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	err = ResponderHandshakeStep1([]byte("bad message"), responderState)
	assert.Error(t, err)
}

func TestSenderHandshakeStep1ReturnsMsgOnly(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	state, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)

	msg, err := SenderHandshakeStep1(state)
	require.NoError(t, err)
	assert.NotEmpty(t, msg)
}

// --- out-of-order step failures ---

func TestSenderSkipsStep1GoesToStep2(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)

	// Sender state has not sent step 1 yet; ReadMessage is wrong at this point.
	err = SenderHandshakeStep2([]byte("anything"), senderState)
	assert.Error(t, err)
}

func TestSenderStep1CalledTwice(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)

	_, err = SenderHandshakeStep1(senderState)
	require.NoError(t, err)

	// WriteMessage is wrong now; state expects ReadMessage.
	_, err = SenderHandshakeStep1(senderState)
	assert.Error(t, err)
}

func TestSenderSkipsStep2GoesToStep3(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)

	_, err = SenderHandshakeStep1(senderState)
	require.NoError(t, err)

	// State expects ReadMessage for step 2; WriteMessage is wrong.
	_, _, err = SenderHandshakeStep3(senderState)
	assert.Error(t, err)
}

func TestResponderSkipsStep1GoesToStep2(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	// Non-initiator state starts expecting ReadMessage; WriteMessage is wrong.
	_, err = ResponderHandshakeStep2(responderState)
	assert.Error(t, err)
}

func TestResponderStep1CalledTwice(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)

	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	// State now expects WriteMessage; ReadMessage is wrong.
	err = ResponderHandshakeStep1(msg1, responderState)
	assert.Error(t, err)
}

// --- corrupted / replayed messages ---

func TestSenderStep2CorruptMessage(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)
	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	err = SenderHandshakeStep2([]byte("corrupt"), senderState)
	assert.Error(t, err)
}

func TestResponderStep3CorruptMessage(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)
	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	msg2, err := ResponderHandshakeStep2(responderState)
	require.NoError(t, err)
	err = SenderHandshakeStep2(msg2, senderState)
	require.NoError(t, err)

	_, _, err = SenderHandshakeStep3(senderState)
	require.NoError(t, err)

	_, err = ResponderHandshakeStep3([]byte("corrupt"), responderState)
	assert.Error(t, err)
}

func TestStep1MessageReplayedAsStep2(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)
	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	// Feed msg1 back to the sender where msg2 is expected.
	err = SenderHandshakeStep2(msg1, senderState)
	assert.Error(t, err)
}

func TestStep2MessageReplayedAsStep3(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	senderState, err := NewSenderHandshakeState(responderKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)
	err = ResponderHandshakeStep1(msg1, responderState)
	require.NoError(t, err)

	msg2, err := ResponderHandshakeStep2(responderState)
	require.NoError(t, err)
	err = SenderHandshakeStep2(msg2, senderState)
	require.NoError(t, err)

	_, _, err = SenderHandshakeStep3(senderState)
	require.NoError(t, err)

	// Feed msg2 to the responder where msg3 is expected.
	_, err = ResponderHandshakeStep3(msg2, responderState)
	assert.Error(t, err)
}

// --- wrong keypair ---

func TestWrongResponderKeyFailsAtStep3(t *testing.T) {
	responderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	wrongKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)
	senderKP, err := GenerateNoiseKeypair()
	require.NoError(t, err)

	// Sender believes the responder has wrongKP's public key.
	senderState, err := NewSenderHandshakeState(wrongKP.Public, senderKP)
	require.NoError(t, err)
	responderState, err := NewResponderHandshakeState(responderKP)
	require.NoError(t, err)

	msg1, err := SenderHandshakeStep1(senderState)
	require.NoError(t, err)

	// should fail because the sender used the wrong pub key
	err = ResponderHandshakeStep1(msg1, responderState)
	require.Error(t, err)
}

// --- post-handshake cipher misuse ---

func TestCrossSessionDecryptFails(t *testing.T) {
	// Two independent sessions produce incompatible keys.
	sessionA, _ := doHandshake(t)
	_, sessionB := doHandshake(t)

	ciphertext, err := sessionA.Enc.Encrypt(nil, nil, []byte("hello"))
	require.NoError(t, err)

	_, err = sessionB.Dec.Decrypt(nil, nil, ciphertext)
	assert.Error(t, err)
}

func TestWrongDirectionDecryptFails(t *testing.T) {
	senderCiphers, _ := doHandshake(t)

	// senderCiphers.Enc is for sender→responder; senderCiphers.Dec is for the
	// opposite direction. Encrypting with Enc and decrypting with Dec should fail.
	ciphertext, err := senderCiphers.Enc.Encrypt(nil, nil, []byte("hello"))
	require.NoError(t, err)

	_, err = senderCiphers.Dec.Decrypt(nil, nil, ciphertext)
	assert.Error(t, err)
}
