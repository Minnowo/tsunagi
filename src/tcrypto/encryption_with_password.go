package tcrypto

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const wpArgon2Iter = 3
const wpArgon2Memory = 64 * 1024
const wpArgon2Parallel = 4

func EncryptWithPassword(data []byte, password []byte) (*PasswordEncryptedData, error) {

	var encdata PasswordEncryptedData

	rand.Read(encdata.Salt[:])

	key := argon2.IDKey(password, encdata.Salt[:],
		wpArgon2Iter, wpArgon2Memory, wpArgon2Parallel, chacha20poly1305.KeySize,
	)

	// wipe the key
	defer rand.Read(key)

	aead, err := chacha20poly1305.NewX(key)

	if err != nil {
		return nil, err
	}

	rand.Read(encdata.Nonce[:])

	encdata.Ciphertext = aead.Seal(nil, encdata.Nonce[:], data, nil)

	return &encdata, nil
}

func DecryptWithPassword(enc *PasswordEncryptedData, password []byte) ([]byte, error) {

	key := argon2.IDKey(password, enc.Salt[:],
		wpArgon2Iter, wpArgon2Memory, wpArgon2Parallel, chacha20poly1305.KeySize,
	)

	// wipe the key
	defer rand.Read(key)

	aead, err := chacha20poly1305.NewX(key)

	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, enc.Nonce[:], enc.Ciphertext, nil)

	if err != nil {
		// wrong password or corrupted data
		return nil, err
	}

	return plaintext, nil
}
