package secrets

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

func GenerateSecrets(count, size int) ([]cipher.AESCipher, error) {
	ciphers := make([]cipher.AESCipher, 0, count)
	for range count {
		key, err := generateRandom(size)
		if err != nil {
			return nil, err
		}
		ciph, err := cipher.New(key)
		if err != nil {
			return nil, err
		}
		ciphers = append(ciphers, *ciph)
	}

	return ciphers, nil
}
