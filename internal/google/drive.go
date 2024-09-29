package google

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ftrbnd/film-sync/internal/util"
	"google.golang.org/api/drive/v3"
)

func CreateFolder(service *drive.Service, name string) (string, error) {
	parent, err := util.LoadEnvVar("DRIVE_FOLDER_ID")
	if err != nil {
		return "", err
	}

	res, err := service.Files.Create(&drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     name,
		Parents:  []string{parent},
	}).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %v", err)
	}

	return res.Id, nil
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

	}

	log.Default().Printf("[Google Drive] Uploaded %s!\n", name)
	return nil
}

func SetFolderName(url string, name string, service *drive.Service) error {
	folderID, _ := strings.CutPrefix(url, "https://drive.google.com/drive/u/0/folders/")

	_, err := service.Files.Update(folderID, &drive.File{
		Name: name,
	}).Do()
	if err != nil {
		return err

	}

	log.Default().Printf("[Google Drive] Set folder name to %s", name)
	return nil
}
