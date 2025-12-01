// Package minio provides a MinIO client wrapper for the Material Service.
// It adapts the generic pkg/minio package to the Material Service's interface.
package minio

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"ngasihtau/internal/material/application"
	pkgminio "ngasihtau/pkg/minio"
)

// Config holds MinIO client configuration for the Material Service.
type Config struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	BucketMaterials string
}

// Client wraps the pkg/minio client with Material Service-specific operations.
type Client struct {
	client *pkgminio.Client
}

// NewClient creates a new MinIO client for the Material Service.
func NewClient(cfg Config) (*Client, error) {
	client, err := pkgminio.New(pkgminio.Config{
		Endpoint:  cfg.Endpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		UseSSL:    cfg.UseSSL,
		Bucket:    cfg.BucketMaterials,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

// EnsureBucket ensures the materials bucket exists.
func (c *Client) EnsureBucket(ctx context.Context) error {
	if err := c.client.EnsureBucket(ctx); err != nil {
		return err
	}
	log.Info().Str("bucket", c.client.GetBucket()).Msg("materials bucket ready")
	return nil
}

// GeneratePresignedPutURL generates a presigned URL for uploading a file.
func (c *Client) GeneratePresignedPutURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	return c.client.GeneratePresignedPutURL(ctx, objectKey, contentType, expiry)
}

// GeneratePresignedGetURL generates a presigned URL for downloading a file.
func (c *Client) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return c.client.GeneratePresignedGetURL(ctx, objectKey, expiry)
}

// FileExists checks if a file exists in the bucket.
func (c *Client) FileExists(ctx context.Context, objectKey string) (bool, error) {
	return c.client.FileExists(ctx, objectKey)
}

// GetFileInfo returns file information (size, content type).
func (c *Client) GetFileInfo(ctx context.Context, objectKey string) (*application.FileInfo, error) {
	info, err := c.client.GetFileInfo(ctx, objectKey)
	if err != nil {
		return nil, err
	}

	return &application.FileInfo{
		Size:        info.Size,
		ContentType: info.ContentType,
		ETag:        info.ETag,
	}, nil
}

// DeleteFile deletes a file from the bucket.
func (c *Client) DeleteFile(ctx context.Context, objectKey string) error {
	return c.client.DeleteFile(ctx, objectKey)
}

// GetBucketName returns the bucket name.
func (c *Client) GetBucketName() string {
	return c.client.GetBucket()
}

// Verify that Client implements application.MinIOClient interface.
var _ application.MinIOClient = (*Client)(nil)
