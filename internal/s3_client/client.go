package s3_client

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client struct {
	sess *session.Session
}

func NewClient(AccessKeyID string, SecretAccessKey string, MyRegion string) S3Client {
	return S3Client{
		sess: getAwsSession(AccessKeyID, SecretAccessKey, MyRegion),
	}
}

func (client *S3Client) UploadData(bucketName string, keyString string, data []byte) {
	uploader := s3manager.NewUploader(client.sess)

	input := &s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyString),
		Body:   bytes.NewReader(data),
	}

	result, err := uploader.Upload(input)
	if err != nil {
		log.Fatalf("failed to upload file, %e!", err)
	}

	log.Printf("Upload Success: %v \n", result)
}

func (client *S3Client) DownloadData(bucketName string, keyString string) []byte {
	bufferWriter := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(client.sess)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyString),
	}

	_, err := downloader.Download(bufferWriter, input)
	if err != nil {
		log.Printf("failed to download data %s/%s: %s", bucketName, keyString, err)
	}

	return bufferWriter.Bytes()
}

func (client *S3Client) ListObjects(bucketName string, prefix string) []string {
	svc := s3.New(client.sess)

	res, i := make([]string, 0, 100), 0

	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
		Prefix: &prefix,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, obj := range p.Contents {
			res = append(res, *obj.Key)
			i++
		}
		return true
	})

	if err != nil {
		log.Fatal(err)
	}

	return res[:i]
}

func (client *S3Client) GetSession() *session.Session {
	return client.sess
}
