package config

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/Artistichek/imaging/logs"

	"github.com/Artistichek/imaging/internal/processor"
	"github.com/Artistichek/imaging/internal/s3"
)

var (
	//go:embed config.yml
	configBytes []byte
)

type Config struct {
	Logger LoggerConfig
	GRPC   GRPCServerConfig

	Processor processor.Config
	S3        s3.Config
}

func (c Config) ConfigBytes() []byte {
	return configBytes
}

func (c Config) EnvPrefix() string {
	return "imaging"
}

type LoggerConfig struct {
	Level logs.Level

	Output logs.Output
}

type GRPCServerConfig struct {
	Port int
}

func Load(cfg Config) (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix(cfg.EnvPrefix())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	v.SetConfigType("yml")
	if err := v.ReadConfig(bytes.NewBuffer(cfg.ConfigBytes())); err != nil {
		return nil, err
	}

	if err := v.MergeInConfig(); err != nil {
		if errors.Is(err, &viper.ConfigParseError{}) {
			return nil, err
		}
	}

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		StringToZerologLevelHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	var c Config
	if err := v.Unmarshal(&c, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, err
	}

	return &c, nil
}

func Log(ctx context.Context, cfg Config) {
	log := logs.FromContext(ctx)
	msg, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Err(err).Msg("log service config")
		return
	}

	log.Info().Msg(fmt.Sprintf("config: %s", msg))
}

func StringToZerologLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(zerolog.Level(1)) || f.Kind() != reflect.String {
			return data, nil
		}

		return zerolog.ParseLevel(data.(string))
	}
}
