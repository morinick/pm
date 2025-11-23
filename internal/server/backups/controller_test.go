package backups

import (
	"os"
	"path/filepath"
	"testing"
)

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

	// test saving backup
	ControllerToSave := New(testDBFilename, testBackupDir, "")
	ControllerToSave.AssetsDir = testAssetsDir
	ControllerToSave.assetsBackupName = filepath.Join(testBackupDir, "assets.zip")

	saveErr := ControllerToSave.SaveBackup()
	if saveErr != nil {
		t.Fatalf("Unexpected error: %v", saveErr)
	}

	// test loading backup
	decryptKey := ControllerToSave.Key
	ControllerToLoad := New(testDBFilename, testBackupDir, decryptKey)
	ControllerToLoad.AssetsDir = testAssetsDir
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
}
