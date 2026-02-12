// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expire int // 小时
}

var GlobalConfig *Config

func Load() *Config {
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "ourerp")
	viper.SetDefault("DB_PASSWORD", "ourerp123")
	viper.SetDefault("DB_NAME", "ourerp")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("JWT_SECRET", "your-secret-key")
	viper.SetDefault("JWT_EXPIRE", 24)

	viper.AutomaticEnv()

	config := &Config{
		Env: viper.GetString("ENV"),
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			Expire: viper.GetInt("JWT_EXPIRE"),
		},
	}

	GlobalConfig = config
	return config
}
