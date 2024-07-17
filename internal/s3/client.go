package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/sync/errgroup"

	"github.com/Artistichek/imaging/internal/s3/types"
)

type APIClient interface {
	UploadImages(parentCtx context.Context, objects []types.PutObjectInput) error
	UploadImage(ctx context.Context, object types.PutObjectInput) error
	DeleteImage(ctx context.Context, object types.DeleteObjectInput) error
	DeleteImages(ctx context.Context, objects types.DeleteObjectsInput) error

	Cfg() *Config
}

type Client struct {
	c *s3.Client

	cfg *Config
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentialsProvider(
		cfg.Credentials.KeyId,
		cfg.Credentials.Secret,
		cfg.Credentials.Session,
	)

	c := &Client{
		c: s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.Credentials = creds
			o.Region = cfg.EndpointResolver.Region
			o.BaseEndpoint = &cfg.EndpointResolver.BaseEndpoint
			o.UsePathStyle = cfg.EndpointResolver.HostnameImmutable
		}),
		cfg: cfg,
	}

	return c, nil
}

// UploadImages allows upload several images to s3 object storage,
// if any upload process is finished with error or upload timeout exceeded,
// method returns non-nil error.
func (c *Client) UploadImages(parentCtx context.Context, objects []types.PutObjectInput) error {
	// childCtx и parentCtx для отлавливания таймаутов на загрузку и общую обработку.
	// uploadCtx для отлавливания ошибок загрузки изображений в s3.
	childCtx, cancel := context.WithTimeout(parentCtx, c.cfg.UploadTimeout)
	eg, uploadCtx := errgroup.WithContext(childCtx)
	defer cancel()

	done := make(chan error, 1)
	defer close(done)

	for _, o := range objects {
		eg.Go(func() error {
			err := c.UploadImage(uploadCtx, o)
			return err
		})
	}

	select {
	case done <- eg.Wait():
		return <-done
	case <-childCtx.Done():
		return &UploadTimeoutError{c.cfg.UploadTimeout}
	case <-parentCtx.Done():
		return parentCtx.Err()
	}
}

func (c *Client) UploadImage(ctx context.Context, object types.PutObjectInput) error {
	_, err := c.c.PutObject(ctx, object.ToS3ObjectInput())
	if err != nil {
		return &OperationError{
			Op:  UploadImage,
			Err: err,
		}
	}
	return err
}

func (c *Client) DeleteImage(ctx context.Context, object types.DeleteObjectInput) error {
	_, err := c.c.DeleteObject(ctx, object.ToS3ObjectInput())
	if err != nil {
		err = &OperationError{
			Op:  DeleteImage,
			Err: err,
		}
	}
	return err
}

func (c *Client) DeleteImages(ctx context.Context, objects types.DeleteObjectsInput) error {
	_, err := c.c.DeleteObjects(ctx, objects.ToS3ObjectsInput())
	if err != nil {
		err = &OperationError{
			Op:  DeleteImages,
			Err: err,
		}
	}
	return err
}

func (c *Client) Cfg() *Config {
	return c.cfg
}
