package cipher

import (
	"crypto/rand"
	"encoding/hex"
)

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateCiphers(count int) ([]AESCipher, error) {
	ciphers := make([]AESCipher, 0, count)
	for range count {
		ciph, err := GenerateCipher()
		if err != nil {
			return nil, err
		}
		ciphers = append(ciphers, *ciph)
	}

	return ciphers, nil
}

func GenerateCipher() (*AESCipher, error) {
	key, err := generateRandom(32)
	if err != nil {
		return nil, err
	}

	return New(hex.EncodeToString(key))
}
