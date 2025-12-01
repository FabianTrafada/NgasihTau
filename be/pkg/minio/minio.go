// Package minio provides a reusable MinIO client wrapper for object storage operations.
// This package can be used across different services for file upload, download,
// and management operations using presigned URLs.
package minio

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// Config holds MinIO client configuration.
type Config struct {
	Endpoint  string // MinIO server endpoint (e.g., "localhost:9000")
	AccessKey string // Access key for authentication
	SecretKey string // Secret key for authentication
	UseSSL    bool   // Whether to use SSL/TLS
	Bucket    string // Default bucket name
}

// FileInfo represents metadata about a stored file.
type FileInfo struct {
	Size         int64     // File size in bytes
	ContentType  string    // MIME type of the file
	ETag         string    // Entity tag for the file
	LastModified time.Time // Last modification time
}

// Client wraps the MinIO client with common operations.
type Client struct {
	client *minio.Client
	bucket string
}

// New creates a new MinIO client with the given configuration.
func New(cfg Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// EnsureBucket ensures the configured bucket exists, creating it if necessary.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", c.bucket).Msg("created bucket")
	}

	return nil
}

// GeneratePresignedPutURL generates a presigned URL for uploading a file.
// The URL is valid for the specified duration.
func (c *Client) GeneratePresignedPutURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, c.bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	// Add content-type header requirement if specified
	if contentType != "" {
		q := presignedURL.Query()
		q.Set("Content-Type", contentType)
		presignedURL.RawQuery = q.Encode()
	}

	return presignedURL.String(), nil
}

// GeneratePresignedGetURL generates a presigned URL for downloading a file.
// The URL is valid for the specified duration.
func (c *Client) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GeneratePresignedGetURLWithFilename generates a presigned URL for downloading a file
// with a custom filename in the Content-Disposition header.
func (c *Client) GeneratePresignedGetURLWithFilename(ctx context.Context, objectKey string, filename string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	return presignedURL.String(), nil
}

// FileExists checks if a file exists in the bucket.
func (c *Client) FileExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := c.client.StatObject(ctx, c.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

// GetFileInfo returns metadata about a file in the bucket.
func (c *Client) GetFileInfo(ctx context.Context, objectKey string) (*FileInfo, error) {
	stat, err := c.client.StatObject(ctx, c.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, fmt.Errorf("file not found: %s", objectKey)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		ETag:         stat.ETag,
		LastModified: stat.LastModified,
	}, nil
}

// DeleteFile deletes a file from the bucket.
func (c *Client) DeleteFile(ctx context.Context, objectKey string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteFiles deletes multiple files from the bucket.
func (c *Client) DeleteFiles(ctx context.Context, objectKeys []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(objectKeys))
	go func() {
		defer close(objectsCh)
		for _, key := range objectKeys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	for err := range c.client.RemoveObjects(ctx, c.bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return fmt.Errorf("failed to delete file %s: %w", err.ObjectName, err.Err)
		}
	}
	return nil
}

// CopyFile copies a file within the same bucket.
func (c *Client) CopyFile(ctx context.Context, srcKey, dstKey string) error {
	src := minio.CopySrcOptions{
		Bucket: c.bucket,
		Object: srcKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: c.bucket,
		Object: dstKey,
	}

	_, err := c.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

// GetBucket returns the configured bucket name.
func (c *Client) GetBucket() string {
	return c.bucket
}

// SetBucket changes the default bucket for subsequent operations.
func (c *Client) SetBucket(bucket string) {
	c.bucket = bucket
}

// BucketExists checks if a bucket exists.
func (c *Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return exists, nil
}

// CreateBucket creates a new bucket.
func (c *Client) CreateBucket(ctx context.Context, bucket string) error {
	err := c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// ListObjects lists objects in the bucket with the given prefix.
func (c *Client) ListObjects(ctx context.Context, prefix string, recursive bool) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}

	for object := range c.client.ListObjects(ctx, c.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// GetUnderlyingClient returns the underlying minio.Client for advanced operations.
// Use with caution - prefer using the wrapper methods when possible.
func (c *Client) GetUnderlyingClient() *minio.Client {
	return c.client
}
