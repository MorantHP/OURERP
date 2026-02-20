package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Shop 店铺模型
type Shop struct {
	ID             int64          `json:"id" gorm:"primaryKey"`
	TenantID       int64          `json:"tenant_id" gorm:"index;not null"` // 租户ID
	Name           string         `json:"name" gorm:"size:100;not null"`
	Platform       string         `json:"platform" gorm:"size:20;not null;index"`
	PlatformShopID string         `json:"platform_shop_id" gorm:"size:100"`
	AppKey         string         `json:"-" gorm:"size:100"`
	AppSecret      string         `json:"-" gorm:"size:100"`
	AccessToken    string         `json:"-" gorm:"size:1000"`
	RefreshToken   string         `json:"-" gorm:"size:1000"`
	TokenExpiresAt *time.Time     `json:"token_expires_at"`
	APIURL         string         `json:"api_url" gorm:"size:500"`
	WebhookURL     string         `json:"webhook_url" gorm:"size:500"`
	WebhookSecret  string         `json:"-" gorm:"size:100"`
	Status         int            `json:"status" gorm:"default:1"` // 1-启用 0-禁用
	SyncInterval   int            `json:"sync_interval" gorm:"default:30"` // 同步间隔(分钟)
	LastSyncAt     *time.Time     `json:"last_sync_at"`
	ExtraConfig    JSONB          `json:"extra_config" gorm:"type:jsonb"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// JSONB 自定义JSON类型用于PostgreSQL
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ShopStatus 店铺状态常量
const (
	ShopStatusDisabled = 0
	ShopStatusEnabled  = 1
)

// IsTokenExpired 检查Token是否过期
func (s *Shop) IsTokenExpired() bool {
	if s.TokenExpiresAt == nil {
		return true
	}
	return s.TokenExpiresAt.Before(time.Now())
}

// IsTokenExpiringSoon 检查Token是否即将过期(24小时内)
func (s *Shop) IsTokenExpiringSoon() bool {
	if s.TokenExpiresAt == nil {
		return true
	}
	return s.TokenExpiresAt.Before(time.Now().Add(24 * time.Hour))
}

// NeedSync 检查是否需要同步
func (s *Shop) NeedSync() bool {
	if s.Status != ShopStatusEnabled {
		return false
	}
	if s.LastSyncAt == nil {
		return true
	}
	return s.LastSyncAt.Add(time.Duration(s.SyncInterval) * time.Minute).Before(time.Now())
}
