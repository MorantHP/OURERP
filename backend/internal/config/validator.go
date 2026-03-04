package config

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// ValidateConfig 验证配置
func ValidateConfig() error {
	// 验证JWT密钥
	if err := validateJWTSecret(); err != nil {
		return err
	}

	// 验证数据库配置
	if err := validateDatabaseConfig(); err != nil {
		return err
	}

	// 验证Redis配置
	if err := validateRedisConfig(); err != nil {
		return err
	}

	return nil
}

// validateJWTSecret 验证JWT密钥强度
func validateJWTSecret() error {
	secret := GlobalConfig.JWT.Secret
	
	if secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// 检查最小长度
	if len(secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long, got %d", len(secret))
	}

	// 检查是否是默认密钥
	defaultSecrets := []string{
		"secret",
		"jwt-secret",
		"your-secret-key",
		"change-me",
		"development",
	}
	for _, def := range defaultSecrets {
		if secret == def {
			return fmt.Errorf("JWT secret must not be a default value: %s", def)
		}
	}

	// 生产环境要求更强的密钥
	if GlobalConfig.Env == "production" {
		if len(secret) < 64 {
			return fmt.Errorf("JWT secret must be at least 64 characters long in production")
		}
	}

	return nil
}

// validateDatabaseConfig 验证数据库配置
func validateDatabaseConfig() error {
	db := GlobalConfig.Database

	if db.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if db.Port == "" {
		return fmt.Errorf("database port is required")
	}

	if db.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if db.User == "" {
		return fmt.Errorf("database user is required")
	}

	return nil
}

// validateRedisConfig 验证Redis配置
func validateRedisConfig() error {
	redis := GlobalConfig.Redis

	// Redis是可选的，如果配置了就验证
	if redis.Host == "" {
		return nil
	}

	if redis.Port == "" {
		return fmt.Errorf("redis port is required when host is set")
	}

	return nil
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 检查密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetEnvOrDefault 获取环境变量或默认值
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment 是否是开发环境
func IsDevelopment() bool {
	return GlobalConfig.Env == "development"
}

// IsProduction 是否是生产环境
func IsProduction() bool {
	return GlobalConfig.Env == "production"
}

// IsTest 是否是测试环境
func IsTest() bool {
	return GlobalConfig.Env == "test"
}
