// Package minio provides a MinIO client wrapper for the Offline Material Service.
// It adapts the generic pkg/minio package to the Offline Service's interface.
package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// Config holds MinIO client configuration for the Offline Service.
type Config struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	BucketMaterials string // Source bucket for original materials
	BucketEncrypted string // Destination bucket for encrypted materials
}

// Client wraps the MinIO client with Offline Service-specific operations.
type Client struct {
	client          *minio.Client
	bucketMaterials string
	bucketEncrypted string
}

// NewClient creates a new MinIO client for the Offline Service.
func NewClient(cfg Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{
		client:          client,
		bucketMaterials: cfg.BucketMaterials,
		bucketEncrypted: cfg.BucketEncrypted,
	}, nil
}

// EnsureBuckets ensures both materials and encrypted buckets exist.
func (c *Client) EnsureBuckets(ctx context.Context) error {
	buckets := []string{c.bucketMaterials, c.bucketEncrypted}

	for _, bucket := range buckets {
		exists, err := c.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to check bucket existence for %s: %w", bucket, err)
		}

		if !exists {
			err = c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
			log.Info().Str("bucket", bucket).Msg("created bucket")
		}
	}

	return nil
}

// GetObject retrieves an object from the materials bucket.
func (c *Client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	// Try encrypted bucket first, then materials bucket
	obj, err := c.client.GetObject(ctx, c.bucketEncrypted, objectKey, minio.GetObjectOptions{})
	if err == nil {
		// Check if object exists by trying to get stat
		_, statErr := obj.Stat()
		if statErr == nil {
			return obj, nil
		}
		obj.Close()
	}

	// Try materials bucket
	obj, err = c.client.GetObject(ctx, c.bucketMaterials, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get object", err)
	}

	// Verify object exists
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "object not found: "+objectKey)
		}
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to stat object", err)
	}

	return obj, nil
}

// GetObjectFromBucket retrieves an object from a specific bucket.
func (c *Client) GetObjectFromBucket(ctx context.Context, bucket, objectKey string) (io.ReadCloser, error) {
	obj, err := c.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get object", err)
	}

	// Verify object exists
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "object not found: "+objectKey)
		}
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to stat object", err)
	}

	return obj, nil
}

// PutObject stores an object in the encrypted bucket.
func (c *Client) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucketEncrypted, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to put object", err)
	}
	return nil
}

// DeleteObject deletes an object from the encrypted bucket.
func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	err := c.client.RemoveObject(ctx, c.bucketEncrypted, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to delete object", err)
	}
	return nil
}

// DeleteObjects deletes multiple objects from the encrypted bucket.
func (c *Client) DeleteObjects(ctx context.Context, objectKeys []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(objectKeys))
	go func() {
		defer close(objectsCh)
		for _, key := range objectKeys {
			objectsCh <- minio.ObjectInfo{Key: key}
		}
	}()

	for err := range c.client.RemoveObjects(ctx, c.bucketEncrypted, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return domain.WrapOfflineError(domain.ErrCodeStorageError,
				fmt.Sprintf("failed to delete object %s", err.ObjectName), err.Err)
		}
	}
	return nil
}

// GetObjectInfo returns metadata about an object.
func (c *Client) GetObjectInfo(ctx context.Context, objectKey string) (*application.ObjectInfo, error) {
	// Try materials bucket first
	stat, err := c.client.StatObject(ctx, c.bucketMaterials, objectKey, minio.StatObjectOptions{})
	if err != nil {
		// Try encrypted bucket
		stat, err = c.client.StatObject(ctx, c.bucketEncrypted, objectKey, minio.StatObjectOptions{})
		if err != nil {
			errResponse := minio.ToErrorResponse(err)
			if errResponse.Code == "NoSuchKey" {
				return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "object not found: "+objectKey)
			}
			return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get object info", err)
		}
	}

	return &application.ObjectInfo{
		Size:        stat.Size,
		ContentType: stat.ContentType,
		ETag:        stat.ETag,
	}, nil
}

// GeneratePresignedGetURL generates a presigned URL for downloading from encrypted bucket.
func (c *Client) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucketEncrypted, objectKey, expiry, nil)
	if err != nil {
		return "", domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to generate presigned URL", err)
	}
	return presignedURL.String(), nil
}

// GeneratePresignedGetURLFromMaterials generates a presigned URL for downloading from materials bucket.
func (c *Client) GeneratePresignedGetURLFromMaterials(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucketMaterials, objectKey, expiry, nil)
	if err != nil {
		return "", domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to generate presigned URL", err)
	}
	return presignedURL.String(), nil
}

// FileExists checks if a file exists in either bucket.
func (c *Client) FileExists(ctx context.Context, objectKey string) (bool, error) {
	// Check materials bucket
	_, err := c.client.StatObject(ctx, c.bucketMaterials, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}

	// Check encrypted bucket
	_, err = c.client.StatObject(ctx, c.bucketEncrypted, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}

	errResponse := minio.ToErrorResponse(err)
	if errResponse.Code == "NoSuchKey" {
		return false, nil
	}

	return false, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to check file existence", err)
}

// GetBucketMaterials returns the materials bucket name.
func (c *Client) GetBucketMaterials() string {
	return c.bucketMaterials
}

// GetBucketEncrypted returns the encrypted bucket name.
func (c *Client) GetBucketEncrypted() string {
	return c.bucketEncrypted
}

// HealthCheck checks if MinIO is accessible.
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.BucketExists(ctx, c.bucketMaterials)
	if err != nil {
		return fmt.Errorf("MinIO health check failed: %w", err)
	}
	return nil
}

// Verify that Client implements application.MinIOStorageClient interface.
var _ application.MinIOStorageClient = (*Client)(nil)
