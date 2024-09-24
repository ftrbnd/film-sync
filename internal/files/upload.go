package files

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	myaws "github.com/ftrbnd/film-sync/internal/aws"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/google"
	"github.com/ftrbnd/film-sync/internal/util"
	"google.golang.org/api/drive/v3"
)

func Upload(from string, zip string, count int, drive *drive.Service) {
	folder := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	_, err := os.ReadDir(from)
	util.CheckError("Failed to read directory", err)

	folderID, driveLink := google.CreateFolder(drive, from)
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
			err = google.Upload(fileBytes, path, folderID, drive)
		}

		return err
	})
	util.CheckError("Failed to walk through directory", err)

	s3Url := fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/%s?region=%s&prefix=%s/", util.LoadEnvVar("AWS_REGION"), util.LoadEnvVar("AWS_BUCKET_NAME"), util.LoadEnvVar("AWS_REGION"), folder)
	message := fmt.Sprintf("Finished uploading **%s** (%d new photos) to [AWS S3](%s) and [Google Drive](%s)", folder, count, s3Url, driveLink)
	discord.SendMessage(message)

	err = os.RemoveAll(from)
	util.CheckError("Failed to remove directory", err)
	err = os.Remove(zip)
	util.CheckError("Failed to remove zip file", err)
}
