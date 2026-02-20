package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

// FinanceHandler 财务处理器
type FinanceHandler struct {
	financeSvc *services.FinanceService
}

// NewFinanceHandler 创建财务处理器
func NewFinanceHandler(financeSvc *services.FinanceService) *FinanceHandler {
	return &FinanceHandler{financeSvc: financeSvc}
}

// ==================== 收支记录 ====================

// ListFinanceRecords 获取收支记录列表
func (h *FinanceHandler) ListFinanceRecords(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.FinanceRecordFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if recordType := c.Query("type"); recordType != "" {
		filter.Type = recordType
	}
	if category := c.Query("category"); category != "" {
		filter.Category = category
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}
	if shopID, err := strconv.ParseInt(c.Query("shop_id"), 10, 64); err == nil {
		filter.ShopID = &shopID
	}
	if startDate, err := time.Parse("2006-01-02", c.Query("start_date")); err == nil {
		filter.StartDate = startDate
	}
	if endDate, err := time.Parse("2006-01-02", c.Query("end_date")); err == nil {
		filter.EndDate = endDate
	}

	records, total, err := h.financeSvc.ListFinanceRecords(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":   records,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// CreateFinanceRecord 创建收支记录
func (h *FinanceHandler) CreateFinanceRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)

	var record models.FinanceRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record.TenantID = tenantID
	record.CreatedBy = userID

	if err := h.financeSvc.CreateFinanceRecord(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"record": record})
}

