package files

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftrbnd/film-sync/internal/cloudinary"
	"github.com/ftrbnd/film-sync/internal/google"
)

func Upload(from string, zip string, count int) (string, string, string, error) {
	folderName := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	_, err := os.ReadDir(from)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read directory: %v", err)
	}

	driveFolderID, cldFolderName, err := createFolders(folderName)
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

		if format == ".png" {
			err = cloudinary.UploadImage(cldFolderName, path)
			if err != nil {
				return err
			}

		} else if format == ".tif" {
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

			err = google.Upload(fileBytes, path, driveFolderID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", "", "", err
	}

	cleanUp(from, zip)
	message := fmt.Sprintf("Finished uploading **%s** (%d new photos)", folderName, count)
	return folderName, driveFolderID, message, nil
}

func SetFolderNames(cldFolder string, driveFolderID string, new string) error {
	err := cloudinary.SetFolderName(cldFolder, new)
	if err != nil {
		return err
	}

	err = google.SetFolderName(driveFolderID, new)
	if err != nil {
		return err
	}

	return nil
}

func FolderLinks(cldFolder string, driveFolderID string) (string, string, error) {
	cldUrl, err := cloudinary.FolderLink(cldFolder)
	if err != nil {
		return "", "", err
	}
	driveUrl := google.FolderLink(driveFolderID)

	return cldUrl, driveUrl, nil
}

func createFolders(name string) (string, string, error) {
	driveFolderID, err := google.CreateFolder(name)
	if err != nil {
		return "", "", err
	}

	cldFolder, err := cloudinary.CreateFolder(name)
	if err != nil {
		return "", "", err
	}

	return driveFolderID, cldFolder.Name, nil
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
