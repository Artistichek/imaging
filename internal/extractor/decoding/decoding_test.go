package decoding

import (
	"bytes"
	"image"
	"testing"

	testbase64 "github.com/Artistichek/imaging/test/base64"
	"github.com/Artistichek/imaging/test/errorcmp"
)

var decodingError *Error

var (
	// image.png 5x5
	png = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAAAAEElEQVR42mP8X88ABYwkMAHtIQd8SvDAUAAAAABJRU5ErkJggg=="
	// image.jpeg 5x5
	jpeg = "/9j/4AAQSkZJRgABAQAAAAAAAAD/4QBiRXhpZgAATU0AKgAAAAgABQESAAMAAAABAAEAAAEaAAUAAAABAAAASgEbAAUAAAABAAAAUgEoAAMAAAABAAEAAAITAAMAAAABAAEAAAAAAAAAAAAAAAAAAQAAAAAAAAAB/9sAQwADAgICAgIDAgICAwMDAwQGBAQEBAQIBgYFBgkICgoJCAkJCgwPDAoLDgsJCQ0RDQ4PEBAREAoMEhMSEBMPEBAQ/8AACwgABQAFAQERAP/EABQAAQAAAAAAAAAAAAAAAAAAAAn/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oACAEBAAA/AFTf/9k="
	// image.gif 5x5
	gif = "R0lGODlhBQAFAIAAAAAAAAAAACH5BAUKAAAALAAAAAAFAAUAAAIEhI+pWAA7"
	// corrupted image 5x5
	corrupted = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAwAsEElEQVSz2mP8Yc8ARYxkMAHtIQd8SvCtugAAAAB4RU7drkJggg=="
)

func TestDecodeImage(t *testing.T) {
	testCases := []struct {
		name string

		file []byte
		want error
	}{
		{
			name: "ok png",
			file: testbase64.Decode(png),
			want: nil,
		},
		{
			name: "ok jpeg",
			file: testbase64.Decode(jpeg),
			want: nil,
		},
		{
			name: "ok gif",
			file: testbase64.Decode(gif),
			want: nil,
		},
		{
			name: "not image",
			file: []byte("not image"),
			want: image.ErrFormat,
		},
		{
			name: "corrupted",
			file: testbase64.Decode(corrupted),
			want: decodingError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, got := DecodeImage(bytes.NewReader(tc.file))

			if diff := errorcmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
