package google

import (
	"bytes"
	"log"
	"path/filepath"

	"github.com/ftrbnd/film-sync/internal/util"
	"google.golang.org/api/drive/v3"
)

func CreateFolder(service *drive.Service, name string) string {
	parent := util.LoadEnvVar("DRIVE_FOLDER_ID")

	res, err := service.Files.Create(&drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     name,
		Parents:  []string{parent},
	}).Do()
	util.CheckError("Failed to create folder", err)

	return res.Id
}

func Upload(bytes *bytes.Reader, filePath string, folderID string, service *drive.Service) error {
	name := filepath.Base(filePath)

	_, err := service.Files.Create(&drive.File{
		Parents:  []string{folderID},
		Name:     name,
		MimeType: "image/tiff",
	}).Media(bytes).Do()

	if err != nil {
		return err
	} else {
		log.Default().Printf("[Google Drive] Uploaded %s!\n", name)
		return nil
	}
}
