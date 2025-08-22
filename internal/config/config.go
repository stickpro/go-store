package config

import (
	"time"

	"github.com/stickpro/go-store/pkg/logger"
)

type (
	Config struct {
		App          AppConfig  `yaml:"app"`
		HTTP         HTTPConfig `yaml:"http"`
		Postgres     PostgresDB `yaml:"postgres"`
		Redis        RedisDB    `yaml:"redis"`
		Log          logger.Config
		KeyValue     KeyValue     `yaml:"key_value"`
		FileStorage  FileStorage  `yaml:"file_storage"`
		SearchEngine SearchEngine `yaml:"search_engine"`
	}

	AppConfig struct {
		Profile string `yaml:"profile" default:"dev"`
	}

	HTTPConfig struct {
		Host               string         `yaml:"host" default:"localhost"`
		Port               string         `yaml:"port" default:"80"`
		ConnectTimeout     time.Duration  `yaml:"connect_timeout" env:"CONNECT_TIMEOUT" default:"5s"`
		ReadTimeout        time.Duration  `yaml:"read_timeout" env:"READ_TIMEOUT" default:"10s"`
		WriteTimeout       time.Duration  `yaml:"write_timeout" env:"WRITE_TIMEOUT" default:"10s"`
		MaxHeaderMegabytes int            `yaml:"max_header_megabytes" env:"MAX_HEADER_MEGABYTES" default:"1"`
		Cors               HTTPCorsConfig `yaml:"cors"`
		MaxBodyLimit       int            `yaml:"max_body_limit" default:"100" example:"100" usage:"maximum body size in mb, default 100MB"`
	}

	HTTPCorsConfig struct {
		Enabled        bool     `yaml:"enabled" default:"true" usage:"allows to disable cors" example:"true / false"`
		AllowedOrigins []string `yaml:"allowed_origins"`
	}

	KeyValue struct {
		Engine KeyValueEngine `yaml:"engine" required:"true" validate:"oneof=redis in_memory" example:"redis / in_memory" default:"redis"`
	}

	FileStorage struct {
		Type   string `yaml:"type" default:"local" example:"local / s3"`
		Path   string `yaml:"path" default:"storage"`
		Bucket string `yaml:"bucket" default:""`
	}
)

type KeyValueEngine string

const (
	KeyValueEngineInMemory KeyValueEngine = "in_memory"
	KeyValueEngineRedis    KeyValueEngine = "redis"
)
