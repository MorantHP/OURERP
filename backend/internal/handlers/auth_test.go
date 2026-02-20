package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAuthHandler_Login 测试登录功能
func TestAuthHandler_Login(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	db.AutoMigrate(&models.User{})

	// 创建测试用户
	userRepo := repository.NewUserRepository(db)
	testUser := &models.User{
		Email:      "test@example.com",
		Name:       "Test User",
		IsApproved: true,
		Status:     1,
	}
	testUser.SetPassword("password123")
	userRepo.Create(testUser)

	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 创建配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing",
			Expire: 24,
		},
	}

	// 创建 handler
	handler := NewAuthHandler(userRepo, cfg)

	// 测试用例
	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{
			name:       "有效登录",
			email:      "test@example.com",
			password:   "password123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "错误密码",
			email:      "test@example.com",
			password:   "wrongpassword",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "不存在的用户",
			email:      "nonexistent@example.com",
			password:   "password123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "空邮箱",
			email:      "",
			password:   "password123",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "空密码",
			email:      "test@example.com",
			password:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建请求
			body := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 创建 Gin 上下文
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 执行登录
			handler.Login(c)

			// 验证状态码
			if w.Code != tt.wantStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// 验证成功登录返回 token
			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				if _, ok := response["token"]; !ok {
					t.Error("Login() should return token on success")
				}
			}
		})
	}
}

// TestAuthHandler_Register 测试注册功能
func TestAuthHandler_Register(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	db.AutoMigrate(&models.User{})
	userRepo := repository.NewUserRepository(db)

	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expire: 24,
		},
	}

	handler := NewAuthHandler(userRepo, cfg)

	tests := []struct {
		name       string
		email      string
		password   string
		nameField  string
		wantStatus int
	}{
		{
			name:       "有效注册",
			email:      "newuser@example.com",
			password:   "password123",
			nameField:  "New User",
			wantStatus: http.StatusOK,
		},
		{
			name:       "邮箱已存在",
			email:      "newuser@example.com", // 第二次注册同一个邮箱
			password:   "password123",
			nameField:  "Another User",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "无效邮箱格式",
			email:      "invalid-email",
			password:   "password123",
			nameField:  "Test",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "密码太短",
			email:      "short@example.com",
			password:   "123",
			nameField:  "Short Pass",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"email":    tt.email,
				"password": tt.password,
				"name":     tt.nameField,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Register(c)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestUserModel_SetPassword 测试密码设置和验证
func TestUserModel_SetPassword(t *testing.T) {
	user := &models.User{}

	password := "mypassword123"
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	// 验证密码已加密存储
	if user.Password == password {
		t.Error("SetPassword() should hash the password")
	}

	// 验证密码校验
	if !user.CheckPassword(password) {
		t.Error("CheckPassword() should return true for correct password")
	}

	if user.CheckPassword("wrongpassword") {
		t.Error("CheckPassword() should return false for wrong password")
	}
}
