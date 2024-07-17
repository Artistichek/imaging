package logs

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Level = zerolog.Level
type Logger = zerolog.Logger

type Output string

const (
	ConsoleOutput Output = "console"
	JSONOutput    Output = "json"
)

const (
	TraceLevel   Level = zerolog.TraceLevel
	DebugLevel   Level = zerolog.DebugLevel
	InfoLevel    Level = zerolog.InfoLevel
	WarningLevel Level = zerolog.WarnLevel
	ErrorLevel   Level = zerolog.ErrorLevel
	FatalLevel   Level = zerolog.FatalLevel
	PanicLevel   Level = zerolog.PanicLevel
)

func New(l Level, output Output) *Logger {
	var w io.Writer
	switch output {
	case ConsoleOutput:
		w = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) { w.TimeFormat = time.RFC3339Nano })
	case JSONOutput:
		w = os.Stdout
	default:
		panic("cannot parse log output")
	}

	logger := log.
		Output(w).
		Level(l).
		With().
		Logger()

	return &logger
}

func FromContext(ctx context.Context) *Logger {
	return zerolog.Ctx(ctx)
}
