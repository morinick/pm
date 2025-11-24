package backups

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"passman/pkg/archivator"
	"passman/pkg/cipher"
)

type ControllerOptions struct {
	DBURL     string
	BackupDir string
	AssetsDir string
	MasterKey string
}

type Controller struct {
	backupDir        string
	dbURL            string
	dbBackupName     string
	assetsDir        string
	assetsBackupName string
	Key              string
}

func New(opts ControllerOptions) *Controller {
	return &Controller{
		backupDir:        opts.BackupDir,
		dbURL:            opts.DBURL,
		dbBackupName:     filepath.Join(opts.BackupDir, "db.bak"),
		assetsDir:        opts.AssetsDir,
		assetsBackupName: filepath.Join(opts.BackupDir, "assets.zip"),
		Key:              opts.MasterKey,
	}
}

func (ctrl *Controller) LoadBackup() error {
	// Check existing components
	if err := ctrl.checkComponents(); err != nil {
		return err
	}

	// Load db
	encryptedDB, err := os.ReadFile(ctrl.dbBackupName)
	if err != nil {
		return fmt.Errorf("failed reading db backup: %w", err)
	}

	decryptedDB, err := ctrl.decryptDB(encryptedDB)
	if err != nil {
		return fmt.Errorf("failed decrypting %s: %w", ctrl.dbBackupName, err)
	}

	if err := ctrl.loadDB(decryptedDB); err != nil {
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

	dbData, err := os.ReadFile(ctrl.dbURL)
	if err != nil {
		return fmt.Errorf("failed reading db: %w", err)
	}
	buf := bytes.NewBuffer(dbData)

	encryptedDB, err := ctrl.encryptDB(buf, ciph)
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
		ciph, err := cipher.GenerateCipher()
		ctrl.Key = ciph.Key()
		return ciph, err
	}
	return cipher.New(ctrl.Key)
}

func (ctrl *Controller) checkComponents() error {
	if len(ctrl.Key) == 0 {
		return fmt.Errorf("empty key")
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
