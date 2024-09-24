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

func Unzip(filename string, dst string, format string) {
	archive, err := zip.OpenReader(filename)
	util.CheckError("Couldn't open .zip file", err)
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		prefix := filepath.Clean(dst) + string(os.PathSeparator)

		if !strings.HasPrefix(filePath, prefix) {
			return
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		util.CheckError("Failed to create output directory", err)

		if !strings.HasSuffix(filePath, format) {
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		util.CheckError(fmt.Sprintf("Failed to open %s", filePath), err)

		fileInZip, err := f.Open()
		util.CheckError("Failed to open", err)

		_, err = io.Copy(dstFile, fileInZip)
		util.CheckError("Failed to copy to destination", err)

		log.Default().Println("Saved", filePath)

		dstFile.Close()
		fileInZip.Close()
	}
}

func ConvertToPNG(format string, dir string) int {
	_, err := os.ReadDir(dir)
	util.CheckError("Failed to read directory", err)

	count := 0
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		return visit(path, d, err, format, &count)
	})
	util.CheckError("Failed to walk through directory", err)

	log.Default().Println("Converted all files!")

	return count
}

func visit(path string, d fs.DirEntry, err error, format string, c *int) error {
	if err != nil {
		return err
	}

	if d.IsDir() || !strings.HasSuffix(path, format) {
		return nil
	}

	src, err := imgconv.Open(path)
	util.CheckError("Failed to open image", err)

	pngPath := strings.Replace(path, "tif", "png", 1)
	dstFile, err := os.Create(pngPath)
	util.CheckError("Failed to create .png file", err)

	err = imgconv.Write(dstFile, src, &imgconv.FormatOption{Format: imgconv.PNG})
	util.CheckError("Failed to convert image", err)

	log.Default().Printf("Converted %s", pngPath)
	*c++
	return nil
}
