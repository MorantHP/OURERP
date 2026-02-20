package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordConfig 密码配置
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig 默认密码配置
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword 使用Argon2id哈希密码
func HashPassword(password string, config *PasswordConfig) (string, error) {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	// 生成随机盐
	salt, err := generateRandomBytes(config.SaltLength)
	if err != nil {
		return "", err
	}

	// 使用Argon2id生成哈希
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// 编码为字符串格式: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		config.Memory,
		config.Iterations,
		config.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析哈希字符串
	config, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同参数生成哈希
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// 使用恒定时间比较防止时序攻击
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (*PasswordConfig, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported algorithm")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version")
	}

	config := &PasswordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&config.Memory,
		&config.Iterations,
		&config.Parallelism,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	config.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	config.KeyLength = uint32(len(hash))

	return config, salt, hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomToken 生成随机令牌
func GenerateRandomToken(length int) (string, error) {
	b, err := generateRandomBytes(uint32(length))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateResetToken 生成密码重置令牌
func GenerateResetToken() (string, error) {
	return GenerateRandomToken(32)
}

// GenerateVerifyToken 生成验证令牌
func GenerateVerifyToken() (string, error) {
	return GenerateRandomToken(32)
}
