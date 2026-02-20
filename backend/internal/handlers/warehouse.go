package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// WarehouseHandler 仓库处理器
type WarehouseHandler struct {
	warehouseRepo *repository.WarehouseRepository
}

// NewWarehouseHandler 创建仓库处理器
func NewWarehouseHandler(warehouseRepo *repository.WarehouseRepository) *WarehouseHandler {
	return &WarehouseHandler{warehouseRepo: warehouseRepo}
}

// List 仓库列表
// GET /api/v1/warehouses
func (h *WarehouseHandler) List(c *gin.Context) {
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	warehouses, err := h.warehouseRepo.ListWithContext(ctx, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": warehouses})
}

// Get 仓库详情
// GET /api/v1/warehouses/:id
func (h *WarehouseHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ctx := c.Request.Context()
	warehouse, err := h.warehouseRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仓库不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"warehouse": warehouse})
}

// Create 创建仓库
// POST /api/v1/warehouses
func (h *WarehouseHandler) Create(c *gin.Context) {
	var req models.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取租户ID
	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择账套"})
		return
	}

	// 检查编码是否已存在
	ctx := c.Request.Context()
	if _, err := h.warehouseRepo.FindByCode(ctx, req.Code); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仓库编码已存在"})
		return
	}

	// 设置默认类型
	whType := req.Type
	if whType == "" {
		whType = models.WarehouseTypeNormal
	}

	warehouse := &models.Warehouse{
		TenantID: tenantID,
		Code:     req.Code,
		Name:     req.Name,
		Address:  req.Address,
		Contact:  req.Contact,
		Phone:    req.Phone,
		Type:     whType,
		Status:   models.WarehouseStatusEnabled,
	}

	if err := h.warehouseRepo.Create(warehouse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "创建成功",
		"warehouse": warehouse,
	})
}

// Update 更新仓库
// PUT /api/v1/warehouses/:id
func (h *WarehouseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req models.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	warehouse, err := h.warehouseRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仓库不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		warehouse.Name = req.Name
	}
	if req.Address != "" {
		warehouse.Address = req.Address
	}
	if req.Contact != "" {
		warehouse.Contact = req.Contact
	}
	if req.Phone != "" {
		warehouse.Phone = req.Phone
	}
	if req.Type != "" {
		warehouse.Type = req.Type
	}
	if req.Status != nil {
		warehouse.Status = *req.Status
	}
	if req.IsDefault != nil {
		warehouse.IsDefault = *req.IsDefault
	}

	if err := h.warehouseRepo.Update(warehouse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "更新成功",
		"warehouse": warehouse,
	})
}

// Delete 删除仓库
// DELETE /api/v1/warehouses/:id
func (h *WarehouseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.warehouseRepo.DeleteWithContext(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SetDefault 设置默认仓库
// POST /api/v1/warehouses/:id/default
func (h *WarehouseHandler) SetDefault(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择账套"})
		return
	}

	ctx := c.Request.Context()
	if err := h.warehouseRepo.SetDefault(ctx, tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}
