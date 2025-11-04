package backups

import (
	"crypto/rand"

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

func GenerateCiphers(count, size int) ([]cipher.AESCipher, error) {
	ciphers := make([]cipher.AESCipher, 0, count)
	for range count {
		ciph, err := GenerateCipher(size)
		if err != nil {
			return nil, err
		}
		ciphers = append(ciphers, *ciph)
	}

	return ciphers, nil
}

func GenerateCipher(size int) (*cipher.AESCipher, error) {
	key, err := generateRandom(size)
	if err != nil {
		return nil, err
	}

	return cipher.New(key)
}
