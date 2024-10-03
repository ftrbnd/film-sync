package google

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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

	log.Default().Printf("[Google Drive] Uploaded %s!\n", name)
	return nil
}

func SetFolderName(folderID string, name string) error {
	_, err := driveSrv.Files.Update(folderID, &drive.File{
		Name: name,
	}).Do()
	if err != nil {
		return err
	}

	log.Default().Printf("[Google Drive] Set folder name to %s", name)
	return nil
}

func FolderLink(folderID string) string {
	return fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", folderID)
}

func DownloadFolder(folderID string, dst string) error {
	// Create the local directory if it doesn't exist
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		log.Fatalf("unable to create local directory: %v", err)
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	fileList, err := driveSrv.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return fmt.Errorf("unable to list files in folder: %v", err)
	}

	// Check if there are any files in the folder
	if len(fileList.Files) == 0 {
		log.Println("No files found in the folder.")
		return nil
	}

	log.Default().Printf("%d FILES FOUND", len(fileList.Files))
	log.Default().Println("FILE NAME:", fileList.Files[0].Name)

	// Download each file in the folder
	for _, file := range fileList.Files {
		if file.MimeType == "application/vnd.google-apps.folder" {
			// If it's a folder, call downloadFolder recursively
			subFolderPath := filepath.Join(dst, file.Name)
			fmt.Printf("Entering folder: %s\n", subFolderPath)
			DownloadFolder(file.Id, subFolderPath)
		} else {
			// If it's a file, download it
			fmt.Printf("Downloading file: %s\n", file.Name)
			downloadFile(file.Id, filepath.Join(dst, file.Name))
		}
	}

	return nil
}

// downloadFile downloads a single file from Google Drive
func downloadFile(fileID, dst string) {
	// Create the request to download the file content
	response, err := driveSrv.Files.Get(fileID).Download()
	if err != nil {
		log.Fatalf("Unable to download file: %v", err)
	}
	defer response.Body.Close()

	// Create a local file to store the downloaded content
	outFile, err := os.Create(dst)
	if err != nil {
		log.Fatalf("Unable to create local file: %v", err)
	}
	defer outFile.Close()

	// Copy the content to the local file
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		log.Fatalf("Unable to copy file content: %v", err)
	}

	fmt.Printf("File %s downloaded successfully!\n", dst)
}
