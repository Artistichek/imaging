package decoding

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

type Error struct {
	fmt string
	err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("unknown decoding error: image format=%s, msg=%v", e.fmt, e.err)
}

// DecodeImage работает с изображениями формата: png, gif, jpeg.
func DecodeImage(r io.Reader) (image.Image, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		if format == "" {
			err = image.ErrFormat
		} else {
			err = &Error{fmt: format, err: err}
		}
	}

	return img, err
}
