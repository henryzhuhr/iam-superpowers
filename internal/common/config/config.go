package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Server   ServerConfig   `mapstructure:"server"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Name           string `mapstructure:"name"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxConnections int    `mapstructure:"maxConnections"`
	IdleTimeout    int    `mapstructure:"idleTimeout"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	AccessTokenTTL  int    `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL int    `mapstructure:"refreshTokenTTL"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"useTLS"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")

	v.AutomaticEnv()
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.sslmode", "DB_SSLMODE")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")
	v.BindEnv("database.maxConnections", "DB_MAX_CONNECTIONS")
	v.BindEnv("database.idleTimeout", "DB_IDLE_TIMEOUT")
	v.BindEnv("redis.db", "REDIS_DB")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("jwt.accessTokenTTL", "JWT_ACCESS_TOKEN_TTL")
	v.BindEnv("jwt.refreshTokenTTL", "JWT_REFRESH_TOKEN_TTL")
	v.BindEnv("smtp.host", "SMTP_HOST")
	v.BindEnv("smtp.port", "SMTP_PORT")
	v.BindEnv("smtp.user", "SMTP_USER")
	v.BindEnv("smtp.password", "SMTP_PASSWORD")
	v.BindEnv("smtp.from", "SMTP_FROM")
	v.BindEnv("smtp.useTLS", "SMTP_USE_TLS")

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
		}
		// If file doesn't exist, skip config file loading — viper falls back to env/defaults
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
