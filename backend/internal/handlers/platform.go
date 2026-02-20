package handlers

import (
	"net/http"

	"github.com/MorantHP/OURERP/internal/platform"
	"github.com/gin-gonic/gin"
)

type PlatformHandler struct{}

func NewPlatformHandler() *PlatformHandler {
	return &PlatformHandler{}
}

// List 获取所有支持的平台
// GET /api/v1/platforms
func (h *PlatformHandler) List(c *gin.Context) {
	platforms := platform.GetAllPlatforms()

	c.JSON(http.StatusOK, gin.H{
		"platforms": platforms,
	})
}

// Get 获取单个平台配置
// GET /api/v1/platforms/:code
func (h *PlatformHandler) Get(c *gin.Context) {
	code := c.Param("code")

	config, ok := platform.GetPlatformConfig(platform.PlatformType(code))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "平台不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"platform": config,
	})
}
