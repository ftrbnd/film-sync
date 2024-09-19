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

	"github.com/ftrbnd/film-sync/internal/util"
	"github.com/sunshineplan/imgconv"
)

func OpenZip(filename string, dst string) {
	archive, err := zip.OpenReader(filename)
	util.CheckError("Couldn't open .zip file", err)
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		prefix := filepath.Clean(dst) + string(os.PathSeparator)

		if !strings.HasPrefix(filePath, prefix) {
			log.Default().Println("Invalid file path: ", filePath)
			return
		}

		if f.FileInfo().IsDir() {
			log.Default().Printf("Creating %s directory...", filePath)
			os.MkdirAll(filePath, os.ModePerm)
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
}

func ConvertToPNG(format string, dir string) {
	_, err := os.ReadDir(dir)
	util.CheckError("Failed to read directory", err)

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		return visit(path, d, err, format)
	})
	util.CheckError("Failed to walk through directory", err)
}

func visit(path string, d fs.DirEntry, err error, format string) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		return nil
	}

	if !strings.HasSuffix(path, format) {
		log.Default().Printf("%s is not a .%s file - skipping...", path, format)
		return nil
	}

	src, err := imgconv.Open(path)
	util.CheckError("Failed to open image", err)

	pngPath := strings.Replace(path, "tif", "png", 1)
	dstFile, err := os.Create(pngPath)
	util.CheckError("Failed to create .png file", err)

	err = imgconv.Write(dstFile, src, &imgconv.FormatOption{Format: imgconv.PNG})
	util.CheckError("Failed to convert image", err)

	log.Default().Printf("Converted %s!", pngPath)
	return nil
}
