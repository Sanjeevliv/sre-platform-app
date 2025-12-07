package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	APIPort        string `mapstructure:"API_PORT"`
	WorkerPort     string `mapstructure:"WORKER_PORT"`
	RedisAddr      string `mapstructure:"REDIS_ADDR"`
	RateLimitRPS   int    `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst int    `mapstructure:"RATE_LIMIT_BURST"`
}

func Load() (*Config, error) {
	// 1. Set Defaults
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("WORKER_PORT", "8081")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("RATE_LIMIT_RPS", 100)
	viper.SetDefault("RATE_LIMIT_BURST", 50)

	// 2. Load from .env file (if present)
	viper.SetConfigName(".env") // name of config file (without extension)
	viper.SetConfigType("env")  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")    // optionally look for config in the working directory

	// 3. Load from Environment Variables (overrides defaults and file)
	// Viper automatically matches checks for env vars that match the keys
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("No .env file found, using defaults and environment variables")
		} else {
			// Config file was found but another error produced
			return nil, err
		}
	} else {
		log.Println("Loaded configuration from .env file")
	}

	// 4. Unmarshal into Struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
