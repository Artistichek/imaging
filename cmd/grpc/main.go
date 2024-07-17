package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	imagingpb "github.com/Artistichek/imaging/api/imaging/v1"

	"github.com/Artistichek/imaging/config"
	"github.com/Artistichek/imaging/logs"

	"github.com/Artistichek/imaging/pkg/server"

	"github.com/Artistichek/imaging/internal/processor"
	"github.com/Artistichek/imaging/internal/s3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load(config.Config{})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load service config error: %v", err)
		os.Exit(1)
	}

	log := logs.New(cfg.Logger.Level, cfg.Logger.Output)
	ctx = log.WithContext(ctx)

	config.Log(ctx, *cfg)

	err = run(ctx, cfg)

	switch {
	case errors.Is(err, context.Canceled):
		log.Info().Msg("gracefully stopped")
	case err != nil:
		log.Err(err).Msg("unexpectedly terminated")
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	log := logs.FromContext(ctx)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}

	var p processor.ImageProcessor
	var c s3.APIClient

	p = processor.New(ctx, &cfg.Processor)
	if c, err = s3.NewClient(ctx, &cfg.S3); err != nil {
		return fmt.Errorf("init client: %w", err)
	}

	imagingServer := server.New(ctx, p, c)

	serverRegister := grpc.NewServer()

	go func() {
		<-ctx.Done()

		log.Info().Msg("GRPC server gracefully stopped")
		serverRegister.GracefulStop()
	}()

	imagingpb.RegisterImagingServiceServer(serverRegister, imagingServer)

	log.Info().Int("port", cfg.GRPC.Port).Msg("GRPC server started")
	if err = serverRegister.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	return nil
}
