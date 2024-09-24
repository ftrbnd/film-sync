package files

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ftrbnd/film-sync/internal/discord"
	"github.com/ftrbnd/film-sync/internal/util"
)

func s3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-1"))
	util.CheckError("Failed to load AWS config", err)

	client := s3.NewFromConfig(cfg)

	return client
}

func Upload(from string, zip string, count int) {
	folder := strings.ReplaceAll(filepath.Base(zip), ".zip", "")

	client := s3Client()

	_, err := os.ReadDir(from)
	util.CheckError("Failed to read directory", err)

	err = filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		return uploadPNG(path, d, err, client, folder)
	})
	util.CheckError("Failed to walk through directory", err)

	url := fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/%s?region=%s&prefix=%s/", util.LoadEnvVar("AWS_REGION"), util.LoadEnvVar("AWS_BUCKET_NAME"), util.LoadEnvVar("AWS_REGION"), folder)
	message := fmt.Sprintf("Finished uploading %d new photos to [%s](%s)", count, folder, url)
	discord.SendMessage(message)

	err = os.RemoveAll(from)
	util.CheckError("Failed to remove directory", err)
	err = os.Remove(zip)
	util.CheckError("Failed to remove zip file", err)
}

func uploadPNG(path string, d fs.DirEntry, err error, client *s3.Client, dst string) error {
	if err != nil {
		return err
	}

	if d.IsDir() || !strings.HasSuffix(path, "png") {
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
	params := &s3.PutObjectInput{
		Bucket:        aws.String(util.LoadEnvVar("AWS_BUCKET_NAME")),
		Key:           aws.String(fmt.Sprintf("%s/%s", dst, filepath.Base(path))),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	resp, err := client.PutObject(context.TODO(), params)
	if err != nil {
		return err
	}
	if resp != nil {
		log.Default().Printf("Uploaded %s to S3!\n", path)
	}

	return nil
}
