package backups

import (
	"crypto/rand"
	"encoding/hex"

	"passman/pkg/cipher"
)

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateCiphers(count int) ([]cipher.AESCipher, error) {
	ciphers := make([]cipher.AESCipher, 0, count)
	for range count {
		ciph, err := GenerateCipher()
		if err != nil {
			return nil, err
		}
		ciphers = append(ciphers, *ciph)
	}

	return ciphers, nil
}

func GenerateCipher() (*cipher.AESCipher, error) {
	key, err := generateRandom(32)
	if err != nil {
		return nil, err
	}

	return cipher.New(hex.EncodeToString(key))
}
