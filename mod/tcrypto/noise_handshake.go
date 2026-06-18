package tcrypto

import (
	"crypto/rand"
	"fmt"

	"github.com/flynn/noise"
)

var (
	noiseSuite = noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2s)
)

var (
	ErrBadHandshakeStep = fmt.Errorf("bad handshake step")
)

// GenerateNoiseKeypair generates a new public private keypair.
func GenerateNoiseKeypair() (noise.DHKey, error) {
	return noiseSuite.GenerateKeypair(rand.Reader)
}

// NewSenderHandshakeState creates a new handshake state for the sender in the XK handshake pattern.
func NewSenderHandshakeState(responderPubStaticKey []byte, staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXK,
		Initiator:     true,
		StaticKeypair: staticKeypair,
		PeerStatic:    responderPubStaticKey,
	})
}

// NewResponderHandshakeState creates a new handshake state for the responder in the XK handshake pattern.
func NewResponderHandshakeState(staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXK,
		StaticKeypair: staticKeypair,
	})
}

func NewSenderAuthHandshakeState(staticKeypair noise.DHKey) (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeIN,
		StaticKeypair: staticKeypair,
	})
}

func NewResponderAuthHandshakeState() (*noise.HandshakeState, error) {

	return noise.NewHandshakeState(noise.Config{
		CipherSuite:   noiseSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeIN,
	})
}

// NoiseCiphers hold the encrypt / decrypt tunnels for a finished noise handshake.
// This is mainly convenience just to make it easy to determine which tunnel should be used.
type NoiseCiphers struct {
	Enc *noise.CipherState
	Dec *noise.CipherState
}

// SenderHandshakeStep1 does step one for the sender in the XK handshake pattern.
func SenderHandshakeStep1(state *noise.HandshakeState) ([]byte, error) {

	// -> e, es
	msg, x, y, err := state.WriteMessage(nil, nil)

	if err == nil && (x != nil || y != nil) {
		return nil, ErrBadHandshakeStep
	}

	return msg, err
}

// ResponderHandshakeStep1 does step one for the responder in the XK handshake pattern.
func ResponderHandshakeStep1(msg []byte, state *noise.HandshakeState) error {

	// -> e, es
	_, x, y, err := state.ReadMessage(nil, msg)

	if err == nil && (x != nil || y != nil) {
		return ErrBadHandshakeStep
	}

	return err
}

// SenderHandshakeStep2 does step two for the sender in the XK handshake pattern.
func SenderHandshakeStep2(msg []byte, state *noise.HandshakeState) error {

	// <- e, ee
	_, x, y, err := state.ReadMessage(nil, msg)

	if err == nil && (x != nil || y != nil) {
		return ErrBadHandshakeStep
	}

	return err
}

// ResponderHandshakeStep2 does step two for the responder in the XK handshake pattern.
func ResponderHandshakeStep2(state *noise.HandshakeState) ([]byte, error) {

	// <- e, ee
	msg, x, y, err := state.WriteMessage(nil, nil)

	if err == nil && (x != nil || y != nil) {
		return nil, ErrBadHandshakeStep
	}

	return msg, err
}

// SenderHandshakeStep3 finishes the handshake for the sender in the XK handshake pattern.
func SenderHandshakeStep3(state *noise.HandshakeState) ([]byte, NoiseCiphers, error) {

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

// ResponderHandshakeStep3 finishes the handshake for the responder in the XK handshake pattern.
func ResponderHandshakeStep3(msg []byte, state *noise.HandshakeState) (NoiseCiphers, error) {

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
