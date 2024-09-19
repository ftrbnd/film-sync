package files

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/imgconv"
)

func OpenZip(filename string) string {
	dst := "output"
	var finalPath string

	archive, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatalf("Couldn't open .zip file: %v", err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		prefix := filepath.Clean(dst) + string(os.PathSeparator)

		if !strings.HasPrefix(filePath, prefix) {
			log.Default().Println("Invalid file path: ", filePath)
			return ""
		}

		if f.FileInfo().IsDir() {
			log.Default().Printf("Creating %s directory...", filePath)
			os.MkdirAll(filePath, os.ModePerm)
			finalPath = filePath
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Fatalf("Failed to open %s: %v", filePath, err)
		}

		fileInZip, err := f.Open()
		if err != nil {
			log.Fatalf("Failed to open: %v", err)
		}

		_, err = io.Copy(dstFile, fileInZip)
		if err != nil {
			log.Fatalf("Failed to copy to destination: %v", err)
		}

		log.Default().Println("Saved ", filePath)

		dstFile.Close()
		fileInZip.Close()
	}

	return finalPath
}

func ConvertToPNG(directory string, format string) {
	items, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("Failed to read output directory: %v", err)
	}

	for _, item := range items {
		if !strings.HasSuffix(item.Name(), format) {
			log.Default().Printf("%s is not a .%s file", item.Name(), format)
			continue
		}

		filePath := filepath.Join(directory, item.Name())
		filePathPNG := strings.Replace(filePath, "tif", "png", 1)
		log.Default().Println("PATH: ", filePath)

		src, err := imgconv.Open(filePath)
		if err != nil {
			log.Fatalf("Failed to open image: %v", err)
		}

		dstFile, err := os.Create(filePathPNG)
		if err != nil {
			log.Fatalf("Failed to create destination file: %v", err)
		}

		err = imgconv.Write(dstFile, src, &imgconv.FormatOption{Format: imgconv.PNG})
		if err != nil {
			log.Fatalf("Failed to convert image: %v", err)
		}
	}

}
