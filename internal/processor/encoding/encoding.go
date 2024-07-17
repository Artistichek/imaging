package encoding

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/kolesa-team/go-webp/webp"
)

type Error struct {
	fmt string
	err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("encoding error: format=%s, msg=%v", e.fmt, e.err)
}

func EncodeImage(img image.Image, format string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var err error

	switch format {
	case "webp":
		err = webp.Encode(&buf, img, nil)
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	}

	if err != nil {
		return buf, &Error{
			fmt: format,
			err: err,
		}
	}

	return buf, err
}
