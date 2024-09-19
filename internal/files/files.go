package files

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/sunshineplan/imgconv"
)

func OpenZip(filename string) string {
	dst := "output"
	var finalPath string

	archive, err := zip.OpenReader(filename)
	util.CheckError("Couldn't open .zip file", err)
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
		util.CheckError("Failed to create output directory", err)

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		util.CheckError(fmt.Sprintf("Failed to open %s", filePath), err)

		fileInZip, err := f.Open()
		util.CheckError("Failed to open", err)

		_, err = io.Copy(dstFile, fileInZip)
		util.CheckError("Failed to copy to destination", err)

		log.Default().Println("Saved ", filePath)

		dstFile.Close()
		fileInZip.Close()
	}

	return finalPath
}

func ConvertToPNG(directory string, format string) {
	items, err := os.ReadDir(directory)
	util.CheckError("Failed to read output directory", err)

	for _, item := range items {
		if !strings.HasSuffix(item.Name(), format) {
			log.Default().Printf("%s is not a .%s file", item.Name(), format)
			continue
		}

		filePath := filepath.Join(directory, item.Name())

		src, err := imgconv.Open(filePath)
		util.CheckError("Failed to open image", err)

		dstFile, err := os.Create(strings.Replace(filePath, "tif", "png", 1))
		util.CheckError("Failed to create .png file", err)

		err = imgconv.Write(dstFile, src, &imgconv.FormatOption{Format: imgconv.PNG})
		util.CheckError("Failed to convert image", err)
	}
}
