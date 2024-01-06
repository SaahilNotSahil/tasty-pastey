package fileupload

import (
	"context"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

var BucketName string
var Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	Client = s3.NewFromConfig(cfg)

	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	BucketName = os.Getenv("BUCKET_NAME")
}

func UploadFile(key string, file *multipart.File) error {
	_, err := Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
		Body:   *file,
	})

	return err
}

func GetFile(key string) (io.ReadCloser, int64, error) {
	result, err := Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Println(err)

		return nil, 0, err
	}

	return result.Body, *result.ContentLength, nil
}

func DeleteFile(key string) error {
	_, err := Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Println(err)

		return err
	}

	return nil
}
