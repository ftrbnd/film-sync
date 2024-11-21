package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ftrbnd/film-sync/internal/util"
)

var client *s3.Client

func StartClient() error {
	region, err := util.LoadEnvVar("AWS_REGION")
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config %v", err)
	}

	client = s3.NewFromConfig(cfg)
	log.Default().Println("[AWS] Successfully started S3 client")
	return nil
}

func Upload(bytes *bytes.Reader, fileType string, size int64, dst string, path string) (string, error) {
	if client == nil {
		return "", errors.New("AWS client hasn't been initialized")
	}

	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/%s", dst, filepath.Base(path))
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	_, err = client.PutObject(context.TODO(), params)
	if err != nil {
		return "", err
	}

	log.Default().Printf("[AWS S3] Uploaded %s!\n", path)
	return key, nil
}

func SetFolderName(old string, new string) error {
	if client == nil {
		return errors.New("AWS client hasn't been initialized")
	}

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
	if len(objects.Contents) == 0 {
		log.Default().Printf("[AWS S3] %s folder has no contents", old)
		return nil
	}

	for _, object := range objects.Contents {
		oldKey := *object.Key
		newKey := strings.Replace(oldKey, old, new, 1)

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

func FolderLink(prefix string) (string, error) {
	region, err := util.LoadEnvVar("AWS_REGION")
	if err != nil {
		return "", err
	}
	bucket, err := util.LoadEnvVar("AWS_BUCKET_NAME")
	if err != nil {
		return "", err
	}

	p := strings.ReplaceAll(prefix, " ", "+")
	url := fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/%s?region=%s&prefix=%s/", region, bucket, region, p)

	return url, nil
}
