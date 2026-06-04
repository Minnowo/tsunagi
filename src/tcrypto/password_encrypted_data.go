package tcrypto

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrTruncated = fmt.Errorf("truncated value")
)

type PasswordEncryptedData struct {
	Salt       [32]byte
	Nonce      [chacha20poly1305.NonceSizeX]byte
	Ciphertext []byte
}

// ToBytes encodes into a []byte prefixing each field with it's size stored as LE Uint32.
func (p *PasswordEncryptedData) ToBytes() []byte {

	totalLen := 4 + len(p.Salt) +
		4 + len(p.Nonce) +
		4 + len(p.Ciphertext)

	out := make([]byte, 0, totalLen)

	var buf4 [4]byte

	// Salt
	binary.LittleEndian.PutUint32(buf4[:], uint32(len(p.Salt)))
	out = append(out, buf4[:]...)
	out = append(out, p.Salt[:]...)

	// Nonce
	binary.LittleEndian.PutUint32(buf4[:], uint32(len(p.Nonce)))
	out = append(out, buf4[:]...)
	out = append(out, p.Nonce[:]...)

	// Ciphertext
	binary.LittleEndian.PutUint32(buf4[:], uint32(len(p.Ciphertext)))
	out = append(out, buf4[:]...)
	out = append(out, p.Ciphertext...)

	return out
}

// FromBytes decodes and validates the []byte using the stored size prefixes.
// On error, the PasswordEncryptedData receiver may have had data written.
func (p *PasswordEncryptedData) FromBytes(data []byte) error {

	readLen := func(offset int) (uint32, error) {

		if offset+4 > len(data) {
			return 0, ErrTruncated
		}

		n := binary.LittleEndian.Uint32(data[offset : offset+4])

		return n, nil
	}

	readBytes := func(offset int, n uint32) ([]byte, error) {

		end := offset + int(n)

		if end > len(data) {
			return nil, ErrTruncated
		}

		return data[offset:end], nil
	}

	var n uint32
	var err error

	offset := 0

	// Salt
	{
		n, err = readLen(offset)
		offset += 4

		if err != nil {
			return err
		}

		if n > uint32(len(p.Salt)) {
			return fmt.Errorf("invalid salt size: %d", n)
		}

		saltBytes, err := readBytes(offset, n)
		offset += int(n)

		if err != nil {
			return err
		}

		copy(p.Salt[:], saltBytes)
	}

	{
		// Nonce
		n, err = readLen(offset)
		offset += 4

		if err != nil {
			return err
		}

		if n > uint32(len(p.Nonce)) {
			return fmt.Errorf("invalid nonce size: %d", n)
		}

		nonceBytes, err := readBytes(offset, n)
		offset += int(n)

		if err != nil {
			return err
		}

		copy(p.Nonce[:], nonceBytes)
	}

	{
		// Ciphertext
		n, err = readLen(offset)
		offset += 4

		if err != nil {
			return err
		}

		ct, err := readBytes(offset, n)
		offset += int(n)

		if err != nil {
			return err
		}

		p.Ciphertext = make([]byte, len(ct))
		copy(p.Ciphertext, ct)
	}

	return nil
}
