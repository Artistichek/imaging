package s3

import (
	"fmt"
	"time"
)

type OperationType string

const (
	UploadImage  OperationType = "UploadImage"
	DeleteImage  OperationType = "DeleteImage"
	DeleteImages OperationType = "DeleteImages"
)

type OperationError struct {
	Op  OperationType
	Err error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("operation '%s' failed, msg=%v", e.Op, e.Err)
}

type UploadTimeoutError struct {
	timeout time.Duration
}

func (u *UploadTimeoutError) Error() string {
	return fmt.Sprintf("upload timeout error: upload was timed out after: %v", u.timeout)
}
