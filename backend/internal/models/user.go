package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	Email       string     `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password    string     `json:"-" gorm:"size:255;not null"`
	Name        string     `json:"name" gorm:"size:50;not null"`
	Phone       string     `json:"phone" gorm:"size:20"`
	Avatar      string     `json:"avatar" gorm:"size:500"`
	Status      int        `json:"status" gorm:"default:1"` // 1正常 0禁用 -1删除
	IsRoot      bool       `json:"is_root" gorm:"default:false"` // 是否超级管理员(root)
	IsApproved  bool       `json:"is_approved" gorm:"default:false"` // 是否已审核通过
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip" gorm:"size:50"`

	// 关联
	Roles  []Role      `json:"roles" gorm:"many2many:user_roles;"`
	Groups []UserGroup `json:"groups" gorm:"many2many:group_users;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// HasPermission 检查是否有权限
func (u *User) HasPermission(permCode string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Code == permCode {
				return true
			}
		}
	}
	return false
}

// IsSuperAdmin 是否超级管理员
func (u *User) IsSuperAdmin() bool {
	// root 用户直接是超级管理员
	if u.IsRoot {
		return true
	}
	for _, role := range u.Roles {
		if role.Code == string(RoleOwner) {
			return true
		}
	}
	return false
}

// 请求结构体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Phone    string `json:"phone"`
}

type CreateUserRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Name     string  `json:"name" binding:"required,min=2,max=50"`
	Phone    string  `json:"phone"`
	RoleIDs  []int64 `json:"role_ids"`
	Status   int     `json:"status" binding:"oneof=0 1"`
}

type UpdateUserRequest struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	RoleIDs []int64 `json:"role_ids"`
	Status  int     `json:"status" binding:"oneof=0 1 -1"`
}
