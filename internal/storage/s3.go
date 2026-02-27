package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 stores files in an S3-compatible bucket and serves via presigned URL redirects.
type S3 struct {
	client *s3.Client
	signer *s3.PresignClient
	bucket string
	prefix string // key prefix inside the bucket
}

// S3Config holds configuration for S3 storage.
type S3Config struct {
	Bucket    string
	Region    string
	Endpoint  string // custom endpoint for MinIO, DO Spaces, etc.
	AccessKey string
	SecretKey string
	KeyPrefix string // optional key prefix (e.g. "uploads/")
}

func NewS3(ctx context.Context, cfg S3Config) (*S3, error) {
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(cfg.Region))

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	return &S3{
		client: client,
		signer: s3.NewPresignClient(client),
		bucket: cfg.Bucket,
		prefix: cfg.KeyPrefix,
	}, nil
}

func (s *S3) Save(ctx context.Context, filename string, r io.Reader) (string, error) {
	key := s.prefix + filename
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return "", fmt.Errorf("s3 put: %w", err)
	}
	// Return a path that our Handler() will intercept and presign
	return "/uploads/" + filename, nil
}

// Handler returns a handler that redirects to presigned S3 URLs.
func (s *S3) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract filename from path: /uploads/filename.pdf
		filename := r.URL.Path
		if len(filename) > 0 && filename[0] == '/' {
			filename = filename[1:]
		}

		key := s.prefix + filename
		presigned, err := s.signer.PresignGetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(15*time.Minute))
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, presigned.URL, http.StatusTemporaryRedirect)
	})
}
