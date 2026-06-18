package cmd

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/flynn/noise"
	"github.com/urfave/cli/v3"
)

func handshakeXX() error {
	suite := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256)

	staticI, _ := suite.GenerateKeypair(rand.Reader)
	staticR, _ := suite.GenerateKeypair(rand.Reader)

	hsI, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite:   suite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		Initiator:     true,
		StaticKeypair: staticI,
	})
	hsR, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite:   suite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		StaticKeypair: staticR,
	})

	/*
		XX:
		-> e
		<- e, ee, s, es
		-> s, se
	*/

	// -> e
	// sender
	msg, _, _, _ := hsI.WriteMessage(nil, nil)
	// responder
	_, _, _, _ = hsR.ReadMessage(nil, msg)

	// <- e, ee, s, es
	// responder
	msg, _, _, _ = hsR.WriteMessage(nil, nil)
	// sender
	_, _, _, _ = hsI.ReadMessage(nil, msg)

	// -> s, se
	// sender
	msg, sEncSS, _, _ := hsI.WriteMessage(nil, nil)
	// responder
	_, rDecSS, _, _ := hsR.ReadMessage(nil, msg)

	cipherText, err := sEncSS.Encrypt(nil, nil, []byte("secret message"))

	if err != nil {
		fmt.Println("encrypt failed")
		return err
	}
	sEncSS.Rekey() // forward ratchet

	plainText, err := rDecSS.Decrypt(nil, nil, cipherText)

	if err != nil {
		fmt.Println("decrypt failed")
		return err
	}
	rDecSS.Rekey() // forward ratchet

	fmt.Println(string(plainText))

	cipherText, err = sEncSS.Encrypt(nil, nil, []byte("second secret message"))

	if err != nil {
		fmt.Println("encrypt failed")
		return err
	}

	plainText, err = rDecSS.Decrypt(nil, nil, cipherText)

	if err != nil {
		fmt.Println("decrypt failed")
		return err
	}
	fmt.Println(string(plainText))

	return nil
}
func handshakeIK() error {
	suite := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256)

	staticI, _ := suite.GenerateKeypair(rand.Reader)
	staticR, _ := suite.GenerateKeypair(rand.Reader)

	hsI, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite:   suite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		Initiator:     true,
		StaticKeypair: staticI,
	})
	hsR, _ := noise.NewHandshakeState(noise.Config{
		CipherSuite:   suite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		StaticKeypair: staticR,
	})

	/*
		XX:
		-> e
		<- e, ee, s, es
		-> s, se
	*/

	// -> e
	// sender
	msg, _, _, _ := hsI.WriteMessage(nil, nil)
	// responder
	_, _, _, _ = hsR.ReadMessage(nil, msg)

	// <- e, ee, s, es
	// responder
	msg, _, _, _ = hsR.WriteMessage(nil, nil)
	// sender
	_, _, _, _ = hsI.ReadMessage(nil, msg)

	// -> s, se
	// sender
	msg, sEncSS, _, _ := hsI.WriteMessage(nil, nil)
	// responder
	_, rDecSS, _, _ := hsR.ReadMessage(nil, msg)

	cipherText, err := sEncSS.Encrypt(nil, nil, []byte("secret message"))

	if err != nil {
		fmt.Println("encrypt failed")
		return err
	}
	sEncSS.Rekey() // forward ratchet

	plainText, err := rDecSS.Decrypt(nil, nil, cipherText)

	if err != nil {
		fmt.Println("decrypt failed")
		return err
	}
	rDecSS.Rekey() // forward ratchet

	fmt.Println(string(plainText))

	cipherText, err = sEncSS.Encrypt(nil, nil, []byte("second secret message"))

	if err != nil {
		fmt.Println("encrypt failed")
		return err
	}

	plainText, err = rDecSS.Decrypt(nil, nil, cipherText)

	if err != nil {
		fmt.Println("decrypt failed")
		return err
	}
	fmt.Println(string(plainText))

	return nil
}

func CmdSessionInit(ctx context.Context, c *cli.Command) error {

	return nil
}
