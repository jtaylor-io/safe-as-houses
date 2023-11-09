package util

import "github.com/spf13/viper"

// Config contains all app configuration
// Values are sourced by viper from config file or environment variables
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig sources app configuration from config file or environment variables
func LoadConfig(path string) (Config, error) {
	var config Config
	var err error
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
