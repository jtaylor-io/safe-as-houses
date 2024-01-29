package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config contains all app configuration
// Values are sourced by viper from config file or environment variables
type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// LoadConfig sources app configuration from config file or environment variables
func LoadConfig(path string) (Config, error) {
	var config Config
	var err error

	viper.SetDefault("DB_DRIVER", "postgres")
	viper.SetDefault("MIGRATION_URL", "file://db/migration")
	viper.SetDefault("REDIS_ADDRESS", "0.0.0.0:6379")
	viper.SetDefault("HTTP_SERVER_ADDRESS", "0.0.0.0:8080")
	viper.SetDefault("GRPC_SERVER_ADDRESS", "0.0.0.0:9090")
	viper.SetDefault("ACCESS_TOKEN_DURATION", "15m")
	viper.SetDefault("REFRESH_TOKEN_DURATION", "24h")

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}
