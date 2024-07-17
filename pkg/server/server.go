package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/Artistichek/imaging/internal/processor/encoding"
	"image"
	"net/url"
	"strconv"

	imagingpb "github.com/Artistichek/imaging/api/imaging/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/Artistichek/imaging/logs"

	"github.com/Artistichek/imaging/internal/extractor"
	"github.com/Artistichek/imaging/internal/extractor/decoding"
	"github.com/Artistichek/imaging/internal/extractor/http"
	"github.com/Artistichek/imaging/internal/processor"
	"github.com/Artistichek/imaging/internal/processor/dominator"
	"github.com/Artistichek/imaging/internal/s3"
	"github.com/Artistichek/imaging/internal/s3/types"
)

var (
	urlErr            *url.Error
	contentTypeErr    *http.ContentTypeError
	clientServerErr   *http.ClientServerError
	decodingErr       *decoding.Error
	encodingErr       *encoding.Error
	operationErr      *s3.OperationError
	uploadTimeoutErr  *s3.UploadTimeoutError
	processTimeoutErr *processor.TimeoutError
)

type ImagingServer struct {
	imagingpb.UnimplementedImagingServiceServer

	p  processor.ImageProcessor
	s3 s3.APIClient

	log *logs.Logger
}

func New(ctx context.Context, p processor.ImageProcessor, s3 s3.APIClient) *ImagingServer {
	log := logs.FromContext(ctx).With().Str("system", "server").Logger()

	return &ImagingServer{
		p:   p,
		s3:  s3,
		log: &log,
	}
}

func (s *ImagingServer) ProcessImage(ctx context.Context, req *imagingpb.ProcessImageRequest) (*imagingpb.ProcessImageResponse, error) {
	resp := &imagingpb.ProcessImageResponse{
		JobId:  req.JobId,
		GameId: req.GameId,
	}
	log := s.log.With().
		Int64("job_id", req.JobId).
		Str("game_id", req.GameId).
		Logger()

	var err error
	var img image.Image

	if img, err = extractor.ExtractImage(req); err != nil {
		log.Err(err).Msg("extract image from request")

		switch {
		case errors.As(err, &urlErr) || errors.As(err, &contentTypeErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_IMAGE_NOT_PROVIDED, codes.InvalidArgument, err)
		case errors.As(err, &clientServerErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_IMAGE_URL_UNREACHABLE, codes.InvalidArgument, err)
		case errors.As(err, &decodingErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_CORRUPTED_IMAGE, codes.InvalidArgument, err)
		case errors.Is(err, image.ErrFormat):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_INVALID_IMAGE_FORMAT, codes.InvalidArgument, err)
		default:
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_PROCESS_RESULT_UNKNOWN, codes.InvalidArgument, err)
		}

		return resp, err
	}

	var color string
	var processed []bytes.Buffer

	if processed, color, err = s.p.ProcessImage(ctx, img); err != nil {
		log.Err(err).Msg("process image")

		switch {
		case errors.As(err, &decodingErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_INTERNAL_ERROR, codes.Internal, err)
		case errors.As(err, &processTimeoutErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_PROCESS_TIMEOUT_EXCEEDED, codes.DeadlineExceeded, err)
		case errors.Is(err, dominator.ErrColorExtraction) || errors.Is(err, encodingErr):
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_INTERNAL_ERROR, codes.Internal, err)
		default:
			resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_PROCESS_RESULT_UNKNOWN, codes.InvalidArgument, err)
		}

		log.Info().Msg("process image finished with error")

		return resp, err
	}

	var data = make([]types.PutObjectInput, len(processed))
	var locations = make([]string, len(processed))

	for i, size := range s.p.Cfg().Sizes {
		key := types.NewObjectKey(req.GameId, strconv.Itoa(size),
			types.WithBase(s.s3.Cfg().BaseDirectory),
			types.WithFormat(s.p.Cfg().EncodingFormat),
		)

		locations[i] = key.String()
		data[i] = types.PutObjectInput{
			Bucket:   s.s3.Cfg().Bucket,
			Key:      key,
			Body:     processed[i].Bytes(),
			Metadata: color,
		}
	}

	if err = s.s3.UploadImages(ctx, data); err == nil {
		resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_OK, codes.OK, err)

		s.log.Info().
			Str("bucket", s.s3.Cfg().Bucket).
			Strs("locations", locations).
			Msg("images uploaded")

		log.Info().Msg("process image finished")

		return resp, nil
	}

	log.Err(err).Msg("upload images")
	log.Info().Msg("upload images finished with error")

	switch {
	case errors.As(err, &uploadTimeoutErr):
		resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_UPLOAD_TIMEOUT_EXCEEDED, codes.DeadlineExceeded, err)
	case errors.As(err, &operationErr):
		resp.Status, resp.Result = statusWithResult(imagingpb.ProcessResult_S3_IMAGE_UPLOAD_ERROR, codes.Internal, err)
	}

	keys := make([]*types.ObjectKey, len(s.p.Cfg().Sizes))
	for i, size := range s.p.Cfg().Sizes {
		keys[i] = types.NewObjectKey(req.GameId, strconv.Itoa(size),
			types.WithBase(s.s3.Cfg().BaseDirectory),
			types.WithFormat(s.p.Cfg().EncodingFormat),
		)
	}

	input := types.DeleteObjectsInput{
		Bucket: s.s3.Cfg().Bucket,
		Keys:   keys,
	}

	if re := s.s3.DeleteImages(ctx, input); re != nil {
		s.log.Err(re).Msg("rollback process image request")
	}

	return resp, err
}

func statusWithResult(res imagingpb.ProcessResult, code codes.Code, err error) (*status.Status, imagingpb.ProcessResult) {
	if res == imagingpb.ProcessResult_OK {
		return grpcstatus.New(code, "process image finished").Proto(), res
	}

	s, _ := grpcstatus.Newf(code, "%v", err).WithDetails(&errdetails.ErrorInfo{
		Reason: res.String(),
	})

	return s.Proto(), res
}
