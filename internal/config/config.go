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
		KeyValue     KeyValue      `yaml:"key_value"`
		FileStorage  FileStorage   `yaml:"file_storage"`
		SearchEngine SearchEngine  `yaml:"search_engine"`
		Kafka        KafkaConfig   `yaml:"kafka"`
		Workers      WorkersConfig `yaml:"workers"`
		Images       ImagesConfig  `yaml:"images"`
		Email        EmailConfig   `yaml:"email"`
		Auth         AuthConfig    `yaml:"auth"`
		Order        OrderConfig   `yaml:"order"`
	}

	// OrderConfig tunes checkout. Money amounts are decimal strings ("0" disables
	// the free-shipping waiver) parsed once at service init.
	OrderConfig struct {
		Currency         string        `yaml:"currency" default:"RUB" usage:"ISO-4217 currency code stamped on new orders"`
		FlatShipping     string        `yaml:"flat_shipping" default:"0" usage:"flat shipping fee added to every order"`
		FreeShippingFrom string        `yaml:"free_shipping_from" default:"0" usage:"subtotal at/above which shipping is waived; 0 disables"`
		PendingTTL       time.Duration `yaml:"pending_ttl" default:"30m" usage:"how long an unpaid order lives before the expiry worker cancels it"`
	}

	// AuthConfig tunes the passwordless email one-time-code flow used by regular
	// users. Admin accounts log in with a password and ignore these.
	AuthConfig struct {
		OTPCodeTTL      time.Duration `yaml:"otp_code_ttl" default:"10m" usage:"how long an emailed login code stays valid"`
		OTPMaxAttempts  int           `yaml:"otp_max_attempts" default:"5" usage:"wrong-code guesses before the code is burned"`
		OTPResendWindow time.Duration `yaml:"otp_resend_window" default:"1m" usage:"minimum gap between code requests for one email"`
		OTPHourlyLimit  int           `yaml:"otp_hourly_limit" default:"5" usage:"max code requests per email per hour"`
	}

	// EmailConfig configures outgoing transactional email. When Enabled is false
	// every message is silently dropped (handy for local runs without an SMTP
	// server). Credentials live here, not in the database.
	EmailConfig struct {
		Enabled  bool   `yaml:"enabled" default:"false" usage:"enable outgoing email; when false messages are dropped"`
		Host     string `yaml:"host" default:"localhost" usage:"SMTP server host"`
		Port     int    `yaml:"port" default:"1025" usage:"SMTP server port"`
		Username string `yaml:"username" usage:"SMTP auth username (leave empty to skip AUTH)"`
		Password string `yaml:"password" secret:"true" usage:"SMTP auth password"`
		From     string `yaml:"from" default:"no-reply@go-store.local" usage:"From header address"`
		FromName string `yaml:"from_name" default:"go-store" usage:"From header display name"`
		BaseURL  string `yaml:"base_url" default:"http://localhost:8888" usage:"public base URL used for links in emails"`
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

	WorkersConfig struct {
		ImageSync int `yaml:"image_sync" default:"3"`
		MailSend  int `yaml:"mail_send" default:"2"`
	}

	// ImagesConfig controls on-the-fly image resizing. Variants are generated on the
	// first request to /storage/public/images/<base>_<size>.<webp|jpg> and cached on
	// disk next to the originals (delete the *_<size>.* files to force regeneration).
	ImagesConfig struct {
		JpegQuality     int           `yaml:"jpeg_quality" default:"82"`
		WebpQuality     int           `yaml:"webp_quality" default:"80"`
		MaxSourcePixels int           `yaml:"max_source_pixels" default:"40000000"`
		Presets         []ImagePreset `yaml:"presets"`
	}

	ImagePreset struct {
		Name string `yaml:"name"`
		// Size is the target square box (both dimensions) and the "_<size>" URL suffix.
		Size      int    `yaml:"size"`
		Fit       string `yaml:"fit" default:"cover" example:"cover / contain"`
		NoUpscale bool   `yaml:"no_upscale"`
	}
)

// DefaultImagePresets is used when config.images.presets is empty. Sizes match the
// front-end contract (thumb 160, card 400, pdp 1320, zoom 1920).
func DefaultImagePresets() []ImagePreset {
	return []ImagePreset{
		{Name: "thumb", Size: 160, Fit: "cover"},
		{Name: "card", Size: 400, Fit: "cover"},
		{Name: "pdp", Size: 1320, Fit: "contain", NoUpscale: true},
		{Name: "zoom", Size: 1920, Fit: "contain", NoUpscale: true},
	}
}

// ResolvedPresets returns the configured presets, or the built-in defaults if none are set.
func (c ImagesConfig) ResolvedPresets() []ImagePreset {
	if len(c.Presets) == 0 {
		return DefaultImagePresets()
	}
	return c.Presets
}

type KeyValueEngine string

const (
	KeyValueEngineInMemory KeyValueEngine = "in_memory"
	KeyValueEngineRedis    KeyValueEngine = "redis"
)
