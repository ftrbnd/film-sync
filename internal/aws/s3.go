package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ftrbnd/film-sync/internal/util"
)

func getClient() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-1"))
	util.CheckError("Failed to load AWS config", err)

	client := s3.NewFromConfig(cfg)

	return client
}

func Upload(bytes *bytes.Reader, fileType string, size int64, dst string, path string) error {
	client := getClient()

	params := &s3.PutObjectInput{
		Bucket:        aws.String(util.LoadEnvVar("AWS_BUCKET_NAME")),
		Key:           aws.String(fmt.Sprintf("%s/%s", dst, filepath.Base(path))),
		Body:          bytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	_, err := client.PutObject(context.TODO(), params)
	if err != nil {
		return err
	} else {
		log.Default().Printf("Uploaded %s to S3!\n", path)
		return nil
	}
}
