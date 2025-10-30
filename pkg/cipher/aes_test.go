package cipher

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestAESCipher(t *testing.T) {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	key, _ := hex.DecodeString(hexKey)
	ciph, _ := New(key)

	src := []byte("some source string")
	encryptedSrc := ciph.Encrypt(src)
	decryptedSrc := ciph.Decrypt(encryptedSrc)
	if strings.Compare(string(decryptedSrc), string(src)) != 0 {
		t.Fatalf("Wrong! Unexpected result!\n\tExpected: %v\n\tActual: %v\n", src, decryptedSrc)
	}
}
