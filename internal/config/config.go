package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	ClickHouse ClickHouseConfig
	Meilisearch MeilisearchConfig
	Auth       AuthConfig
	SMTP       SMTPConfig
	S3         S3Config
	Log        LogConfig
	RateLimit  RateLimitConfig
}

type AppConfig struct {
	Env         string `mapstructure:"env"`
	Name        string `mapstructure:"name"`
	Port        int    `mapstructure:"port"`
	BaseURL     string `mapstructure:"base_url"`
	RedirectURL string `mapstructure:"redirect_url"`
	FrontendURL string `mapstructure:"frontend_url"`
	SecretKey   string `mapstructure:"secret_key"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ClickHouseConfig struct {
	URL      string `mapstructure:"url"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type MeilisearchConfig struct {
	URL    string `mapstructure:"url"`
	APIKey string `mapstructure:"api_key"`
}

type AuthConfig struct {
	TokenSecret       string        `mapstructure:"token_secret"`
	AccessTokenExpiry time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// Load reads configuration from config.yaml and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("/etc/linkrift")

	// Environment variable mapping
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map env vars to config keys
	bindEnvVars(v)

	// Set defaults
	setDefaults(v)

	// Read config file (optional — env vars take precedence)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}

func bindEnvVars(v *viper.Viper) {
	_ = v.BindEnv("app.env", "APP_ENV")
	_ = v.BindEnv("app.name", "APP_NAME")
	_ = v.BindEnv("app.port", "APP_PORT")
	_ = v.BindEnv("app.base_url", "APP_BASE_URL")
	_ = v.BindEnv("app.redirect_url", "APP_REDIRECT_URL")
	_ = v.BindEnv("app.frontend_url", "APP_FRONTEND_URL")
	_ = v.BindEnv("app.secret_key", "APP_SECRET_KEY")
	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("clickhouse.url", "CLICKHOUSE_URL")
	_ = v.BindEnv("clickhouse.database", "CLICKHOUSE_DATABASE")
	_ = v.BindEnv("clickhouse.user", "CLICKHOUSE_USER")
	_ = v.BindEnv("clickhouse.password", "CLICKHOUSE_PASSWORD")
	_ = v.BindEnv("meilisearch.url", "MEILISEARCH_URL")
	_ = v.BindEnv("meilisearch.api_key", "MEILISEARCH_API_KEY")
	_ = v.BindEnv("auth.token_secret", "AUTH_TOKEN_SECRET")
	_ = v.BindEnv("auth.access_token_expiry", "AUTH_ACCESS_TOKEN_EXPIRY")
	_ = v.BindEnv("auth.refresh_token_expiry", "AUTH_REFRESH_TOKEN_EXPIRY")
	_ = v.BindEnv("smtp.host", "SMTP_HOST")
	_ = v.BindEnv("smtp.port", "SMTP_PORT")
	_ = v.BindEnv("smtp.user", "SMTP_USER")
	_ = v.BindEnv("smtp.password", "SMTP_PASSWORD")
	_ = v.BindEnv("smtp.from", "SMTP_FROM")
	_ = v.BindEnv("s3.endpoint", "S3_ENDPOINT")
	_ = v.BindEnv("s3.bucket", "S3_BUCKET")
	_ = v.BindEnv("s3.access_key", "S3_ACCESS_KEY")
	_ = v.BindEnv("s3.secret_key", "S3_SECRET_KEY")
	_ = v.BindEnv("s3.region", "S3_REGION")
	_ = v.BindEnv("log.level", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "LOG_FORMAT")
	_ = v.BindEnv("ratelimit.requests", "RATE_LIMIT_REQUESTS")
	_ = v.BindEnv("ratelimit.window", "RATE_LIMIT_WINDOW")
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("app.name", "linkrift")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.base_url", "http://localhost:8080")
	v.SetDefault("app.redirect_url", "http://localhost:8081")
	v.SetDefault("app.frontend_url", "http://localhost:3000")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("redis.db", 0)
	v.SetDefault("clickhouse.database", "linkrift_analytics")
	v.SetDefault("auth.access_token_expiry", "15m")
	v.SetDefault("auth.refresh_token_expiry", "168h")
	v.SetDefault("smtp.host", "localhost")
	v.SetDefault("smtp.port", 1025)
	v.SetDefault("smtp.from", "noreply@linkrift.io")
	v.SetDefault("s3.region", "us-east-1")
	v.SetDefault("s3.bucket", "linkrift")
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.format", "console")
	v.SetDefault("ratelimit.requests", 100)
	v.SetDefault("ratelimit.window", "1m")
}
