package creds

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"passman/pkg/cipher"

	"github.com/google/uuid"
)

type QueryParams struct {
	UserID      uuid.UUID
	ServiceName string
}

type CredsRecordTransfer struct {
	QueryParams
	Name     string
	Login    string
	Password string
}

func (crt *CredsRecordTransfer) ToCredsRecord(serviceID uuid.UUID, ciphers []cipher.AESCipher) CredsRecord {
	keyIndx := time.Now().Nanosecond() % len(ciphers)

	src := make([]byte, 0, len(crt.Login)+len(crt.Password)+7)
	src = fmt.Appendf(src, "'%s'-:-'%s'", crt.Login, crt.Password)

	encryptedSrc := ciphers[keyIndx].Encrypt(src)

	return CredsRecord{
		ID:        uuid.New(),
		UserID:    crt.UserID,
		ServiceID: serviceID,
		Name:      crt.Name,
		Secret:    int64(keyIndx),
		Payload:   hex.EncodeToString(encryptedSrc),
	}
}

type CredsRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ServiceID uuid.UUID
	Name      string
	Secret    int64
	Payload   string
}

func (cr *CredsRecord) ToCredsRecordTransfer(ciphers []cipher.AESCipher) (CredsRecordTransfer, error) {
	src, err := hex.DecodeString(cr.Payload)
	if err != nil {
		return CredsRecordTransfer{}, fmt.Errorf("payload is not in hex encoding")
	}

	decryptedSrc := ciphers[cr.Secret].Decrypt(src)

	splited := strings.Split(string(decryptedSrc), "'-:-'")

	if len(splited) != 2 {
		return CredsRecordTransfer{}, fmt.Errorf("separator not found")
	}

	return CredsRecordTransfer{
		QueryParams: QueryParams{UserID: cr.UserID},
		Name:        cr.Name,
		Login:       splited[0][1:],
		Password:    splited[1][:len(splited[1])-1],
	}, nil
}
