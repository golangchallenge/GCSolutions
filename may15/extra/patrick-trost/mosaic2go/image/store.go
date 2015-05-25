package image

import (
	"fmt"
	"io"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

// CloudFileStorage defines an interface for cloud file storage.
type CloudFileStorage interface {
	Save(data io.ReadSeeker, filename string) (string, error)
	Delete(filename string) error
}

// S3FileStorage implements the CloudFileStorage interface and provides methods for storing
// and deleting files form Amanon's S3 service.
type S3FileStorage struct {
	region, bucket string
}

// Save stores a files with the given name at a S3 bucket.
func (s S3FileStorage) Save(data io.ReadSeeker, filename string) (string, error) {
	cred := aws.DefaultChainCredentials
	cred.Get()
	svc := s3.New(&aws.Config{Region: s.region, Credentials: cred, LogLevel: 0})
	params := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		Body:   data,
	}
	_, err := svc.PutObject(params)
	if err != nil {
		return "", err
	}

	return s.absFilepath(filename), nil
}

func (s S3FileStorage) absFilepath(filename string) string {
	return fmt.Sprintf("http://%s.s3-%s.amazonaws.com/%s", s.bucket, s.region, filename)
}

// Delete stores a files with the given name at a S3 bucket.
func (s S3FileStorage) Delete(filename string) error {
	cred := aws.DefaultChainCredentials
	svc := s3.New(&aws.Config{Region: s.region, Credentials: cred, LogLevel: 0})
	params := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &s3.Delete{
			Objects: []*s3.ObjectIdentifier{
				&s3.ObjectIdentifier{
					Key: &filename,
				},
			},
			Quiet: aws.Boolean(true),
		},
	}
	_, err := svc.DeleteObjects(params)

	if awserr := aws.Error(err); awserr != nil {
		return err
	} else if err != nil {
		// A non-service error occurred.
		panic(err)
	}

	return nil
}

// NewS3FileStorage creates an new S3FileStorage pointer.
func NewS3FileStorage(region, bucket string) S3FileStorage {
	return S3FileStorage{region, bucket}
}
