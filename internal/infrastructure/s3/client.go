package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.ObjectStore = (*Client)(nil)

type Client struct {
	mc  *minio.Client
	cfg config.S3Config
}

func New(ctx context.Context, cfg config.S3Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
		if resp := minio.ToErrorResponse(err); resp.Code != "BucketAlreadyExists" && resp.Code != "BucketAlreadyOwnedByYou" {
			return nil, fmt.Errorf("ensure bucket: %w", err)
		}
	}

	return &Client{mc: mc, cfg: cfg}, nil
}

func (c *Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := c.mc.PutObject(ctx, c.cfg.Bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put object %s: %w", key, err)
	}
	return fmt.Sprintf("%s/%s/%s", c.publicBase(), c.cfg.Bucket, key), nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.mc.RemoveObject(ctx, c.cfg.Bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object %s: %w", key, err)
	}
	return nil
}

func (c *Client) publicBase() string {
	pub := c.cfg.PublicEndpoint
	if pub == "" {
		pub = c.cfg.Endpoint
	}
	scheme := "http"
	if c.cfg.UseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, pub)
}
