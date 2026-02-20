package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// InventoryHandler 库存处理器
type InventoryHandler struct {
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
	warehouseRepo *repository.WarehouseRepository
}

// NewInventoryHandler 创建库存处理器
func NewInventoryHandler(
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	warehouseRepo *repository.WarehouseRepository,
) *InventoryHandler {
	return &InventoryHandler{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		warehouseRepo: warehouseRepo,
	}
}

// List 库存列表
// GET /api/v1/inventory
func (h *InventoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")

	var warehouseID int64
	if wid := c.Query("warehouse_id"); wid != "" {
		warehouseID, _ = strconv.ParseInt(wid, 10, 64)
	}

	var productID int64
	if pid := c.Query("product_id"); pid != "" {
		productID, _ = strconv.ParseInt(pid, 10, 64)
	}

	lowStock := c.Query("low_stock") == "true"

	ctx := c.Request.Context()
	inventories, total, err := h.inventoryRepo.ListWithDetails(ctx, page, size, warehouseID, productID, lowStock, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": inventories,
		"pagination": gin.H{
			"page":  page,
			"size": size,
			"total": total,
		},
	})
}

// Get 库存详情
// GET /api/v1/inventory/:id
func (h *InventoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ctx := c.Request.Context()
	inventory, err := h.inventoryRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "库存记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"inventory": inventory})
}

// Update 更新库存配置
// PUT /api/v1/inventory/:id
func (h *InventoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req models.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	inventory, err := h.inventoryRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "库存记录不存在"})
		return
	}

	// 更新字段
	if req.AlertQty != nil {
		inventory.AlertQty = *req.AlertQty
	}
	if req.Location != "" {
		inventory.Location = req.Location
	}
	if req.BatchNo != "" {
		inventory.BatchNo = req.BatchNo
	}
	if req.ExpireAt != nil {
		t, err := time.Parse("2006-01-02", *req.ExpireAt)
		if err == nil {
			inventory.ExpireAt = &t
		}
	}

	if err := h.inventoryRepo.Update(inventory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "更新成功",
		"inventory": inventory,
	})
}

// Adjust 库存调整
// POST /api/v1/inventory/adjust
func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req models.AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择账套"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	ctx := c.Request.Context()

	// 验证商品和仓库存在
	product, err := h.productRepo.FindByIDWithContext(ctx, req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品不存在"})
		return
	}

	warehouse, err := h.warehouseRepo.FindByIDWithContext(ctx, req.WarehouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仓库不存在"})
		return
	}

	// 调整库存并记录流水
	err = h.inventoryRepo.AdjustQuantity(ctx, tenantID, req.ProductID, req.WarehouseID, req.ChangeQty, func(inv *models.Inventory) error {
		// 创建库存流水
		log := &models.InventoryLog{
			TenantID:    tenantID,
			ProductID:   req.ProductID,
			WarehouseID: req.WarehouseID,
			ChangeQty:   req.ChangeQty,
			BeforeQty:   inv.Quantity - req.ChangeQty,
			AfterQty:    inv.Quantity,
			RefType:     models.RefTypeAdjust,
			RefNo:       "ADJ" + time.Now().Format("20060102150405"),
			OperatorID:  uid,
			Remark:      req.Remark,
		}
		return h.inventoryRepo.CreateLog(log)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调整失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "调整成功",
		"product":     product.Name,
		"warehouse":   warehouse.Name,
		"change_qty":  req.ChangeQty,
	})
}

// Logs 库存流水
// GET /api/v1/inventory/logs
func (h *InventoryHandler) Logs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	refType := c.Query("ref_type")

	var productID int64
	if pid := c.Query("product_id"); pid != "" {
		productID, _ = strconv.ParseInt(pid, 10, 64)
	}

	var warehouseID int64
	if wid := c.Query("warehouse_id"); wid != "" {
		warehouseID, _ = strconv.ParseInt(wid, 10, 64)
	}

	ctx := c.Request.Context()
	logs, total, err := h.inventoryRepo.ListLogs(ctx, page, size, productID, warehouseID, refType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": logs,
		"pagination": gin.H{
			"page":  page,
			"size": size,
			"total": total,
		},
	})
}

// Alert 库存预警列表
// GET /api/v1/inventory/alert
func (h *InventoryHandler) Alert(c *gin.Context) {
	ctx := c.Request.Context()
	inventories, err := h.inventoryRepo.GetLowStockAlert(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": inventories})
}
