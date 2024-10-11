package google

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"

	"github.com/ftrbnd/film-sync/internal/util"
	"google.golang.org/api/drive/v3"
)

func CreateFolder(name string) (string, error) {
	parent, err := util.LoadEnvVar("DRIVE_FOLDER_ID")
	if err != nil {
		return "", err
	}

	res, err := driveSrv.Files.Create(&drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     name,
		Parents:  []string{parent},
	}).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %v", err)
	}

	return res.Id, nil
}

func Upload(bytes *bytes.Reader, filePath string, folderID string) error {
	name := filepath.Base(filePath)

	_, err := driveSrv.Files.Create(&drive.File{
		Parents:  []string{folderID},
		Name:     name,
		MimeType: "image/tiff",
	}).Media(bytes).Do()
	if err != nil {
		return err

	}

	log.Default().Printf("[Google] Uploaded %s to Drive!\n", name)
	return nil
}

func SetFolderName(folderID string, name string) error {
	_, err := driveSrv.Files.Update(folderID, &drive.File{
		Name: name,
	}).Do()
	if err != nil {
		return err
	}

	log.Default().Printf("[Google] Set Drive folder name to %s", name)
	return nil
}

func FolderLink(folderID string) string {
	return fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", folderID)
}
