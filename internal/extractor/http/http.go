package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ClientServerError struct {
	status string
	url    string
}

func (e *ClientServerError) Error() string {
	return fmt.Sprintf("client or server error: url=%s, status=%s", e.url, e.status)
}

type ContentTypeError struct {
	contentType string
}

func (e *ContentTypeError) Error() string {
	return fmt.Sprintf("content type error: resource contentType=%s, expected image/*", e.contentType)
}

func GetImage(rawUrl string) ([]byte, error) {
	if _, err := url.ParseRequestURI(rawUrl); err != nil {
		return nil, err
	}

	res, err := http.Get(rawUrl)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, &ClientServerError{
			status: res.Status,
			url:    rawUrl,
		}
	}

	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image") {
		return nil, &ContentTypeError{
			contentType: ct,
		}
	}

	return io.ReadAll(res.Body)
}
