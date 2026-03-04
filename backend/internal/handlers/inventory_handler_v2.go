package handlers

import (
	"strconv"

	"github.com/MorantHP/OURERP/internal/pkg/response"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

// InventoryHandlerV2 库存处理器（重构版）
type InventoryHandlerV2 struct {
	inventoryService *services.InventoryService
}

// NewInventoryHandlerV2 创建库存处理器
func NewInventoryHandlerV2(inventoryService *services.InventoryService) *InventoryHandlerV2 {
	return &InventoryHandlerV2{
		inventoryService: inventoryService,
	}
}

// List 库存列表
// GET /api/v1/inventory
func (h *InventoryHandlerV2) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var warehouseID, productID int64
	if w := c.Query("warehouse_id"); w != "" {
		warehouseID, _ = strconv.ParseInt(w, 10, 64)
	}
	if p := c.Query("product_id"); p != "" {
		productID, _ = strconv.ParseInt(p, 10, 64)
	}

	lowStock := c.Query("low_stock") == "true"
	keyword := c.Query("keyword")

	ctx := c.Request.Context()
	inventories, total, err := h.inventoryService.ListInventory(ctx, page, size, warehouseID, productID, lowStock, keyword)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPage(c, inventories, page, size, total)
}

// Get 库存详情
// GET /api/v1/inventory/:product_id/:warehouse_id
func (h *InventoryHandlerV2) Get(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	warehouseID, err := strconv.ParseInt(c.Param("warehouse_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的仓库ID")
		return
	}

	ctx := c.Request.Context()
	inventory, err := h.inventoryService.GetInventory(ctx, productID, warehouseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, inventory)
}

// Adjust 库存调整
// POST /api/v1/inventory/adjust
func (h *InventoryHandlerV2) Adjust(c *gin.Context) {
	var req struct {
		ProductID   int64  `json:"product_id" binding:"required"`
		WarehouseID int64  `json:"warehouse_id" binding:"required"`
		ChangeQty   int    `json:"change_qty" binding:"required"`
		Remark      string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.AdjustInventory(ctx, req.ProductID, req.WarehouseID, req.ChangeQty, "manual", "", req.Remark); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "调整成功", nil)
}

// Transfer 库存调拨
// POST /api/v1/inventory/transfer
func (h *InventoryHandlerV2) Transfer(c *gin.Context) {
	var req struct {
		ProductID       int64  `json:"product_id" binding:"required"`
		FromWarehouseID int64  `json:"from_warehouse_id" binding:"required"`
		ToWarehouseID   int64  `json:"to_warehouse_id" binding:"required"`
		Qty             int    `json:"qty" binding:"required"`
		RefNo           string `json:"ref_no"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.TransferStock(ctx, req.ProductID, req.FromWarehouseID, req.ToWarehouseID, req.Qty, req.RefNo); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "调拨成功", nil)
}

// Alerts 库存预警
// GET /api/v1/inventory/alerts
func (h *InventoryHandlerV2) Alerts(c *gin.Context) {
	ctx := c.Request.Context()
	inventories, err := h.inventoryService.GetLowStockAlert(ctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, inventories)
}
