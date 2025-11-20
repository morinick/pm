package accounts

import (
	"testing"

	"passman/pkg/cipher"

	"github.com/google/uuid"
)

func TestEntities(t *testing.T) {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	c, _ := cipher.New(hexKey)
	ciphs := []cipher.AESCipher{*c}

	correctTransfer := AccountDTO{
		QueryParams: QueryParams{
			UserID: uuid.New(),
		},
		Name:     "name",
		Login:    "login",
		Password: "password",
	}
	correctRecord := correctTransfer.ToAccount(uuid.New(), ciphs)
	if checkTransfer, err := correctRecord.ToAccountDTO(ciphs); err != nil {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", nil, err)
	} else if checkTransfer != correctTransfer {
		t.Errorf("Wrong! Unexpected convertation result!\n\tExpected: %v\n\tActual: %v\n", correctTransfer, checkTransfer)
	}

	notHexRecord := correctRecord
	notHexRecord.Payload = "not in hex"
	errMsg := "payload is not in hex encoding"
	if _, err := notHexRecord.ToAccountDTO(ciphs); err != nil {
		if err.Error() != errMsg {
			t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", errMsg, err.Error())
		}
	} else {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: nil\n", errMsg)
	}

	withoutSeparatorRecord := correctRecord
	withoutSeparatorRecord.Payload = hexKey
	errMsg = "separator not found"
	if _, err := withoutSeparatorRecord.ToAccountDTO(ciphs); err != nil {
		if err.Error() != errMsg {
			t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v\n", errMsg, err.Error())
		}
	} else {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: nil\n", errMsg)
	}
}
