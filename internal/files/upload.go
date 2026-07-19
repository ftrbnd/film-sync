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

	cldFolderName, driveFolderID, err := createFolders(folderName)
	if err != nil {
		return "", "", "", err
	}

	driveSkipped := driveFolderID == ""
	var driveFailures int

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
			return nil
		}

		// .tif -> Google Drive (best-effort; never abort Cloudinary uploads)
		if driveFolderID == "" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Default().Printf("[Google] Skipping Drive upload for %s: %v", path, err)
			driveFailures++
			return nil
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			log.Default().Printf("[Google] Skipping Drive upload for %s: %v", path, err)
			driveFailures++
			return nil
		}
		size := fileInfo.Size()
		buffer := make([]byte, size)
		if _, err := file.Read(buffer); err != nil {
			log.Default().Printf("[Google] Skipping Drive upload for %s: %v", path, err)
			driveFailures++
			return nil
		}
		fileBytes := bytes.NewReader(buffer)

		err = google.Upload(fileBytes, path, driveFolderID)
		if err != nil {
			log.Default().Printf("[Google] Drive upload failed for %s (continuing Cloudinary): %v", filepath.Base(path), err)
			driveFailures++
			return nil
		}

		return nil
	})
	if err != nil {
		return "", "", "", err
	}

	cleanUp(from, zip)
	message := fmt.Sprintf("Finished uploading **%s** (%d new photos)", folderName, count)
	if driveSkipped {
		message += "\n⚠️ Google Drive folder could not be created; Cloudinary upload completed."
	} else if driveFailures > 0 {
		message += fmt.Sprintf("\n⚠️ Google Drive: %d file(s) failed to upload; Cloudinary upload completed.", driveFailures)
	}
	return cldFolderName, driveFolderID, message, nil
}

func SetFolderNames(cldFolder string, driveFolderID string, new string) error {
	err := cloudinary.SetFolderName(cldFolder, new)
	if err != nil {
		return err
	}

	if driveFolderID == "" {
		return nil
	}
	err = google.SetFolderName(driveFolderID, new)
	if err != nil {
		log.Default().Printf("[Google] Failed to rename Drive folder (Cloudinary rename succeeded): %v", err)
	}

	return nil
}

func FolderLinks(cldFolder string, driveFolderID string) (string, string, error) {
	cldUrl, err := cloudinary.FolderLink(cldFolder)
	if err != nil {
		return "", "", err
	}
	driveUrl := ""
	if driveFolderID != "" {
		driveUrl = google.FolderLink(driveFolderID)
	}

	return cldUrl, driveUrl, nil
}

// createFolders creates the Cloudinary folder (required) and Drive folder (best-effort).
// A Drive failure returns an empty driveFolderID and a nil error so Cloudinary can proceed.
func createFolders(name string) (cldFolderName string, driveFolderID string, err error) {
	cldFolder, err := cloudinary.CreateFolder(name)
	if err != nil {
		return "", "", err
	}

	driveFolderID, err = google.CreateFolder(name)
	if err != nil {
		log.Default().Printf("[Google] Failed to create Drive folder %q (continuing with Cloudinary only): %v", name, err)
		return cldFolder.Name, "", nil
	}

	return cldFolder.Name, driveFolderID, nil
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
