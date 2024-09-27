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
		log.Default().Printf("[AWS S3] Uploaded %s!\n", path)
		return nil
	}
}

func SetFolderName(url string, name string) error {
	client := getClient()
	newFolder := name + "/"

	awsRegion := util.LoadEnvVar("AWS_REGION")
	a, _ := strings.CutPrefix(url, fmt.Sprintf("https://%s.console.aws.amazon.com/s3/buckets/", awsRegion))
	a = strings.ReplaceAll(a, fmt.Sprintf("?region=%s&prefix", awsRegion), "")
	bucketAndFolder := strings.Split(a, "=")
	bucketName := bucketAndFolder[0]
	oldFolder := bucketAndFolder[1]

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(oldFolder),
	}

	objects, err := client.ListObjectsV2(context.TODO(), params)
	if err != nil {
		return err
	}

	for _, object := range objects.Contents {
		oldKey := *object.Key
		newKey := strings.Replace(oldKey, oldFolder, newFolder, 1)

		_, err := client.CopyObject(context.TODO(), &s3.CopyObjectInput{
			Bucket:     aws.String(bucketName),
			CopySource: aws.String(bucketName + "/" + oldKey),
			Key:        aws.String(newKey),
		})
		if err != nil {
			return err
		}
	}

	for _, object := range objects.Contents {
		_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		})
		if err != nil {
			return err
		}
	}

	log.Default().Printf("[AWS S3] Set folder name to %s", name)
	return nil
}
