// AWS Connection module
package aws

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var (
	scv       *AWS
	singleton sync.Once
)

type Config struct {
	Address string
	Region  string
	Profile string
	ID      string
	Secret  string
	Bucket  string
}

// Load config from env
func LoadConfig() Config {
	return Config{
		Address: os.Getenv("AWS_S3_ENDPOINT_URL"),
		Region:  os.Getenv("AWS_S3_REGION_NAME"),
		Profile: "localstack",
		ID:      os.Getenv("AWS_ACCESS_KEY_ID"),
		Secret:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Bucket:  os.Getenv("AWS_STORAGE_BUCKET_NAME"),
	}
}

type AWS struct {
	Session *session.Session
	Config  Config
}

// Creates a new connection to AWS
func New(config Config) (*AWS, error) {
	sess, err := session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Credentials:      credentials.NewStaticCredentials(config.ID, config.Secret, ""),
				Region:           aws.String(config.Region),
				Endpoint:         aws.String(config.Address),
				S3ForcePathStyle: aws.Bool(true),
			},
			Profile: config.Profile,
		},
	)

	if err != nil {
		return nil, err
	}
	return &AWS{
		Session: sess,
		Config:  config,
	}, nil
}

// Upload a file to the defauklt bucket
func (ins *AWS) UploadFile(file *multipart.FileHeader) string {
	// Create a new S3 service client
	svc := s3.New(ins.Session)

	// Open the file
	fileReader, err := file.Open()
	if err != nil {
		panic("File error")
	}

	newName := uuid.New().String() + filepath.Ext(file.Filename)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(ins.Config.Bucket),
		Key:    aws.String(newName),
		Body:   fileReader,
	})
	if err != nil {
		panic("Upload error")
	}

	url := strings.Join([]string{ins.Config.Address, ins.Config.Bucket, newName}, "/")
	return url
}

// Singleton for AWS connection
func GetConnection() *AWS {
	singleton.Do(func() {
		var err error
		scv, err = New(LoadConfig())
		if err != nil {
			panic("Something gone wrong while communicating with AWS")
		}
	})
	return scv
}
