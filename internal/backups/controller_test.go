package backups

import (
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"

	"passman/pkg/cipher"
)

type testError struct {
	msg string
}

func newTestError(msg string) error {
	return &testError{msg: msg}
}

func (te *testError) Error() string {
	return te.msg
}

func (te *testError) Is(target error) bool {
	return strings.Contains(target.Error(), te.Error())
}

func TestSaveDatabase(t *testing.T) {
	testDBFileData := `d5dae80065273421f95e5a3d7f4702be5ab9209614b01607fa5c6630acd672df
8f0d643fd2e92070fc37049ecae4b93e3bcfe28bccfcc63698fec23135cbfe27
c7c6873edb8b518fa3742f3ad16bebde3e5593fd07e4dfb14752ee85e5c1ff10
---data---736f6d652073696d706c652073746f726167652064617461`
	testDBFileName := "test_db_filename.txt"
	file, _ := os.Create(testDBFileName)
	_, _ = file.Write([]byte(testDBFileData))
	file.Close()
	defer os.Remove(testDBFileName)

	testInput := "file_for_test.txt"
	defer os.Remove(testInput)

	tests := []struct {
		name       string
		input      string
		dbFileName string
		expResult  error
	}{
		{
			name:       "empty_db_filename",
			dbFileName: "",
			expResult:  newTestError("failed reading data from file: "),
		},
		{
			name:       "empty_input_filename",
			input:      "",
			dbFileName: testDBFileName,
			expResult:  newTestError("failed creating backup file: "),
		},
		{
			name:       "success",
			input:      testInput,
			dbFileName: testDBFileName,
			expResult:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrlr := &Controller{
				dbFileName: test.dbFileName,
				bakupDir:   ".",
			}

			actErr := ctrlr.SaveDatabase(test.input)

			if got, want := actErr, test.expResult; !errors.Is(want, got) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if test.expResult == nil {
				if _, err := hex.DecodeString(ctrlr.DecryptKey); err != nil {
					t.Errorf("Wrong! Returned key must be in hex encoding!")
				}
			}
		})
	}
}

func TestLoadDatabase(t *testing.T) {
	// Preparing encrypted database for successful test result
	payload := "736f6d652073696d706c652073746f726167652064617461"
	file, _ := os.Create("test_db_file.txt")
	_, _ = file.Write([]byte(payload))
	file.Close()
	defer os.Remove("test_db_file.txt")

	key1, _ := hex.DecodeString("d5dae80065273421f95e5a3d7f4702be5ab9209614b01607fa5c6630acd672df")
	ciph1, _ := cipher.New(key1)
	key2, _ := hex.DecodeString("8f0d643fd2e92070fc37049ecae4b93e3bcfe28bccfcc63698fec23135cbfe27")
	ciph2, _ := cipher.New(key2)

	c := &Controller{
		dbFileName: "test_db_file.txt",
		Ciphers:    []cipher.AESCipher{*ciph1, *ciph2},
		bakupDir:   ".",
	}

	testInputFileName := "test_backup_file.txt"
	_ = c.SaveDatabase(testInputFileName)
	defer os.Remove(testInputFileName)

	decryptKey := c.DecryptKey

	// For checking error with creating new cipher
	invalidDecryptKey := "d5dae80065273421f95e5a3d7f4702be5ab9209614b01607fa5c6630acd6"

	tests := []struct {
		name       string
		decryptKey string
		input      string
		dbFileName string
		expErr     error
	}{
		{
			name:       "empty_key",
			decryptKey: "",
			expErr:     newTestError("empty decryption key"),
		},
		{
			name:       "empty_input_file",
			decryptKey: "sss",
			input:      "",
			expErr:     newTestError("failed reading data from file: "),
		},
		{
			name:       "another_key_encoding",
			decryptKey: "sss",
			input:      testInputFileName,
			expErr:     newTestError("failed decoding key from hex: "),
		},
		{
			name:       "invalid_binary_key_size",
			decryptKey: invalidDecryptKey,
			input:      testInputFileName,
			expErr:     newTestError("crypto/aes: invalid key size "),
		},
		{
			name:       "empty_db_file",
			decryptKey: decryptKey,
			input:      testInputFileName,
			dbFileName: "",
			expErr:     newTestError("failed opening db file: "),
		},
		{
			name:       "success",
			decryptKey: decryptKey,
			input:      testInputFileName,
			dbFileName: c.dbFileName,
			expErr:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrlr := &Controller{
				dbFileName: test.dbFileName,
				bakupDir:   ".",
				DecryptKey: test.decryptKey,
			}

			actErr := ctrlr.LoadDatabase(test.input)

			if got, want := actErr, test.expErr; !errors.Is(want, got) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			// Check payload in decrypted db file
			if test.expErr == nil {
				checkData, _ := os.ReadFile(ctrlr.dbFileName)
				if payload != string(checkData) {
					t.Errorf("Wrong! Payload was changed (lost data)!\n\tExpected: %s\n\tActual: %s", payload, string(checkData))
				}
			}
		})
	}
}

