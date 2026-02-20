package handlers

import (
	"net/http"
	"time"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		config:   cfg,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查邮箱是否已存在
	existingUser, _ := h.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
		return
	}

	// 创建用户（默认未审核）
	user := &models.User{
		Email:      req.Email,
		Name:       req.Name,
		IsApproved: false, // 需要等待 root 审核
		Status:     1,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功，请等待管理员审核",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号已被禁用"})
		return
	}

	// 检查是否已审核（root 用户跳过审核检查）
	if !user.IsRoot && !user.IsApproved {
		c.JSON(http.StatusForbidden, gin.H{"error": "账号尚未通过审核，请等待管理员审核"})
		return
	}

	// 生成 JWT
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"is_root":   user.IsRoot,
			"is_approved": user.IsApproved,
		},
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.userRepo.FindByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"phone":       user.Phone,
			"is_root":     user.IsRoot,
			"is_approved": user.IsApproved,
		},
	})
}

// ListUsers 列出所有用户（仅 root）
func (h *AuthHandler) ListUsers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUser, _ := h.userRepo.FindByID(userID.(int64))

	// 只有 root 可以查看所有用户
	if !currentUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	users, err := h.userRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构建返回数据
	result := make([]gin.H, 0)
	for _, u := range users {
		result = append(result, gin.H{
			"id":          u.ID,
			"email":       u.Email,
			"name":        u.Name,
			"phone":       u.Phone,
			"status":      u.Status,
			"is_root":     u.IsRoot,
			"is_approved": u.IsApproved,
			"created_at":  u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

// ApproveUser 审核通过用户（仅 root）
func (h *AuthHandler) ApproveUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUser, _ := h.userRepo.FindByID(userID.(int64))

	// 只有 root 可以审核用户
	if !currentUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	targetID := c.Param("id")
	var req struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID := parseInt64(targetID)
	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	targetUser, err := h.userRepo.FindByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不能修改 root 用户
	if targetUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能修改 root 用户"})
		return
	}

	// 更新审核状态
	targetUser.IsApproved = req.Approved
	if err := h.userRepo.Update(targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	message := "审核通过"
	if !req.Approved {
		message = "已拒绝"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"user": gin.H{
			"id":          targetUser.ID,
			"email":       targetUser.Email,
			"name":        targetUser.Name,
			"is_approved": targetUser.IsApproved,
		},
	})
}

// DeleteUser 删除用户（仅 root）
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUser, _ := h.userRepo.FindByID(userID.(int64))

	// 只有 root 可以删除用户
	if !currentUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	targetID := c.Param("id")
	targetUserID := parseInt64(targetID)
	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	targetUser, err := h.userRepo.FindByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不能删除 root 用户
	if targetUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除 root 用户"})
		return
	}

	if err := h.userRepo.Delete(targetUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SetUserStatus 设置用户状态（仅 root）
func (h *AuthHandler) SetUserStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUser, _ := h.userRepo.FindByID(userID.(int64))

	// 只有 root 可以设置用户状态
	if !currentUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	targetID := c.Param("id")
	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID := parseInt64(targetID)
	if targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	targetUser, err := h.userRepo.FindByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不能修改 root 用户
	if targetUser.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能修改 root 用户"})
		return
	}

	targetUser.Status = req.Status
	if err := h.userRepo.Update(targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "状态更新成功",
		"user": gin.H{
			"id":     targetUser.ID,
			"status": targetUser.Status,
		},
	})
}

func parseInt64(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}

// generateToken 生成 JWT token
func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(h.config.JWT.Expire) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWT.Secret))
}