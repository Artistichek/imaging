package processor

import (
	"bytes"
	"context"
	"image"
	"testing"
	"time"

	"github.com/Artistichek/imaging/logs"

	testbase64 "github.com/Artistichek/imaging/test/base64"
	"github.com/Artistichek/imaging/test/errorcmp"
)

var timeoutErr *TimeoutError

// image.png 5x5
var valid = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAAAAEElEQVR42mP8X88ABYwkMAHtIQd8SvDAUAAAAABJRU5ErkJggg=="

var p = &Processor{
	cfg: &Config{
		Sizes:          []int{1, 2, 3, 4},
		EncodingFormat: "png",
		ProcessTimeout: 2 * time.Second,
	},
	log: logs.New(logs.ErrorLevel, logs.JSONOutput),
}

func TestProcessor_ProcessImage(t *testing.T) {
	var valid = testbase64.Decode(valid)

	testCases := []struct {
		name string

		want   error
		buffer []byte
	}{
		{
			name:   "ok",
			buffer: valid,
			want:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, got := p.ProcessImage(context.Background(), decodeImage(tc.buffer))

			if diff := errorcmp.Diff(got, tc.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func decodeImage(buff []byte) image.Image {
	img, _, _ := image.Decode(bytes.NewReader(buff))
	return img
}
