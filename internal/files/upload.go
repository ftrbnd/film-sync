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

func Upload(from string, zip string, count int) (string, string, string, error) {
	folderName := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	_, err := os.ReadDir(from)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read directory: %v", err)
	}

	driveFolderID, err := google.CreateFolder(folderName)
	if err != nil {
		return "", "", "", err
	}

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
			err = myaws.Upload(fileBytes, fileType, size, folderName, path)
		} else if format == ".tif" {
			err = google.Upload(fileBytes, path, driveFolderID)
		}

		return err
	})
	if err != nil {
		return "", "", "", err
	}

	message := fmt.Sprintf("Finished uploading **%s** (%d new photos)", folderName, count)

	err = os.RemoveAll(from)
	if err != nil {
		log.Default().Printf("Failed to remove directory: %v", err)
	}

	err = os.Remove(zip)
	if err != nil {
		log.Default().Printf("Failed to remove zip file: %v", err)

	}

	return folderName, driveFolderID, message, nil
}
