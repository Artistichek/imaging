package http

import (
	"github.com/Artistichek/imaging/test/errorcmp"
	"github.com/jarcoal/httpmock"
	"net/http"
	"net/url"
	"testing"
)

var (
	contentTypeErr  *ContentTypeError
	clientServerErr *ClientServerError
	urlErr          *url.Error
)

func TestGetImage(t *testing.T) {
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
			res: httpmock.NewStringResponder(200, "test").
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

		_, got := GetImage(tc.url)

		if diff := errorcmp.Diff(tc.want, got); diff != "" {
			t.Fatal(diff)
		}
	}
}
