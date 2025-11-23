package starter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMakeCiphers(t *testing.T) {
	masterKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	correctKeys := []string{
		"db745dca87ba28d883587ec0670af6ad15ace4f49e61ff738a42186f78b437b6",
		"d269f2c593ad43c06e1a50217f496399d98225fdf5d0e76346e2c7e2fe704025",
		"d7a7b882279c0a0cf6f0bf4414e071fdf8ba8fe592ee191f1666cbaaacb56778",
	}
	correctEncryptedKeys := []string{
		"7bb0c2e0383f5561779d6d318ad589325431caf9c00f0920562d6183ee8137310e16c225d2414d571481886fe5755b9cfd5524d643692cf672ebfd7ef43ed38b",
		"b96aa232fca55abdfcfa15cc24a90a0d52af7fdfe77476be6b5fa530d87ea731a6aee0083661c13dfe450fb3b6fe3cce7baf20b869e12895f09529fd968c6d0d",
		"3649b33e0ba649f1260a8923b941a04a9dc516d3201550a55b35be4b1c5242e4acf2ea93b48a8eaeb3385d2336aa01ddc8369fa8c11f05bca8c2e22a44b751c7",
	}

	incorrectEncryptedKeys := []string{"incorrect key"}

	tests := []struct {
		name      string
		inputKeys []string
		expErr    error
	}{
		{
			name:      "invalid_key",
			inputKeys: incorrectEncryptedKeys,
			expErr:    errInvalidKey,
		},
		{
			name:      "success",
			inputKeys: correctEncryptedKeys,
			expErr:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actCiphs, actErr := makeCiphers(test.inputKeys, masterKey)

			if got, want := actErr, test.expErr; got != want {
				t.Errorf("Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if test.name == "success" {
				for i, ciph := range actCiphs {
					if ciph.Key() != correctKeys[i] {
						t.Errorf("Mismatch key!\n\tExpected: %s\n\tActual: %s", correctKeys[i], ciph.Key())
					}
				}
			}
		})
	}
}

func TestAddKeysToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed creatng sqlmock: %v", err)
	}

	ctx := context.Background()
	expectRollback := errors.New("rollback")
	errorTriggerKey := "error-trigger"

	tests := []struct {
		name   string
		keys   []string
		expErr error
	}{
		{
			name:   "failed_transaction",
			keys:   []string{"key1", "key2", errorTriggerKey, "key4"},
			expErr: expectRollback,
		},
		{
			name:   "successful_transaction",
			keys:   []string{"key1", "key2", "key3"},
			expErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock.ExpectBegin()

			for _, key := range test.keys {
				q := mock.ExpectExec("insert into ciphers").WithArgs(sqlmock.AnyArg(), key)
				if key == errorTriggerKey {
					q.WillReturnError(expectRollback)
					break
				}
				q.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			if test.expErr == expectRollback {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}

			actErr := addKeysToDB(ctx, db, test.keys)

			if got, want := actErr, test.expErr; got != want {
				t.Errorf("Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestAddAssetsToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed creatng sqlmock: %v", err)
	}

	ctx := context.Background()
	expectRollback := errors.New("rollback")
	errorTriggerFile := "error-trigger"

	tempDir := t.TempDir()

	failedDir := filepath.Join(tempDir, "failed")
	os.Mkdir(failedDir, 0o764)
	failedDirFiles := []string{"file1", "file2", errorTriggerFile}
	for _, filename := range failedDirFiles {
		_, err := os.Create(filepath.Join(failedDir, filename))
		if err != nil {
			t.Fatalf("Failed creating file: %v", err)
		}
	}

	successDir := filepath.Join(tempDir, "success")
	os.Mkdir(successDir, 0o764)
	successDirFiles := []string{"file1", "file2", "file3"}
	for _, filename := range successDirFiles {
		_, err := os.Create(filepath.Join(successDir, filename))
		if err != nil {
			t.Fatalf("Failed creating file: %v", err)
		}
	}

	tests := []struct {
		name      string
		assetsDir string
		expErr    error
	}{
		{
			name:      "failed_transaction",
			assetsDir: failedDir,
			expErr:    expectRollback,
		},
		{
			name:      "successful_transaction",
			assetsDir: successDir,
			expErr:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock.ExpectBegin()

			files, err := os.ReadDir(test.assetsDir)
			if err != nil {
				t.Fatalf("Failed reading directory: %v", err)
			}

			for _, file := range files {
				q := mock.ExpectExec("insert into services").
					WithArgs(
						sqlmock.AnyArg(),
						file.Name(),
						filepath.Join(test.assetsDir, file.Name()),
					)
				if file.Name() == errorTriggerFile {
					q.WillReturnError(expectRollback)
					break
				}
				q.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			if test.expErr == expectRollback {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}

			actErr := addAssetsToDB(ctx, db, test.assetsDir)

			if got, want := actErr, test.expErr; got != want {
				t.Errorf("Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}
