package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"os"
)

type DownloadedFile struct {
	Content io.ReadCloser
	Info    minio.ObjectInfo
}

type Client interface {
	Connect() error
	UploadFile(ctx context.Context, file io.Reader, objectName string, contentType string) (string, error)
	IsConfigured() bool
	DownloadFile(ctx context.Context, objectName string) (*DownloadedFile, error)
}

type minioClient struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewClient() Client {
	return &minioClient{}
}

func (mc *minioClient) Connect() error {
	if os.Getenv("MINIO_ENABLED") == "true" {
		log.Println("MINIO_ENABLED is true. Connecting to MinIO...")
	} else {
		log.Println("WARNING: MINIO_ENABLED is not 'true'. MinIO client will not connect.")
	}
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return fmt.Errorf("one or more required MinIO environment variables are missing (ENDPOINT, ACCESS_KEY, SECRET_KEY, BUCKET_NAME)")
	}

	minioLibClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO library client: %w", err)
	}

	mc.client = minioLibClient
	mc.bucketName = bucketName
	mc.endpoint = endpoint
	mc.useSSL = useSSL

	ctx := context.Background()
	exists, err := mc.client.BucketExists(ctx, mc.bucketName)
	if err != nil {
		return fmt.Errorf("error checking if bucket '%s' exists: %w", mc.bucketName, err)
	}

	if !exists {
		log.Printf("Bucket '%s' does not exist. Creating it now...", mc.bucketName)
		if err = mc.client.MakeBucket(ctx, mc.bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket '%s': %w", mc.bucketName, err)
		}
	}

	log.Printf("MinIO client initialized and connected successfully for bucket '%s'", mc.bucketName)
	return nil
}

func (mc *minioClient) UploadFile(ctx context.Context, file io.Reader, objectName string, contentType string) (string, error) {
	if !mc.IsConfigured() {
		return "", fmt.Errorf("cannot upload file: MinIO client is not configured or connected")
	}

	_, err := mc.client.PutObject(ctx, mc.bucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error uploading file to MinIO: %w", err)
	}

	protocol := "http"
	if mc.useSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, mc.endpoint, mc.bucketName, objectName)

	return url, nil
}

func (mc *minioClient) DownloadFile(ctx context.Context, objectName string) (*DownloadedFile, error) {
	if !mc.IsConfigured() {
		return nil, fmt.Errorf("cannot download file: MinIO client is not configured or connected")
	}

	objInfo, err := mc.client.StatObject(ctx, mc.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for object '%s': %w", objectName, err)
	}

	object, err := mc.client.GetObject(ctx, mc.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object '%s': %w", objectName, err)
	}

	return &DownloadedFile{
		Content: object,
		Info:    objInfo,
	}, nil
}

func (mc *minioClient) IsConfigured() bool {
	return mc.client != nil
}
