package tcrypto

import (
	"crypto/rand"
	"errors"

	"github.com/flynn/noise"
	"golang.org/x/crypto/curve25519"
)

var (
	ErrBadHandshakeStep = errors.New("bad handshake step")
	ErrInvalidKeySize   = errors.New("invalid key size")
)

// NoiseCiphers hold the encrypt / decrypt tunnels for a finished noise handshake.
// This is mainly convenience just to make it easy to determine which tunnel should be used.
type NoiseCiphers struct {
	Enc *noise.CipherState
	Dec *noise.CipherState
}

const (
	// NoiseKeySize the number of bytes of the public key
	NoiseKeySize = curve25519.PointSize
)

var (
	noiseSuite = noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2s)
)

// GenerateNoiseKeypair generates a new public private keypair.
func GenerateNoiseKeypair() (noise.DHKey, error) {
	return noiseSuite.GenerateKeypair(rand.Reader)
}
