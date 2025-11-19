package backups

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"passman/pkg/archivator"
	"passman/pkg/cipher"
)

var ErrEmptyKey = errors.New("empty key")

type Controller struct {
	backupDir        string
	dbURL            string
	dbBackupName     string
	assetsDir        string
	assetsBackupName string
	Ciphers          []cipher.AESCipher
	Key              string
}

func New(dbURL, backupDir, key string) *Controller {
	return &Controller{
		backupDir:        backupDir,
		dbURL:            dbURL,
		dbBackupName:     filepath.Join(backupDir, "db.bak"),
		assetsDir:        "/assets",
		assetsBackupName: filepath.Join(backupDir, "assets.zip"),
		Ciphers:          nil,
		Key:              key,
	}
}

func (ctrl *Controller) LoadBackup() error {
	// Check existing components
	if err := ctrl.checkComponents(); err != nil {
		return err
	}

	// Load db and ciphers
	encryptedDB, err := os.ReadFile(ctrl.dbBackupName)
	if err != nil {
		return fmt.Errorf("failed reading db backup: %w", err)
	}

	decryptedDB, err := ctrl.decryptDB(encryptedDB)
	if err != nil {
		return fmt.Errorf("failed decrypting %s: %w", ctrl.dbBackupName, err)
	}

	ciphs, dbData, err := ctrl.parseDecryptedDB(decryptedDB)
	if err != nil {
		return fmt.Errorf("failed parsing decrypted db: %w", err)
	}

	ctrl.Ciphers = ciphs

	if err := ctrl.loadDB(dbData); err != nil {
		return fmt.Errorf("failed loading db: %w", err)
	}

	// Load assets
	if err := ctrl.deleteStandartAssets(); err != nil {
		return fmt.Errorf("failed deliting standart assets: %w", err)
	}

	if err := archivator.Decompress(ctrl.assetsBackupName, ctrl.assetsDir); err != nil {
		return fmt.Errorf("failed decompressing assets: %w", err)
	}

	return nil
}

func (ctrl *Controller) SaveBackup() error {
	ciph, err := ctrl.getCipher()
	if err != nil {
		return fmt.Errorf("failed getting cipher: %w", err)
	}

	preparedDB, err := ctrl.prepareDB()
	if err != nil {
		return fmt.Errorf("failed preparing db: %w", err)
	}

	encryptedDB, err := ctrl.encryptDB(preparedDB, ciph)
	if err != nil {
		return fmt.Errorf("failed encrypting db: %w", err)
	}

	if err := ctrl.saveDB(encryptedDB); err != nil {
		return fmt.Errorf("failed saving db to backup: %w", err)
	}

	if err := archivator.Compress(ctrl.assetsDir, ctrl.assetsBackupName); err != nil {
		return fmt.Errorf("failed saving assets: %w", err)
	}

	return nil
}

func (ctrl *Controller) getCipher() (*cipher.AESCipher, error) {
	if len(ctrl.Key) == 0 {
		ciph, err := GenerateCipher()
		ctrl.Key = ciph.Key()
		return ciph, err
	}
	return cipher.New(ctrl.Key)
}

func (ctrl *Controller) checkComponents() error {
	if len(ctrl.Key) == 0 {
		return ErrEmptyKey
	}

	files, err := os.ReadDir(ctrl.backupDir)
	if err != nil {
		return fmt.Errorf("failed opening backup directory: %w", err)
	}

	countExisted := 0
	_, dbfile := filepath.Split(ctrl.dbBackupName)
	_, assetsFile := filepath.Split(ctrl.assetsBackupName)
	for _, file := range files {
		if file.Name() == dbfile || file.Name() == assetsFile {
			countExisted++
		}
	}

	if countExisted != 2 {
		return fmt.Errorf("not all components are found (%s or %s)", ctrl.dbBackupName, ctrl.assetsBackupName)
	}

	return nil
}

// prepareDB parses ctrl.Ciphers and db data into one buffer
func (ctrl *Controller) prepareDB() (*bytes.Buffer, error) {
	if ctrl.Ciphers == nil {
		return nil, fmt.Errorf("ciphers are nil")
	}

	dbData, err := os.ReadFile(ctrl.dbURL)
	if err != nil {
		return nil, err
	}

	lenKeys := 65 * len(ctrl.Ciphers)              // 64 bytes for key and 1 byte for '\n'
	buf := make([]byte, 0, len(dbData)+lenKeys+10) // 10 bytes for separator
	res := bytes.NewBuffer(buf)

	// Add keys for ciphers
	for _, ciph := range ctrl.Ciphers {
		if _, err := res.WriteString(ciph.Key() + "\n"); err != nil {
			return nil, err
		}
	}

	// Add separator
	if _, err := res.WriteString("---data---"); err != nil {
		return nil, err
	}

	// Add db data
	if _, err := res.Write(dbData); err != nil {
		return nil, err
	}

	return res, nil
}

func (ctrl *Controller) parseDecryptedDB(data []byte) ([]cipher.AESCipher, []byte, error) {
	ciphersKeys, dbData, err := splitBackupData(data)
	if err != nil {
		return nil, nil, err
	}

	ciphs, err := parseCiphers(ciphersKeys)
	if err != nil {
		return nil, nil, err
	}

	return ciphs, dbData, nil
}

func (ctrl *Controller) encryptDB(data *bytes.Buffer, ciph *cipher.AESCipher) (*bytes.Buffer, error) {
	encryptedData := ciph.Encrypt(data.Bytes())
	data.Reset()
	_, err := data.Write(encryptedData)
	return data, err
}

func (ctrl *Controller) decryptDB(data []byte) ([]byte, error) {
	ciph, err := ctrl.getCipher()
	if err != nil {
		return nil, err
	}
	decryptedDB := ciph.Decrypt(data)
	return decryptedDB, nil
}

func (ctrl *Controller) saveDB(data *bytes.Buffer) error {
	saveFile, err := os.Create(ctrl.dbBackupName)
	if err != nil {
		return err
	}
	defer saveFile.Close()

	if _, err := io.Copy(saveFile, data); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) loadDB(data []byte) error {
	dbFile, err := os.OpenFile(ctrl.dbURL, os.O_RDWR, 0o664)
	if err != nil {
		return fmt.Errorf("failed opening db file: %w", err)
	}

	if _, err := dbFile.Write(data); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) deleteStandartAssets() error {
	files, err := os.ReadDir(ctrl.assetsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(filepath.Join(ctrl.assetsDir, file.Name())); err != nil {
			return err
		}
	}

	return nil
}

func parseCiphers(keys []byte) ([]cipher.AESCipher, error) {
	stringKeys := strings.Split(string(keys), "\n")
	ciphs := make([]cipher.AESCipher, 0, len(stringKeys))

	for _, key := range stringKeys {
		ciph, err := cipher.New(key)
		if err != nil {
			return nil, fmt.Errorf("failed creating cipher: %w", err)
		}
		ciphs = append(ciphs, *ciph)
	}

	return ciphs, nil
}

func splitBackupData(data []byte) ([]byte, []byte, error) {
	peaces := strings.Split(string(data), "---data---")
	if len(peaces) != 2 {
		return nil, nil, fmt.Errorf("separator not found")
	}

	return []byte(peaces[0][:len(peaces[0])-1]), []byte(peaces[1]), nil
}
