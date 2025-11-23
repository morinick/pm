package starter

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"passman/internal/server/backups"
	"passman/internal/server/starter/queries"
	"passman/pkg/cipher"

	"github.com/google/uuid"
)

var errInvalidKey = errors.New("invalid key encoding")

type StartOptions struct {
	DB               *sql.DB
	MasterKey        string
	BackupController *backups.Controller
}

func Start(ctx context.Context, opts StartOptions) error {
	keys, err := queries.New(opts.DB).GetKeys(ctx)
	if err != nil {
		return err
	}

	// Check if it's the first initialization
	if len(keys) == 0 {
		if len(opts.MasterKey) == 0 {
			return initData(ctx, opts)
		}

		if err := opts.BackupController.LoadBackup(); err != nil {
			return fmt.Errorf("failed loading backup: %w", err)
		}

		keys, err := queries.New(opts.DB).GetKeys(ctx)
		if err != nil {
			return err
		}

		ciphs, err := makeCiphers(keys, opts.MasterKey)
		if err != nil {
			return err
		}

		opts.BackupController.Ciphers = ciphs

		return nil
	}

	// Means that container was restarted
	if len(opts.MasterKey) == 0 {
		return fmt.Errorf("key not found")
	}

	ciphs, err := makeCiphers(keys, opts.MasterKey)
	if err != nil {
		return err
	}

	opts.BackupController.Ciphers = ciphs

	return nil
}

func initData(ctx context.Context, opts StartOptions) error {
	masterCipher, err := backups.GenerateCipher()
	if err != nil {
		return err
	}

	ciphers, err := backups.GenerateCiphers(10)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(ciphers))
	for _, ciph := range ciphers {
		binaryKey := masterCipher.Encrypt([]byte(ciph.Key()))
		keys = append(keys, hex.EncodeToString(binaryKey))
	}

	if err := addKeysToDB(ctx, opts.DB, keys); err != nil {
		return err
	}

	if err := addAssetsToDB(ctx, opts.DB, opts.BackupController.AssetsDir); err != nil {
		return err
	}

	opts.BackupController.Key = masterCipher.Key()
	opts.BackupController.Ciphers = ciphers

	return nil
}

func makeCiphers(keys []string, masterKey string) ([]cipher.AESCipher, error) {
	masterCipher, err := cipher.New(masterKey)
	if err != nil {
		return nil, err
	}

	result := make([]cipher.AESCipher, 0, len(keys))
	for _, key := range keys {
		binaryKey, err := hex.DecodeString(key)
		if err != nil {
			return nil, errInvalidKey
		}

		decryptedKey := masterCipher.Decrypt(binaryKey)

		ciph, err := cipher.New(string(decryptedKey))
		if err != nil {
			return nil, err
		}

		result = append(result, *ciph)
	}

	return result, nil
}

func addKeysToDB(ctx context.Context, db *sql.DB, keys []string) (err error) {
	sqlTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed starting transaction: %w", err)
	}
	defer func() {
		rollbackErr := sqlTx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, rollbackErr)
		}
	}()

	tx := queries.New(db).WithTx(sqlTx)

	for _, key := range keys {
		param := queries.AddKeysParams{
			ID:       uuid.New(),
			KeyValue: key,
		}

		if err = tx.AddKeys(ctx, param); err != nil {
			return err
		}
	}

	return sqlTx.Commit()
}

func addAssetsToDB(ctx context.Context, db *sql.DB, assetsPath string) (err error) {
	sqlTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed starting transaction: %w", err)
	}
	defer func() {
		rollbackErr := sqlTx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, rollbackErr)
		}
	}()

	tx := queries.New(db).WithTx(sqlTx)

	assetsDir, err := os.ReadDir(assetsPath)
	if err != nil {
		return fmt.Errorf("failed reading assets directory: %w", err)
	}

	for _, file := range assetsDir {
		param := queries.AddAssetsParams{
			ID:   uuid.New(),
			Name: file.Name(),
			Logo: filepath.Join(assetsPath, file.Name()),
		}
		if err = tx.AddAssets(ctx, param); err != nil {
			return err
		}
	}

	return sqlTx.Commit()
}
