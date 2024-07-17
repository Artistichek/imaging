package dominator

import (
	"context"
	"errors"
	"image"
	"image/color"

	"github.com/cenkalti/dominantcolor"
)

var ErrColorExtraction = errors.New("error on extract dominant color: no colors found")

func GetDominantColor(ctx context.Context, img image.Image) (string, error) {
	var colors = make([]color.RGBA, 3)
	var extracted = make(chan struct{})
	var err error

	go func() {
		defer close(extracted)

		if colors = dominantcolor.FindN(img, 3); len(colors) == 0 {
			err = ErrColorExtraction
		}

		extracted <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-extracted:
		return dominantcolor.Hex(colors[0]), err
	}
}