func TestAddCiphers(t *testing.T) {
	ctrlr := &Controller{Ciphers: nil}

	correctKeys := `10db8e3e72f4e87abafa0628c8f960214ee2cad25cf35efc9ac466d1c760a620
29acce1aeff41bf68c5311d107cfbc44cac7ed542f85a06eff01efac5aeda6ab
1e0d340f9caf15133e6e7846b16303b51d0052c0a0e04ed1b5e65ac9bd079219`
	incorrectKeys := `some not hexed keys`
	incorrectBitesCount := `10db8e3e72f4e87abafa0628c8f960214ee2cad25cf35efc9ac466d1c760a6`

	tests := []struct {
		name   string
		input  []byte
		expErr error
	}{
		{
			name:   "key_not_in_hex",
			input:  []byte(incorrectKeys),
			expErr: newTestError("failed decoding internal key from hex: "),
		},
		{
			name:   "failed_creating_cipher",
			input:  []byte(incorrectBitesCount),
			expErr: newTestError("failed creating cipher: "),
		},
		{
			name:   "success",
			input:  []byte(correctKeys),
			expErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actErr := ctrlr.addCiphers(test.input)

			if got, want := actErr, test.expErr; !errors.Is(want, got) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if test.expErr == nil {
				l := 3
				if len(ctrlr.Ciphers) != l {
					t.Errorf("Wrong! Not all ciphers saved!\n\tExpected: %d\n\tActual: %d", l, len(ctrlr.Ciphers))
				}
			}
		})
	}
}

func TestParseDataFromFile(t *testing.T) {
	// Preparing test data
	data := "some data"

	file, _ := os.Create("test_parsing.txt")
	defer os.Remove(file.Name())
	_, _ = file.Write([]byte(data))
	file.Close()

	var emptyErr error
	readingErr := newTestError("failed reading data from file: ")

	// Tests
	// Success
	actData, actErr := parseDataFromFile(file.Name())

	if got, want := string(actData), data; got != want {
		t.Errorf("Wrong! Incorrect data!\n\tExpected: %s\n\tActual: %s", want, got)
	}

	if got, want := actErr, emptyErr; !errors.Is(want, got) {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
	}

	// Error
	_, actErr = parseDataFromFile("incorrect_file_name.txt")

	if got, want := actErr, readingErr; !errors.Is(want, got) {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
	}
}

func TestSplitBackupData(t *testing.T) {
	// Preparing test data
	data := `10db8e3e72f4e87abafa0628c8f960214ee2cad25cf35efc9ac466d1c760a620
29acce1aeff41bf68c5311d107cfbc44cac7ed542f85a06eff01efac5aeda6ab
1e0d340f9caf15133e6e7846b16303b51d0052c0a0e04ed1b5e65ac9bd079219
---data---736f6d652073696d706c652073746f726167652064617461`

	incorrectData := `some data without separator`

	expectedSuccessKeys := `10db8e3e72f4e87abafa0628c8f960214ee2cad25cf35efc9ac466d1c760a620
29acce1aeff41bf68c5311d107cfbc44cac7ed542f85a06eff01efac5aeda6ab
1e0d340f9caf15133e6e7846b16303b51d0052c0a0e04ed1b5e65ac9bd079219`

	expectedSuccessData := `736f6d652073696d706c652073746f726167652064617461`

	var emptyErr error
	separatorErr := newTestError("separator not found")

	// Tests
	// Success
	actualKeys, actualData, actualErr := splitBackupData([]byte(data))

	if got, want := string(actualKeys), expectedSuccessKeys; got != want {
		t.Errorf("Wrong! Incorrect keys!\n\tExpected: %s\n\tActual: %s", want, got)
	}

	if got, want := string(actualData), expectedSuccessData; got != want {
		t.Errorf("Wrong! Incorrect data!\n\tExpected: %s\n\tActual: %s", want, got)
	}

	if got, want := actualErr, emptyErr; !errors.Is(want, got) {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
	}

	// Error
	_, _, actualErr = splitBackupData([]byte(incorrectData))

	if got, want := actualErr, separatorErr; !errors.Is(want, got) {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
	}
}
