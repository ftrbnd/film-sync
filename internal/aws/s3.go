package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ftrbnd/film-sync/internal/util"
)

func getClient() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return client, nil
}

func Upload(bytes *bytes.Reader, fileType string, size int64, dst string, path string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return err
	}

	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(fmt.Sprintf("%s/%s", dst, filepath.Base(path))),
		Body:          bytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	_, err = client.PutObject(context.TODO(), params)
	if err != nil {
		return err
	}

	log.Default().Printf("[AWS S3] Uploaded %s!\n", path)
	return nil
}

func SetFolderName(old string, new string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	newFolder := new + "/"
	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return err
	}

	objects, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(old),
	})
	if err != nil {
		return err
	}

	for _, object := range objects.Contents {
		oldKey := *object.Key
		newKey := strings.Replace(oldKey, old, newFolder, 1)

		_, err := client.CopyObject(context.TODO(), &s3.CopyObjectInput{
			Bucket:     aws.String(bucket),
			CopySource: aws.String(bucket + "/" + oldKey),
			Key:        aws.String(newKey),
		})
		if err != nil {
			return err
		}
	}

	for _, object := range objects.Contents {
		_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    object.Key,
		})
		if err != nil {
			return err
		}
	}

	log.Default().Printf("[AWS S3] Set folder name to %s", new)
	return nil
}
