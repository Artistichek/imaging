package s3

import "time"

type Config struct {
	// Учетные данные для доступа к S3.
	Credentials Credentials // Тег нужен для того, чтобы не логировать конфиденциальные значения.

	// Настройки преобразования конечных точек для доступа к S3 API
	EndpointResolver EndpointResolver

	// Имя бакета.
	Bucket string

	// Директория в которой хранятся файлы в s3 хранилище.
	BaseDirectory string

	// Таймаут загрузки изображений в хранилище.
	UploadTimeout time.Duration
}

type Credentials struct {
	// AWS Access Key ID.
	KeyId string

	// AWS Secret Access Key.
	Secret string

	// AWS Session Token.
	Session string
}

type EndpointResolver struct {
	// Базовое имя хоста.
	BaseEndpoint string

	// Параметр задает, необходимо ли изменять конечную точку при выполнении сервисных операций.
	HostnameImmutable bool

	// Регион доступа к S3 API.
	Region string
}
