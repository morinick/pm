package creds

import (
	"encoding/hex"
	"passman/pkg/cipher"
	"testing"

	"github.com/google/uuid"
)

func TestToService(t *testing.T) {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	key, _ := hex.DecodeString(hexKey)
	c, _ := cipher.New(key)
	ciphs := []cipher.AESCipher{*c}

	correctService := Service{Name: "name", Login: "login", Password: "password"}
	correctDAO := correctService.ToServiceDAO(uuid.New(), ciphs)
	if checkService, err := correctDAO.ToService(ciphs); err != nil {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", nil, err)
	} else if checkService != correctService {
		t.Errorf("Wrong! Unexpected convertation result!\n\tExpected: %v\n\tActual: %v\n", correctService, checkService)
	}

	notHexDAO := correctDAO
	notHexDAO.Payload = "not in hex"
	errMsg := "payload is not in hex encoding"
	if _, err := notHexDAO.ToService(ciphs); err != nil {
		if err.Error() != errMsg {
			t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", errMsg, err.Error())
		}
	} else {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: nil\n", errMsg)
	}

	withoutSeparatorDAO := correctDAO
	withoutSeparatorDAO.Payload = hexKey
	errMsg = "separator not found"
	if _, err := withoutSeparatorDAO.ToService(ciphs); err != nil {
		if err.Error() != errMsg {
			t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", errMsg, err.Error())
		}
	} else {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: nil\n", errMsg)
	}
}
