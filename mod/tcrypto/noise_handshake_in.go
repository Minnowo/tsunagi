package tcrypto

import (
	"crypto/rand"

	"github.com/flynn/noise"
)

// NewSenderHandshakeIN creates a new handshake state for the sender in the IN pattern.
func NewSenderHandshakeIN(staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeIN,
		Initiator:     true,
		StaticKeypair: staticKeypair,
	})
}

// NewResponderHandshakeIN creates a new handshake state for the responder in the IN pattern.
func NewResponderHandshakeIN() (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite: noiseSuite,
		Random:      rand.Reader,
		Pattern:     noise.HandshakeIN,
	})
}

// SenderHandshakeINStep1 does step one for the sender in the IN handshake pattern.
func SenderHandshakeINStep1(state *noise.HandshakeState) ([]byte, error) {

	// -> e, s
	msg, x, y, err := state.WriteMessage(nil, nil)

	if err == nil && (x != nil || y != nil) {
		return nil, ErrBadHandshakeStep
	}

	return msg, err
}

// ResponderHandshakeINStep1 does step one for the responder in the IN handshake pattern.
func ResponderHandshakeINStep1(msg []byte, state *noise.HandshakeState) ([]byte, NoiseCiphers, error) {

	// -> e, es
	if _, x, y, err := state.ReadMessage(nil, msg); err != nil {
		return nil, NoiseCiphers{}, err
	} else if x != nil || y != nil {
		return nil, NoiseCiphers{}, ErrBadHandshakeStep
	}

	// <- e, ee, se
	msg, dec, enc, err := state.WriteMessage(nil, nil)

	if err != nil {
		return nil, NoiseCiphers{}, err
	}

	if enc == nil || dec == nil {
		return nil, NoiseCiphers{}, ErrBadHandshakeStep
	}

	return msg, NoiseCiphers{
		Enc: enc,
		Dec: dec,
	}, nil
}

// SenderHandshakeINStep2 does step two for the sender in the IN handshake pattern.
func SenderHandshakeINStep2(msg []byte, state *noise.HandshakeState) (NoiseCiphers, error) {

	// <- e, ee, se
	_, enc, dec, err := state.ReadMessage(nil, msg)

	if err != nil {
		return NoiseCiphers{}, err
	}

	if enc == nil || dec == nil {
		return NoiseCiphers{}, ErrBadHandshakeStep
	}

	return NoiseCiphers{
		Enc: enc,
		Dec: dec,
	}, nil
}
