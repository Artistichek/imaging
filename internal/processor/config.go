package processor

import (
	"time"
)

type Config struct {
	// Максимальные размеры изображения.
	Sizes []int

	// Параметры кодировщика изображений.
	EncodingFormat string

	// Таймаут обработки изображения.
	ProcessTimeout time.Duration
}
