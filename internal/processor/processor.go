package processor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"sort"
	"time"

	"github.com/Artistichek/imaging/logs"

	"github.com/Artistichek/imaging/internal/processor/dominator"
	"github.com/Artistichek/imaging/internal/processor/encoding"
	"github.com/Artistichek/imaging/internal/processor/resizer"
)

type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("processing timeout error: image processing was timed out after: %v", e.Timeout)
}

type ImageProcessor interface {
	ProcessImage(ctx context.Context, img image.Image) ([]bytes.Buffer, string, error)
	Cfg() *Config
}

type Processor struct {
	cfg *Config

	log *logs.Logger
}

func New(ctx context.Context, cfg *Config) *Processor {
	log := logs.FromContext(ctx).With().
		Str("system", "processor").
		Logger()

	return &Processor{
		cfg: cfg,
		log: &log,
	}
}

// ProcessImage run image processing according next flow:
// get dominant color -> transform its to several sizes -> encode each image to webp format
// and returns encoded images with dominant color and error, if its exists.
func (p *Processor) ProcessImage(parentCtx context.Context, img image.Image) ([]bytes.Buffer, string, error) {
	// childCtx и parentCtx для отлавливания таймаутов на обработку изображения и общую обработку.
	childCtx, cancel := context.WithTimeout(parentCtx, p.cfg.ProcessTimeout)
	defer cancel()

	var err error

	var color string
	if color, err = dominator.GetDominantColor(childCtx, img); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			err = &TimeoutError{Timeout: p.cfg.ProcessTimeout}
		}

		return nil, "", err
	}

	var resized []image.Image
	if resized, err = resizer.ResizeImage(childCtx, img, p.cfg.Sizes); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			err = &TimeoutError{Timeout: p.cfg.ProcessTimeout}
		}

		return nil, color, err
	}

	sizes := make([][]int, len(resized))
	for i, img := range resized {
		sizes[i] = []int{img.Bounds().Dx(), img.Bounds().Dy()}
	}

	p.log.Info().
		Any("sizes", sizes).
		Msg("image resized")

	sort.SliceStable(resized, func(i, j int) bool {
		return maxBound(resized[i]) < maxBound(resized[j])
	})

	var buffered = make([]bytes.Buffer, len(resized))
	for i, img := range resized {
		if buffered[i], err = encoding.EncodeImage(img, p.cfg.EncodingFormat); err != nil {
			return nil, color, err
		}
	}

	select {
	case <-childCtx.Done():
		err = &TimeoutError{p.cfg.ProcessTimeout}
	case <-parentCtx.Done():
		err = parentCtx.Err()
	default:
		err = nil
	}

	return buffered, color, nil
}

func (p *Processor) Cfg() *Config {
	return p.cfg
}

func maxBound(img image.Image) int {
	if img.Bounds().Dx() > img.Bounds().Dy() {
		return img.Bounds().Dx()
	}
	return img.Bounds().Dy()
}
