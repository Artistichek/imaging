package resizer

import (
	"context"
	"image"
	"sync"

	"github.com/anthonynsimon/bild/transform"
)

func ResizeImage(ctx context.Context, img image.Image, sizes []int) ([]image.Image, error) {
	var images = make([]image.Image, len(sizes))
	var resized = make(chan struct{})
	var wg = sync.WaitGroup{}

	go func() {
		defer close(resized)

		for i, size := range sizes {
			wg.Add(1)

			go func() {
				defer wg.Done()
				images[i] = Resize(img, size)
			}()
		}

		wg.Wait()
		resized <- struct{}{}
	}()

	select {
	case <-resized:
		return images, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func Resize(img image.Image, size int) image.Image {
	w, h := (img).Bounds().Dx(), (img).Bounds().Dy()
	if maxBound(w, h) < size {
		return img
	}

	w, h = adjustBounds(w, h, size)
	img = transform.Resize(img, w, h, transform.Lanczos)

	return img
}

func adjustBounds(w, h, target int) (int, int) {
	ratio := float32(w) / float32(h)

	switch {
	case ratio > 1:
		w, h = target, int(float32(target)/ratio)
	case ratio < 1:
		w, h = int(float32(target)*ratio), target
	default:
		w, h = target, target
	}

	return w, h
}

func maxBound(w, h int) int {
	if w > h {
		return w
	}
	return h
}
