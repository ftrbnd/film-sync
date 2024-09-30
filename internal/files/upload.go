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
	"github.com/ftrbnd/film-sync/internal/util"
)

func Upload(from string, zip string, count int) (string, string, string, error) {
	folder := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	_, err := os.ReadDir(from)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read directory: %v", err)
	}

	folderID, err := google.CreateFolder(from)
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
			err = myaws.Upload(fileBytes, fileType, size, from, path)
		} else if format == ".tif" {
			err = google.Upload(fileBytes, path, folderID)
		}

		return err
	})
	if err != nil {
		return "", "", "", err
	}

	region, err := util.LoadEnvVar("AWS_REGION")
	if err != nil {
		return "", "", "", err
	}
	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return "", "", "", err
	}

	s3Url := fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/%s?region=%s&prefix=%s/", region, bucket, region, folder)
	driveUrl := fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", folderID)

	message := fmt.Sprintf("Finished uploading **%s** (%d new photos)", folder, count)

	err = os.RemoveAll(from)
	if err != nil {
		log.Default().Printf("Failed to remove directory: %v", err)
	}

	err = os.Remove(zip)
	if err != nil {
		log.Default().Printf("Failed to remove zip file: %v", err)

	}

	return s3Url, driveUrl, message, nil
}