// GetFinanceRecord 获取收支记录详情
func (h *FinanceHandler) GetFinanceRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	record, err := h.financeSvc.GetFinanceRecord(tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

// UpdateFinanceRecord 更新收支记录
func (h *FinanceHandler) UpdateFinanceRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.UpdateFinanceRecord(tenantID, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteFinanceRecord 删除收支记录
func (h *FinanceHandler) DeleteFinanceRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.financeSvc.DeleteFinanceRecord(tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ApproveFinanceRecord 审核收支记录
func (h *FinanceHandler) ApproveFinanceRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.financeSvc.ApproveFinanceRecord(tenantID, id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "approved"})
}

// ==================== 平台账单 ====================

// ListPlatformBills 获取平台账单列表
func (h *FinanceHandler) ListPlatformBills(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.BillFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if platform := c.Query("platform"); platform != "" {
		filter.Platform = platform
	}
	if billPeriod := c.Query("bill_period"); billPeriod != "" {
		filter.BillPeriod = billPeriod
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}
	if shopID, err := strconv.ParseInt(c.Query("shop_id"), 10, 64); err == nil {
		filter.ShopID = &shopID
	}

	bills, total, err := h.financeSvc.ListPlatformBills(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bills":     bills,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// CreatePlatformBill 创建平台账单
func (h *FinanceHandler) CreatePlatformBill(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var bill models.PlatformBill
	if err := c.ShouldBindJSON(&bill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill.TenantID = tenantID

	if err := h.financeSvc.CreatePlatformBill(&bill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"bill": bill})
}

// GetPlatformBill 获取平台账单详情
func (h *FinanceHandler) GetPlatformBill(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	bill, err := h.financeSvc.GetPlatformBill(tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bill not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bill": bill})
}

// GetBillDetails 获取账单明细
func (h *FinanceHandler) GetBillDetails(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	details, err := h.financeSvc.GetBillDetails(tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"details": details})
}

// ReconcileBillDetail 对账单明细
func (h *FinanceHandler) ReconcileBillDetail(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	detailID, err := strconv.ParseInt(c.Param("detail_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid detail id"})
		return
	}

	var req struct {
		OrderID int64 `json:"order_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.ReconcileBillDetail(tenantID, detailID, req.OrderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reconciled"})
}

// ==================== 供应商 ====================

// ListSuppliers 获取供应商列表
func (h *FinanceHandler) ListSuppliers(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.SupplierFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if name := c.Query("name"); name != "" {
		filter.Name = name
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}

	suppliers, total, err := h.financeSvc.ListSuppliers(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suppliers": suppliers,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// CreateSupplier 创建供应商
func (h *FinanceHandler) CreateSupplier(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var supplier models.Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supplier.TenantID = tenantID

	if err := h.financeSvc.CreateSupplier(&supplier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"supplier": supplier})
}

// GetSupplier 获取供应商详情
func (h *FinanceHandler) GetSupplier(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	supplier, err := h.financeSvc.GetSupplier(tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "supplier not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"supplier": supplier})
}

// UpdateSupplier 更新供应商
func (h *FinanceHandler) UpdateSupplier(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.UpdateSupplier(tenantID, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteSupplier 删除供应商
func (h *FinanceHandler) DeleteSupplier(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.financeSvc.DeleteSupplier(tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== 采购结算 ====================

// ListPurchaseSettlements 获取采购结算单列表
func (h *FinanceHandler) ListPurchaseSettlements(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.SettlementFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}
	if supplierID, err := strconv.ParseInt(c.Query("supplier_id"), 10, 64); err == nil {
		filter.SupplierID = &supplierID
	}

	settlements, total, err := h.financeSvc.ListPurchaseSettlements(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settlements": settlements,
		"total":       total,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	})
}

// CreatePurchaseSettlement 创建采购结算单
func (h *FinanceHandler) CreatePurchaseSettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)

	var settlement models.PurchaseSettlement
	if err := c.ShouldBindJSON(&settlement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settlement.TenantID = tenantID
	settlement.CreatedBy = userID

	if err := h.financeSvc.CreatePurchaseSettlement(&settlement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"settlement": settlement})
}

// GetPurchaseSettlement 获取采购结算单详情
func (h *FinanceHandler) GetPurchaseSettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	settlement, err := h.financeSvc.GetPurchaseSettlement(tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "settlement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlement": settlement})
}

// PaySettlement 结算单付款
func (h *FinanceHandler) PaySettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var payment models.PurchasePayment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.PaySettlement(tenantID, id, userID, &payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment recorded", "payment": payment})
}

// GetSettlementPayments 获取结算单付款记录
func (h *FinanceHandler) GetSettlementPayments(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	payments, err := h.financeSvc.GetSettlementPayments(tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// ==================== 商品成本 ====================

// ListProductCosts 获取商品成本列表
func (h *FinanceHandler) ListProductCosts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.ProductCostFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if productSku := c.Query("product_sku"); productSku != "" {
		filter.ProductSku = productSku
	}

	costs, total, err := h.financeSvc.ListProductCosts(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"costs":     costs,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// UpdateProductCost 更新商品成本
func (h *FinanceHandler) UpdateProductCost(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.UpdateProductCost(tenantID, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// BatchUpdateProductCosts 批量更新商品成本
func (h *FinanceHandler) BatchUpdateProductCosts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var req struct {
		Costs []models.ProductCost `json:"costs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.BatchUpdateProductCosts(tenantID, req.Costs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch updated"})
}

// ==================== 订单成本 ====================

// ListOrderCosts 获取订单成本列表
func (h *FinanceHandler) ListOrderCosts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.OrderCostFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if orderNo := c.Query("order_no"); orderNo != "" {
		filter.OrderNo = orderNo
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}

	costs, total, err := h.financeSvc.ListOrderCosts(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"costs":     costs,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// CalculateOrderCost 计算订单成本
func (h *FinanceHandler) CalculateOrderCost(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cost, err := h.financeSvc.CalculateOrderCost(tenantID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cost": cost})
}

// GetProfitAnalysis 获取利润分析
func (h *FinanceHandler) GetProfitAnalysis(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	startDate, _ := time.Parse("2006-01-02", c.Query("start_date"))
	endDate, _ := time.Parse("2006-01-02", c.Query("end_date"))

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, -1, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	analysis, err := h.financeSvc.GetProfitAnalysis(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

// ==================== 库存成本快照 ====================

// ListInventorySnapshots 获取库存成本快照列表
func (h *FinanceHandler) ListInventorySnapshots(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.SnapshotFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if warehouseID, err := strconv.ParseInt(c.Query("warehouse_id"), 10, 64); err == nil {
		filter.WarehouseID = &warehouseID
	}
	if productID, err := strconv.ParseInt(c.Query("product_id"), 10, 64); err == nil {
		filter.ProductID = &productID
	}
	if snapshotDate, err := time.Parse("2006-01-02", c.Query("snapshot_date")); err == nil {
		filter.SnapshotDate = snapshotDate
	}

	snapshots, total, err := h.financeSvc.ListInventoryCostSnapshots(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots,
		"total":     total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

// GenerateInventorySnapshot 生成库存成本快照
func (h *FinanceHandler) GenerateInventorySnapshot(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var req struct {
		Date string `json:"date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	if err := h.financeSvc.GenerateInventorySnapshot(tenantID, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "snapshot generated"})
}

// ==================== 财务结算 ====================

// ListMonthlySettlements 获取月度结算列表
func (h *FinanceHandler) ListMonthlySettlements(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.FinancialSettlementFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}

	settlements, total, err := h.financeSvc.ListMonthlySettlements(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settlements": settlements,
		"total":       total,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	})
}

// GenerateMonthlySettlement 生成月度结算
func (h *FinanceHandler) GenerateMonthlySettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var req struct {
		Period string `json:"period"`
		ShopID *int64 `json:"shop_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settlement, err := h.financeSvc.GenerateMonthlySettlement(tenantID, req.Period, req.ShopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlement": settlement})
}

// ConfirmMonthlySettlement 确认月度结算
func (h *FinanceHandler) ConfirmMonthlySettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	period := c.Param("period")

	if err := h.financeSvc.ConfirmMonthlySettlement(tenantID, userID, period); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "confirmed"})
}

// ListYearlySettlements 获取年度结算列表
func (h *FinanceHandler) ListYearlySettlements(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	filter := &repository.FinancialSettlementFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if status, err := strconv.Atoi(c.Query("status")); err == nil {
		filter.Status = status
	} else {
		filter.Status = -1
	}

	settlements, total, err := h.financeSvc.ListYearlySettlements(tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settlements": settlements,
		"total":       total,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	})
}

// GenerateYearlySettlement 生成年度结算
func (h *FinanceHandler) GenerateYearlySettlement(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var req struct {
		Year string `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settlement, err := h.financeSvc.GenerateYearlySettlement(tenantID, req.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlement": settlement})
}

// ==================== 结算账户 ====================

// ListBankAccounts 获取结算账户列表
func (h *FinanceHandler) ListBankAccounts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	accounts, err := h.financeSvc.ListBankAccounts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// CreateBankAccount 创建结算账户
func (h *FinanceHandler) CreateBankAccount(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)

	var account models.FinanceBankAccount
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.TenantID = tenantID

	if err := h.financeSvc.CreateBankAccount(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"account": account})
}

// UpdateBankAccount 更新结算账户
func (h *FinanceHandler) UpdateBankAccount(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeSvc.UpdateBankAccount(tenantID, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteBankAccount 删除结算账户
func (h *FinanceHandler) DeleteBankAccount(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.financeSvc.DeleteBankAccount(tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== 辅助函数 ====================

// GetTenantIDFromGin 从上下文获取租户ID
func GetTenantIDFromGin(c *gin.Context) int64 {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return 0
	}
	return tenantID.(int64)
}

// GetUserIDFromGin 从上下文获取用户ID
func GetUserIDFromGin(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int64)
}
