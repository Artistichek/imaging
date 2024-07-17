package resizer

import (
	"image"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResizeImage(t *testing.T) {
	testCases := []struct {
		name string

		target int
		img    image.Image
		want   image.Image
	}{
		{
			name:   "empty",
			target: 1,
			img:    image.NewRGBA(image.Rect(0, 0, 0, 0)),
			want:   image.NewRGBA(image.Rect(0, 0, 0, 0)),
		},
		{
			name:   "x1",
			target: 1,
			img:    image.NewRGBA(image.Rect(0, 0, 1, 1)),
			want:   image.NewRGBA(image.Rect(0, 0, 1, 1)),
		},
		{
			name:   "x2",
			target: 2,
			img:    image.NewRGBA(image.Rect(0, 0, 1, 1)),
			want:   image.NewRGBA(image.Rect(0, 0, 1, 1)),
		},
		{
			name:   "x0.5",
			target: 1,
			img:    image.NewRGBA(image.Rect(0, 0, 2, 2)),
			want:   image.NewRGBA(image.Rect(0, 0, 1, 1)),
		},
		{
			name:   "horizontal x0.5",
			target: 2,
			img:    image.NewRGBA(image.Rect(0, 0, 10, 5)),
			want:   image.NewRGBA(image.Rect(0, 0, 2, 1)),
		},
		{
			name:   "vertical x0.5",
			target: 2,
			img:    image.NewRGBA(image.Rect(0, 0, 5, 10)),
			want:   image.NewRGBA(image.Rect(0, 0, 1, 2)),
		},
		{
			name:   "bad target: horizontal x0.5",
			target: 1,
			img:    image.NewRGBA(image.Rect(0, 0, 2, 1)),
			want:   image.NewRGBA(image.Rect(0, 0, 0, 0)),
		},
		{
			name:   "bad target: vertical x0.5",
			target: 1,
			img:    image.NewRGBA(image.Rect(0, 0, 1, 2)),
			want:   image.NewRGBA(image.Rect(0, 0, 0, 0)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Resize(tc.img, tc.target)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}
