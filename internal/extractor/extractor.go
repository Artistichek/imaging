package extractor

import (
	"bytes"
	"image"

	imagingpb "github.com/Artistichek/imaging/api/imaging/v1"

	"github.com/Artistichek/imaging/internal/extractor/decoding"
	"github.com/Artistichek/imaging/internal/extractor/http"
)

func ExtractImage(req *imagingpb.ProcessImageRequest) (image.Image, error) {
	var buffer []byte
	var err error

	switch i := req.Image.(type) {
	case *imagingpb.ProcessImageRequest_File:
		buffer = i.File
	case *imagingpb.ProcessImageRequest_Url:
		if buffer, err = http.GetImage(i.Url); err != nil {
			return nil, err
		}
	}

	return decoding.DecodeImage(bytes.NewReader(buffer))
}
