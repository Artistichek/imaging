package dominator

import (
	"context"
	"image"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetDominantColor(t *testing.T) {
	ctx := context.TODO()
	testCases := []struct {
		name string

		img  image.Image
		want string
	}{
		{
			name: "100",
			img: &image.RGBA{
				Stride: 1 * 4,
				Rect:   image.Rect(0, 0, 1, 1),
				Pix:    []uint8{0xFF, 0xFF, 0xFF, 0xFF},
			},
			want: "#FFFFFF",
		},
		{
			name: "66/33",
			img: &image.RGBA{
				Stride: 1 * 4,
				Rect:   image.Rect(0, 0, 1, 3),
				Pix: []uint8{
					0x00, 0x00, 0x00, 0xFF,
					0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF,
				},
			},
			want: "#000000",
		},
		{
			name: "50/50",
			img: &image.RGBA{
				Stride: 1 * 4,
				Rect:   image.Rect(0, 0, 1, 2),
				Pix: []uint8{
					0x00, 0x00, 0x00, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF,
				},
			},
			want: "#000000",
		},
		{
			name: "33/33/33",
			img: &image.RGBA{
				Stride: 1 * 4,
				Rect:   image.Rect(0, 0, 1, 3),
				Pix: []uint8{
					0xFF, 0xFF, 0xFF, 0xFF,
					0x00, 0x00, 0xFF, 0xFF,
					0xFF, 0x00, 0x00, 0xFF,
				},
			},
			want: "#FFFFFF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetDominantColor(ctx, tc.img)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}
