package backups

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestLoadAndSaveBackup(t *testing.T) {
	testDir := t.TempDir()
	testBackupDir := filepath.Join(testDir, "backup")
	_ = os.Mkdir(testBackupDir, 0o777)

	// prepare data.db
	testDBFilename := filepath.Join(testDir, "data.db")
	dbData := "some database data"
	testDBFile, err := os.Create(testDBFilename)
	if err != nil {
		t.Fatalf("Database not saved: %v", err)
	}
	_, _ = testDBFile.WriteString(dbData)
	testDBFile.Close()

	// prepare assets
	testAssetsDir := filepath.Join(testDir, "assets")
	_ = os.Mkdir(testAssetsDir, 0o777)
	file1, err := os.Create(filepath.Join(testAssetsDir, "file1"))
	if err != nil {
		t.Fatalf("Asset 1 not saved: %v", err)
	}
	_, _ = file1.WriteString("asset1")
	file1.Close()
	file2, err := os.Create(filepath.Join(testAssetsDir, "file2"))
	if err != nil {
		t.Fatalf("Asset 2 not saved: %v", err)
	}
	_, _ = file2.WriteString("asset2")
	file2.Close()

	// prepare ciphers
	ciph1, _ := GenerateCipher()
	ciph2, _ := GenerateCipher()

	// test saving backup
	ControllerToSave := New(testDBFilename, testBackupDir, "")
	ControllerToSave.assetsDir = testAssetsDir
	ControllerToSave.assetsBackupName = filepath.Join(testBackupDir, "assets.zip")
	ControllerToSave.Ciphers = append(ControllerToSave.Ciphers, *ciph1, *ciph2)

	saveErr := ControllerToSave.SaveBackup()
	if saveErr != nil {
		t.Fatalf("Unexpected error: %v", saveErr)
	}

	// test loading backup
	decryptKey := ControllerToSave.Key
	ControllerToLoad := New(testDBFilename, testBackupDir, decryptKey)
	ControllerToLoad.assetsDir = testAssetsDir
	ControllerToLoad.assetsBackupName = filepath.Join(testBackupDir, "assets.zip")

	loadErr := ControllerToLoad.LoadBackup()
	if loadErr != nil {
		t.Fatalf("Unexpected error: %v", loadErr)
	}

	// check db
	loadedDB, _ := os.ReadFile(testDBFilename)
	if string(loadedDB) != dbData {
		t.Fatalf("Wrong! Mismatch database data!\n\tExpect: %s\n\tActual: %s", dbData, string(loadedDB))
	}

	// check assets
	asset1, _ := os.ReadFile(filepath.Join(testAssetsDir, "file1"))
	if string(asset1) != "asset1" {
		t.Fatalf("Wrong! Mismatch asset1!\n\tExpect: asset1\n\tActual: %s", string(asset1))
	}
	asset2, _ := os.ReadFile(filepath.Join(testAssetsDir, "file2"))
	if string(asset2) != "asset2" {
		t.Fatalf("Wrong! Mismatch asset2!\n\tExpect: asset2\n\tActual: %s", string(asset2))
	}

	// check ciphers
	if len(ControllerToLoad.Ciphers) == 0 {
		t.Fatal("Cipher not saved")
	}
	if ControllerToLoad.Ciphers[0].Key() != ciph1.Key() {
		t.Fatalf("Wrong! Mismatch cipher1!\n\tExpect: %s\n\tActual: %s", ciph1.Key(), ControllerToLoad.Ciphers[0].Key())
	}
	if ControllerToLoad.Ciphers[1].Key() != ciph2.Key() {
		t.Fatalf("Wrong! Mismatch cipher2!\n\tExpect: %s\n\tActual: %s", ciph2.Key(), ControllerToLoad.Ciphers[1].Key())
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
