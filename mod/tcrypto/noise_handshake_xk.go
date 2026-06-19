package tcrypto

import (
	"crypto/rand"

	"github.com/flynn/noise"
)

// NewSenderHandshakeXK creates a new handshake state for the sender in the XK handshake pattern.
func NewSenderHandshakeXK(responderPubStaticKey []byte, staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXK,
		Initiator:     true,
		StaticKeypair: staticKeypair,
		PeerStatic:    responderPubStaticKey,
	})
}

// NewResponderHandshakeXK creates a new handshake state for the responder in the XK handshake pattern.
func NewResponderHandshakeXK(staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXK,
		StaticKeypair: staticKeypair,
	})
}

// SenderHandshakeXKStep1 does step one for the sender in the XK handshake pattern.
func SenderHandshakeXKStep1(state *noise.HandshakeState) ([]byte, error) {

	// -> e, es
	msg, x, y, err := state.WriteMessage(nil, nil)

	if err == nil && (x != nil || y != nil) {
		return nil, ErrBadHandshakeStep
	}

	return msg, err
}

// ResponderHandshakeXKStep1 does the first read and write step for the responder in the XK handshake pattern.
func ResponderHandshakeXKStep1(msg []byte, state *noise.HandshakeState) ([]byte, error) {

	// -> e, es
	if _, x, y, err := state.ReadMessage(nil, msg); err != nil {
		return nil, err
	} else if x != nil || y != nil {
		return nil, ErrBadHandshakeStep
	}

	// <- e, ee
	msg, x, y, err := state.WriteMessage(nil, nil)

	if err == nil && (x != nil || y != nil) {
		return nil, ErrBadHandshakeStep
	}

	return msg, err
}

// SenderHandshakeXKStep2 does step two for the sender in the XK handshake pattern.
func SenderHandshakeXKStep2(msg []byte, state *noise.HandshakeState) error {

	// <- e, ee
	_, x, y, err := state.ReadMessage(nil, msg)

	if err == nil && (x != nil || y != nil) {
		return ErrBadHandshakeStep
	}

	return err
}

// ResponderHandshakeXKStep2 does step two for the responder in the XK handshake pattern.
func ResponderHandshakeXKStep2(msg []byte, state *noise.HandshakeState) (NoiseCiphers, error) {

	// -> s, se
	_, dec, enc, err := state.ReadMessage(nil, msg)

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

// SenderHandshakeXKStep3 finishes the handshake for the sender in the XK handshake pattern.
func SenderHandshakeXKStep3(state *noise.HandshakeState) ([]byte, NoiseCiphers, error) {

	// -> s, se
	msg, enc, dec, err := state.WriteMessage(nil, nil)

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
