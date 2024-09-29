package files

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/imgconv"
)

func Unzip(filename string, dst string, format string) error {
	archive, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		prefix := filepath.Clean(dst) + string(os.PathSeparator)

		if !strings.HasPrefix(filePath, prefix) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		if !strings.HasSuffix(filePath, format) {
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open %s: %v", filePath, err)
		}

		fileInZip, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open: %v", err)
		}

		_, err = io.Copy(dstFile, fileInZip)
		if err != nil {
			return fmt.Errorf("failed to copy to destination: %v", err)
		}

		log.Default().Println("Saved", filePath)

		dstFile.Close()
		fileInZip.Close()
	}

	return nil
}

func ConvertToPNG(format string, dir string) (int, error) {
	_, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read directory: %v", err)
	}

	count := 0
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		return visit(path, d, err, format, &count)
	})
	if err != nil {
		return count, err
	}

	log.Default().Println("Converted all files!")

	return count, nil
}

func visit(path string, d fs.DirEntry, err error, format string, c *int) error {
	if err != nil {
		return err
	}

	if d.IsDir() || !strings.HasSuffix(path, format) {
		return nil
	}

	src, err := imgconv.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}

	pngPath := strings.Replace(path, "tif", "png", 1)
	dstFile, err := os.Create(pngPath)
	if err != nil {
		return fmt.Errorf("failed to create .png file: %v", err)
	}

	err = imgconv.Write(dstFile, src, &imgconv.FormatOption{Format: imgconv.PNG})
	if err != nil {
		return fmt.Errorf("failed to convert image: %v", err)
	}

	log.Default().Printf("Converted %s", pngPath)
	*c++
	return nil
}
