package creds

import (
	"encoding/hex"
	"fmt"
	"passman/pkg/cipher"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	Name     string
	Login    string
	Password string
}

func (s *Service) ToServiceDAO(owner uuid.UUID, ciphrs []cipher.AESCipher) ServiceDAO {
	keyIndx := time.Now().Nanosecond() % len(ciphrs)

	src := make([]byte, 0, len(s.Login)+len(s.Password)+7)
	src = fmt.Appendf(src, "'%s'-:-'%s'", s.Login, s.Password)

	encryptedSrc := ciphrs[keyIndx].Encrypt(src)

	return ServiceDAO{
		ID:      uuid.New(),
		Owner:   owner,
		Name:    s.Name,
		Key:     keyIndx,
		Payload: hex.EncodeToString(encryptedSrc),
	}
}

type ServiceDAO struct {
	ID      uuid.UUID
	Owner   uuid.UUID
	Name    string
	Key     int
	Payload string
}

func (sdao *ServiceDAO) ToService(ciphrs []cipher.AESCipher) (Service, error) {
	src, err := hex.DecodeString(sdao.Payload)
	if err != nil {
		return Service{}, fmt.Errorf("payload is not in hex encoding")
	}

	decrypted := ciphrs[sdao.Key].Decrypt(src)

	splited := strings.Split(string(decrypted), "'-:-'")

	if len(splited) != 2 {
		return Service{}, fmt.Errorf("separator not found")
	}

	s := Service{Name: sdao.Name}
	s.Login = splited[0][1:]
	s.Password = splited[1][:len(splited[1])-1]

	return s, nil
}
