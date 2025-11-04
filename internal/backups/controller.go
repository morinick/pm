package backups

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"passman/pkg/cipher"
)

type Controller struct {
	dbFileName string
	bakupDir   string
	Ciphers    []cipher.AESCipher
	DecryptKey string
}

func New(dbFileName, bakupDir, decryptKey string) (*Controller, error) {
	c := &Controller{
		dbFileName: dbFileName,
		bakupDir:   bakupDir,
		Ciphers:    nil,
		DecryptKey: decryptKey,
	}

	// Check existing backup directory
	// And if exist try to load database from saved_database.bak
	if bDir, err := os.ReadDir(bakupDir); err != nil {
		return nil, fmt.Errorf("failed to open backup directory: %w", err)
	} else {
		for _, file := range bDir {
			if file.Name() == "saved_database.bak" {
				if err := c.LoadDatabase("saved_database.bak"); err != nil {
					return nil, err
				}
			}
		}
	}

	return c, nil
}

func (c *Controller) LoadDatabase(fileName string) error {
	if len(c.DecryptKey) == 0 {
		return fmt.Errorf("empty decryption key")
	}

	backupData, err := parseDataFromFile(c.bakupDir + "/" + fileName)
	if err != nil {
		return err
	}

	binaryKey, err := hex.DecodeString(c.DecryptKey)
	if err != nil {
		return fmt.Errorf("failed decoding key from hex: %w", err)
	}

	ciph, err := cipher.New(binaryKey)
	if err != nil {
		return err
	}

	data := ciph.Decrypt(backupData)

	keys, dbData, err := splitBackupData(data)
	if err != nil {
		return err
	}

	if err := c.addCiphers(keys); err != nil {
		return fmt.Errorf("failed adding ciphers: %w", err)
	}

	dbFile, err := os.OpenFile(c.dbFileName, os.O_RDWR, 0o664)
	if err != nil {
		return fmt.Errorf("failed opening db file: %w", err)
	}
	defer dbFile.Close()

	if _, err := dbFile.Write(dbData); err != nil {
		return fmt.Errorf("failed writing data to db: %w", err)
	}

	return nil
}

func (c *Controller) SaveDatabase(fileName string) error {
	fileData, err := parseDataFromFile(c.dbFileName)
	if err != nil {
		return err
	}

	ciphers := c.Ciphers
	lenKeys := 65 * len(ciphers) // 64 bytes for key and 1 byte for '\n'
	keys := make([]byte, 0, lenKeys)
	for i := range ciphers {
		hexKey := hex.EncodeToString(ciphers[i].Key()) + "\n"
		keys = append(keys, []byte(hexKey)...)
	}

	data := make([]byte, 0, len(fileData)+lenKeys+10)
	data = append(data, keys...)
	data = append(data, []byte("---data---")...)
	data = append(data, fileData...)

	encryptedData, key, err := encryptDatabase(data)
	if err != nil {
		return fmt.Errorf("failed encrypting database: %w", err)
	}

	saveFile, err := os.Create(c.bakupDir + "/" + fileName)
	if err != nil {
		return fmt.Errorf("failed creating backup file: %w", err)
	}
	defer saveFile.Close()

	if _, err := saveFile.Write(encryptedData); err != nil {
		return fmt.Errorf("failed saving data to backup file: %w", err)
	}

	c.DecryptKey = hex.EncodeToString(key)

	return nil
}

func (c *Controller) addCiphers(keys []byte) error {
	stringKeys := strings.Split(string(keys), "\n")

	for _, key := range stringKeys {
		binaryKey, err := hex.DecodeString(key)
		if err != nil {
			return fmt.Errorf("failed decoding internal key from hex: %w", err)
		}

		ciph, err := cipher.New(binaryKey)
		if err != nil {
			return fmt.Errorf("failed creating cipher: %w", err)
		}
		c.Ciphers = append(c.Ciphers, *ciph)
	}
	return nil
}

func encryptDatabase(data []byte) ([]byte, []byte, error) {
	ciph, err := GenerateCipher(32)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating cipher: %w", err)
	}

	encryptedData := ciph.Encrypt(data)

	return encryptedData, ciph.Key(), nil
}

func parseDataFromFile(fileName string) ([]byte, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed reading data from file: %w", err)
	}
	return data, nil
}

func splitBackupData(data []byte) ([]byte, []byte, error) {
	peaces := strings.Split(string(data), "---data---")
	if len(peaces) != 2 {
		return nil, nil, fmt.Errorf("separator not found")
	}

	return []byte(peaces[0][:len(peaces[0])-1]), []byte(peaces[1]), nil
}
