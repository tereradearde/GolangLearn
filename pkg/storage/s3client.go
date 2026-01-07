package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	client  *minio.Client
	bucket  string
	baseURL string // e.g. https://s3.twcstorage.ru
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	BaseURL   string // public base URL; if empty, derive from Endpoint
}

func NewS3Client(cfg S3Config) (*S3Client, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	base := cfg.BaseURL
	if base == "" {
		scheme := "https"
		if !cfg.UseSSL {
			scheme = "http"
		}
		base = fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)
	}
	return &S3Client{client: cli, bucket: cfg.Bucket, baseURL: base}, nil
}

func (s *S3Client) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	}
	return nil
}

func (s *S3Client) PresignPut(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	if contentType != "" {
		reqParams.Set("response-content-type", contentType)
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, key, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *S3Client) ObjectURL(key string) string {
	// path-style URL
	return fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucket, key)
}

func (s *S3Client) GenerateKey(prefix, ext string) string {
	if prefix == "" {
		prefix = "assets"
	}
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	id := uuid.New().String()
	now := time.Now()
	return fmt.Sprintf("%s/%04d/%02d/%s%s", prefix, now.Year(), now.Month(), id, ext)
}
