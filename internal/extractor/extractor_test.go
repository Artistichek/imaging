package extractor

import (
	"github.com/Artistichek/imaging/internal/extractor/decoding"
	testbase64 "github.com/Artistichek/imaging/test/base64"
	"image"
	"net/http"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"

	imagingpb "github.com/Artistichek/imaging/api/imaging/v1"

	"github.com/Artistichek/imaging/test/errorcmp"

	extractorhttp "github.com/Artistichek/imaging/internal/extractor/http"
)

var (
	contentTypeErr  *extractorhttp.ContentTypeError
	clientServerErr *extractorhttp.ClientServerError
	urlErr          *url.Error
)

func TestExtractImage_URL(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	testCases := []struct {
		name string

		url  string
		res  httpmock.Responder
		want error
	}{
		{
			name: "valid resource",
			url:  "https://www.test.com/200.jpg",
			res: httpmock.NewBytesResponder(200, testbase64.Decode(png)).
				HeaderSet(http.Header{"Content-Type": {"image/jpeg"}}),
			want: nil,
		},
		{
			name: "invalid content type resource",
			url:  "https://www.test.com/200",
			res: httpmock.
				NewStringResponder(200, "test").
				HeaderSet(http.Header{"Content-Type": {"text/html"}}),
			want: contentTypeErr,
		},
		{
			name: "unreachable resource",
			url:  "https://www.test.com/404",
			res:  httpmock.NewStringResponder(404, "test"),
			want: clientServerErr,
		},
		{
			name: "invalid url",
			url:  "test",
			res:  httpmock.InitialTransport.RoundTrip,
			want: urlErr,
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

		_, got := ExtractImage(req)

		if diff := errorcmp.Diff(tc.want, got); diff != "" {
			t.Fatal(diff)
		}
	}
}

var decodingError *decoding.Error

var (
	// image.png 5x5
	png = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAAAAEElEQVR42mP8X88ABYwkMAHtIQd8SvDAUAAAAABJRU5ErkJggg=="
	// image.jpeg 5x5
	jpeg = "/9j/4AAQSkZJRgABAQAAAAAAAAD/4QBiRXhpZgAATU0AKgAAAAgABQESAAMAAAABAAEAAAEaAAUAAAABAAAASgEbAAUAAAABAAAAUgEoAAMAAAABAAEAAAITAAMAAAABAAEAAAAAAAAAAAAAAAAAAQAAAAAAAAAB/9sAQwADAgICAgIDAgICAwMDAwQGBAQEBAQIBgYFBgkICgoJCAkJCgwPDAoLDgsJCQ0RDQ4PEBAREAoMEhMSEBMPEBAQ/8AACwgABQAFAQERAP/EABQAAQAAAAAAAAAAAAAAAAAAAAn/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oACAEBAAA/AFTf/9k="
	// image.gif 5x5
	gif = "R0lGODlhBQAFAIAAAAAAAAAAACH5BAUKAAAALAAAAAAFAAUAAAIEhI+pWAA7"
	// corrupted image.png 5x5
	corrupted = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAQAAAAnZu5uAwAsEElEQVSz2mP8Yc8ARYxkMAHtIQd8SvCtugAAAAB4RU7drkJggg=="
)

func TestExtractImage_File(t *testing.T) {
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
		req := &imagingpb.ProcessImageRequest{
			JobId:  0,
			GameId: "0",
			Image: &imagingpb.ProcessImageRequest_File{
				File: tc.file,
			},
		}

		t.Run(tc.name, func(t *testing.T) {
			_, got := ExtractImage(req)

			if diff := errorcmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
