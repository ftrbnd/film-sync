package files

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	myaws "github.com/ftrbnd/film-sync/internal/aws"
	"github.com/ftrbnd/film-sync/internal/google"
)

func Upload(from string, zip string, count int) (string, string, []string, string, error) {
	folderName := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	_, err := os.ReadDir(from)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("failed to read directory: %v", err)
	}

	driveFolderID, err := google.CreateFolder(folderName)
	if err != nil {
		return "", "", nil, "", err
	}

	var keys []string
	err = filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		format := filepath.Ext(path)
		if d.IsDir() || (format != ".png" && format != ".tif") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		fileInfo, _ := file.Stat()
		size := fileInfo.Size()
		buffer := make([]byte, size) // read file content to buffer

		file.Read(buffer)
		fileBytes := bytes.NewReader(buffer)
		fileType := http.DetectContentType(buffer)

		if format == ".png" {
			key, e := myaws.Upload(fileBytes, fileType, size, folderName, path)
			err = e
			keys = append(keys, key)
		} else if format == ".tif" {
			err = google.Upload(fileBytes, path, driveFolderID)
		}

		return err
	})
	if err != nil {
		return "", "", nil, "", err
	}

	cleanUp(from, zip)
	message := fmt.Sprintf("Finished uploading **%s** (%d new photos)", folderName, count)
	return folderName, driveFolderID, keys, message, nil
}

func cleanUp(from string, zip string) {
	err := os.RemoveAll(from)
	if err != nil {
		log.Default().Printf("Failed to remove directory: %v", err)
	}

	err = os.Remove(zip)
	if err != nil {
		log.Default().Printf("Failed to remove zip file: %v", err)
	}
}
