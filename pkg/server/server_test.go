package server

import (
	"context"
	"github.com/Artistichek/imaging/internal/processor"
	"github.com/Artistichek/imaging/internal/processor/dominator"
	"github.com/Artistichek/imaging/internal/s3"
	"github.com/Artistichek/imaging/logs"
	"image"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/jarcoal/httpmock"
	"google.golang.org/protobuf/testing/protocmp"

	imagingpb "github.com/Artistichek/imaging/api/imaging/v1"

	pcsmock "github.com/Artistichek/imaging/internal/processor/mock"
	s3mock "github.com/Artistichek/imaging/internal/s3/mock"

	testbase64 "github.com/Artistichek/imaging/test/base64"
	"github.com/Artistichek/imaging/test/errorcmp"
)

var (
	// image.png 5x5
	valid = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAAAAEElEQVR42mP8X88ABYwkMAHtIQd8SvDAUAAAAABJRU5ErkJggg=="
	// corrupted image.png 5x5
	corrupted = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAwAsEElEQVSz2mP8Yc8ARYxkMAHtIQd8SvCtugAAAAB4RU7drkJggg=="
)

func TestServer_ProcessImage_ExtractImageError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ctrl := gomock.NewController(t)

	var (
		p   = pcsmock.NewMockImageProcessor(ctrl)
		c   = s3mock.NewMockAPIClient(ctrl)
		log = logs.FromContext(context.TODO())
	)

	srv := &ImagingServer{
		p:   p,
		s3:  c,
		log: log,
	}

	testCases := []struct {
		name string

		url  string
		want imagingpb.ProcessResult
		res  httpmock.Responder
		err  error
	}{
		{
			name: "fail content type",
			url:  "https://test.com/200",
			want: imagingpb.ProcessResult_IMAGE_NOT_PROVIDED,
			res: httpmock.
				NewStringResponder(200, "test").
				HeaderSet(http.Header{"Content-Type": {"text/html"}}),
			err: contentTypeErr,
		},
		{
			name: "fail bad url",
			url:  "test",
			want: imagingpb.ProcessResult_IMAGE_NOT_PROVIDED,
			res:  httpmock.NewStringResponder(200, "test"),
			err:  urlErr,
		},
		{
			name: "fail unreachable resource",
			url:  "https://test.com/400",
			want: imagingpb.ProcessResult_IMAGE_URL_UNREACHABLE,
			res: httpmock.NewStringResponder(400, "test").
				HeaderSet(http.Header{"Content-Type": {"image/jpeg"}}),
			err: clientServerErr,
		},
		{
			name: "fail corrupted image",
			url:  "https://test.com/200",
			want: imagingpb.ProcessResult_CORRUPTED_IMAGE,
			res: httpmock.NewBytesResponder(200, testbase64.Decode(corrupted)).
				HeaderSet(http.Header{"Content-Type": {"image/png"}}),
			err: decodingErr,
		},
		{
			name: "fail unknown image format",
			url:  "https://test.com/200",
			want: imagingpb.ProcessResult_INVALID_IMAGE_FORMAT,
			res: httpmock.NewStringResponder(200, "image with not supported format").
				HeaderSet(http.Header{"Content-Type": {"image/fmt"}}),
			err: image.ErrFormat,
		},
	}

	for _, tc := range testCases {
		httpmock.RegisterResponder("GET", tc.url, tc.res)

		req := &imagingpb.ProcessImageRequest{
			JobId:  0,
			GameId: "0",
			Image: &imagingpb.ProcessImageRequest_Url{
				Url: tc.url,
			},
		}

		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.ProcessImage(context.Background(), req)

			if diff := cmp.Diff(tc.want, got.Result); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}

			if diff := errorcmp.Diff(tc.err, err); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestServer_ProcessImage_ProcessError(t *testing.T) {
	var valid = testbase64.Decode(valid)
	var ctrl = gomock.NewController(t)

	var (
		p   = pcsmock.NewMockImageProcessor(ctrl)
		c   = s3mock.NewMockAPIClient(ctrl)
		log = logs.FromContext(context.TODO())
	)

	srv := &ImagingServer{
		p:   p,
		s3:  c,
		log: log,
	}

	testCases := []struct {
		name string

		want imagingpb.ProcessResult
		err  error
	}{
		{
			name: "fail extract dominant color",
			want: imagingpb.ProcessResult_INTERNAL_ERROR,
			err:  dominator.ErrColorExtraction,
		},
		{
			name: "fail encode",
			want: imagingpb.ProcessResult_INTERNAL_ERROR,
			err:  encodingErr,
		},
		{
			name: "fail timeout",
			want: imagingpb.ProcessResult_PROCESS_TIMEOUT_EXCEEDED,
			err:  processTimeoutErr,
		},
	}

	for _, tc := range testCases {
		req := &imagingpb.ProcessImageRequest{
			JobId:  0,
			GameId: "0",
			Image: &imagingpb.ProcessImageRequest_File{
				File: valid,
			},
		}

		p.EXPECT().
			ProcessImage(gomock.Any(), gomock.Any()).
			Return(nil, "", tc.err).
			Times(1)

		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.ProcessImage(context.Background(), req)

			if diff := cmp.Diff(tc.want, got.Result, protocmp.Transform()); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}

			if diff := errorcmp.Diff(tc.err, err); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestServer_ProcessImage_UploadImagesError(t *testing.T) {
	var valid = testbase64.Decode(valid)
	var ctrl = gomock.NewController(t)

	var (
		p   = pcsmock.NewMockImageProcessor(ctrl)
		c   = s3mock.NewMockAPIClient(ctrl)
		log = logs.FromContext(context.TODO())
	)

	srv := &ImagingServer{
		p:   p,
		s3:  c,
		log: log,
	}

	testCases := []struct {
		name string

		want imagingpb.ProcessResult
		err  error
	}{
		{
			name: "fail operation error",
			want: imagingpb.ProcessResult_S3_IMAGE_UPLOAD_ERROR,
			err:  operationErr,
		},
		{
			name: "fail upload timeout",
			want: imagingpb.ProcessResult_UPLOAD_TIMEOUT_EXCEEDED,
			err:  uploadTimeoutErr,
		},
	}

	for _, tc := range testCases {
		req := &imagingpb.ProcessImageRequest{
			JobId:  0,
			GameId: "0",
			Image: &imagingpb.ProcessImageRequest_File{
				File: valid,
			},
		}

		p.EXPECT().Cfg().Return(&processor.Config{
			Sizes:          []int{},
			EncodingFormat: "test",
			ProcessTimeout: 100 * time.Millisecond,
		}).AnyTimes()

		c.EXPECT().Cfg().Return(&s3.Config{
			Credentials:      s3.Credentials{},
			EndpointResolver: s3.EndpointResolver{},
			Bucket:           "test",
			BaseDirectory:    "test",
			UploadTimeout:    100 * time.Millisecond,
		}).AnyTimes()

		p.EXPECT().
			ProcessImage(gomock.Any(), gomock.Any()).
			Return(nil, "", nil).
			Times(1)

		c.EXPECT().
			UploadImages(gomock.Any(), gomock.Any()).
			Return(tc.err).
			Times(1)

		c.EXPECT().
			DeleteImages(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.ProcessImage(context.Background(), req)

			if diff := cmp.Diff(tc.want, got.Result, protocmp.Transform()); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}

			if diff := errorcmp.Diff(tc.err, err); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
