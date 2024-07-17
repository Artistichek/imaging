package types

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PutObjectInput struct {
	Bucket   string
	Key      *ObjectKey
	Body     []byte
	Metadata string
}

func (p PutObjectInput) ToS3ObjectInput() *s3.PutObjectInput {
	format := extractFormat(p.Key.String())
	contentType := fmt.Sprintf("image/%s; color=%s", format, p.Metadata)

	return &s3.PutObjectInput{
		Bucket:      aws.String(p.Bucket),
		Key:         aws.String(p.Key.String()),
		Body:        bytes.NewReader(p.Body),
		ContentType: aws.String(contentType),
	}
}

// extractFormat извлекает формат изображения исходя из пути "{base_directory}/{game_id}/{image_name}.{format}".
// Если в пути нет формата, то возвращается пустая строка.
func extractFormat(key string) string {
	chars := make([]byte, 0)
	for i := len(key) - 1; i >= 0 && key[i] != '.'; i-- {
		chars = append(chars, key[i])
	}

	if len(chars) == len(key) {
		return ""
	}

	for i := 0; i < len(chars)/2; i++ {
		chars[i], chars[len(chars)-1-i] = chars[len(chars)-1-i], chars[i]
	}

	return string(chars)
}

type DeleteObjectInput struct {
	Bucket string
	Key    *ObjectKey
}

func (d DeleteObjectInput) ToS3ObjectInput() *s3.DeleteObjectInput {
	return &s3.DeleteObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(d.Key.String()),
	}
}

type DeleteObjectsInput struct {
	Bucket string
	Keys   []*ObjectKey
}

func (d DeleteObjectsInput) ToS3ObjectsInput() *s3.DeleteObjectsInput {
	objectIds := make([]types.ObjectIdentifier, 0)
	for _, key := range d.Keys {
		id := types.ObjectIdentifier{
			Key: aws.String(key.String()),
		}

		objectIds = append(objectIds, id)
	}

	return &s3.DeleteObjectsInput{
		Bucket: aws.String(d.Bucket),
		Delete: &types.Delete{
			Objects: objectIds,
		},
	}
}
