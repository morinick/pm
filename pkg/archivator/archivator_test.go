package archivator

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestToZipRoot(t *testing.T) {
	tests := []struct {
		name      string
		root      string
		path      string
		expResult string
	}{
		{
			name:      "check_absolute",
			root:      "/some/absolute/root",
			path:      "/some/absolute/root/to/path",
			expResult: "to/path",
		},
		{
			name:      "check_relative",
			root:      "./some/absolute/root",
			path:      "some/absolute/root/to/path",
			expResult: "to/path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actResult := toZipRoot(test.root, test.path)

			if got, want := actResult, test.expResult; got != want {
				t.Errorf("Wrong! Unexpected result!\n\tExpected: %s\n\tActual: %s", want, got)
			}
		})
	}
}

func TestCompressAndDecompress(t *testing.T) {
	root := t.TempDir()
	testDir := filepath.Join(root, "test_dir")

	paths := []string{
		"/dir1/file1.txt",
		"/dir1/file2.txt",
		"/dir2/inner_dir1/file3.txt",
		"/file4.txt",
	}

	for _, path := range paths {
		// prepare temp fs
		abs := filepath.Join(testDir, path)
		dir, _ := filepath.Split(abs)
		_ = os.MkdirAll(dir, 0o764)

		// create test file
		file, _ := os.Create(abs)

		// fill test file by random data
		buf := make([]byte, 32)
		_, _ = rand.Read(buf)
		_, _ = file.Write(buf)

		file.Close()
	}

	testZipFile := filepath.Join(root, "test.zip")
	testDecompressedDir := filepath.Join(root, "test_dir_decompressed")

	// try compressing
	if err := Compress(testDir, testZipFile); err != nil {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", nil, err)
	}

	// try decompressing
	if err := Decompress(testZipFile, testDecompressedDir); err != nil {
		t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", nil, err)
	}

	// check result (compare decompressed directory with original)
	for _, path := range paths {
		originalData, _ := os.ReadFile(filepath.Join(testDir, path))
		decompressedData, _ := os.ReadFile(filepath.Join(testDecompressedDir, path))

		if slices.Compare(originalData, decompressedData) != 0 {
			t.Errorf("Wrong! Mismatch data in %s file!", path)
		}
	}
}
