package archivator

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(src, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed creating file: %v", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	if err := filepath.Walk(src, walkerClosure(src, w)); err != nil {
		return fmt.Errorf("failed compressing file: %v", err)
	}
	return nil
}

func Decompress(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed oprening zip reader: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if err := prepareFS(dst, f.Name); err != nil {
			return fmt.Errorf("failed preparing fs: %v", err)
		}

		dstFile, err := os.Create(filepath.Join(dst, f.Name))
		if err != nil {
			return fmt.Errorf("failed creating file: %v", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed opening zip entity: %v", err)
		}

		if _, err = io.Copy(dstFile, rc); err != nil {
			return fmt.Errorf("failed decompressing file: %v", err)
		}
		rc.Close()
		dstFile.Close()
	}

	return nil
}

func walkerClosure(src string, w *zip.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(toZipRoot(src, path))
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, file); err != nil {
			return err
		}

		return nil
	}
}

func prepareFS(root, path string) error {
	dir, _ := filepath.Split(path)
	return os.MkdirAll(filepath.Join(root, dir), 0o764)
}

func toZipRoot(root, path string) string {
	if root[0] == '.' {
		root = root[2:]
	}

	res, _ := strings.CutPrefix(path, root)
	if res[0] == '/' {
		return res[1:]
	}

	return res
}
