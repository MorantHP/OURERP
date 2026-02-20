package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/services"

	"github.com/gin-gonic/gin"
)

type DatacenterHandler struct {
	realtimeSvc  *services.RealtimeService
	customerSvc  *services.CustomerAnalysisService
	productSvc   *services.ProductAnalysisService
	compareSvc   *services.CompareAnalysisService
	alertSvc     *services.AlertService
}

func NewDatacenterHandler(
	realtimeSvc *services.RealtimeService,
	customerSvc *services.CustomerAnalysisService,
	productSvc *services.ProductAnalysisService,
	compareSvc *services.CompareAnalysisService,
	alertSvc *services.AlertService,
) *DatacenterHandler {
	return &DatacenterHandler{
		realtimeSvc: realtimeSvc,
		customerSvc: customerSvc,
		productSvc:  productSvc,
		compareSvc:  compareSvc,
		alertSvc:    alertSvc,
	}
}

// =============== 实时监控 API ===============

// GetRealtimeOverview 获取实时概览
func (h *DatacenterHandler) GetRealtimeOverview(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	overview, err := h.realtimeSvc.GetOverview(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

// GetRealtimeInventory 获取实时库存状态
func (h *DatacenterHandler) GetRealtimeInventory(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	status, err := h.realtimeSvc.GetInventoryStatus(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// GetHourlyTrend 获取小时趋势
func (h *DatacenterHandler) GetHourlyTrend(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	trends, err := h.realtimeSvc.GetHourlyTrend(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

// =============== 客户分析 API ===============

// GetCustomerAnalysis 获取客户分析
func (h *DatacenterHandler) GetCustomerAnalysis(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	analysis, err := h.customerSvc.GetCustomerAnalysis(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

// GetCustomerValueDistribution 获取客户价值分布
func (h *DatacenterHandler) GetCustomerValueDistribution(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	distribution, err := h.customerSvc.GetValueDistribution(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

// GetGeographyDistribution 获取地域分布
func (h *DatacenterHandler) GetGeographyDistribution(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	distribution, err := h.customerSvc.GetGeographyDistribution(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

// GetCityDistribution 获取城市分布
func (h *DatacenterHandler) GetCityDistribution(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	province := c.Query("province")
	startDate, endDate := parseDateRange(c)

	distribution, err := h.customerSvc.GetCityDistribution(tenantID, province, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

// GetRepurchaseAnalysis 获取复购分析
func (h *DatacenterHandler) GetRepurchaseAnalysis(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	analysis, err := h.customerSvc.GetRepurchaseAnalysis(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// =============== 商品分析 API ===============

// GetProductTurnover 获取商品动销率
func (h *DatacenterHandler) GetProductTurnover(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	turnover, err := h.productSvc.GetTurnoverRate(tenantID, startDate, endDate, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"turnover": turnover})
}

// GetInventoryLevel 获取库存水位
func (h *DatacenterHandler) GetInventoryLevel(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	level, err := h.productSvc.GetInventoryLevel(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"levels": level})
}

// GetPurchaseStrategy 获取进货策略
func (h *DatacenterHandler) GetPurchaseStrategy(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	strategy, err := h.productSvc.GetPurchaseStrategy(tenantID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"strategies": strategy})
}

// GetLowStockProducts 获取低库存商品
func (h *DatacenterHandler) GetLowStockProducts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	products, err := h.productSvc.GetLowStockProducts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetInventorySummary 获取库存汇总
func (h *DatacenterHandler) GetInventorySummary(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	summary, err := h.productSvc.GetInventorySummary(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// =============== 对比分析 API ===============

// GetPeriodCompare 获取期间对比
func (h *DatacenterHandler) GetPeriodCompare(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	currentStart, currentEnd := parseDateRangeWithPrefix(c, "current_")
	compareStart, compareEnd := parseDateRangeWithPrefix(c, "compare_")

	result, err := h.compareSvc.PeriodCompare(tenantID, currentStart, currentEnd, compareStart, compareEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetYOYCompare 获取同比分析
func (h *DatacenterHandler) GetYOYCompare(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	result, err := h.compareSvc.GetYOYCompare(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMOMCompare 获取环比分析
func (h *DatacenterHandler) GetMOMCompare(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	result, err := h.compareSvc.GetMOMCompare(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetShopCompare 获取店铺对比
func (h *DatacenterHandler) GetShopCompare(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	var shopIDs []int64
	if ids := c.Query("shop_ids"); ids != "" {
		// 解析逗号分隔的ID列表
		for _, idStr := range splitIDs(ids) {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				shopIDs = append(shopIDs, id)
			}
		}
	}

	result, err := h.compareSvc.GetShopCompare(tenantID, shopIDs, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPlatformCompare 获取平台对比
func (h *DatacenterHandler) GetPlatformCompare(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate, endDate := parseDateRange(c)

	result, err := h.compareSvc.GetPlatformCompare(tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// =============== 预警管理 API ===============

// ListAlertRules 列出预警规则
func (h *DatacenterHandler) ListAlertRules(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := &models.AlertRuleFilter{
		Type: c.Query("type"),
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &status
		}
	}

	rules, total, err := h.alertSvc.ListRules(tenantID, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": total,
	})
}

// CreateAlertRule 创建预警规则
func (h *DatacenterHandler) CreateAlertRule(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var rule models.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.TenantID = tenantID
	rule.CreatedBy = userID

	if err := h.alertSvc.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rule": rule})
}

// UpdateAlertRule 更新预警规则
func (h *DatacenterHandler) UpdateAlertRule(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rule, err := h.alertSvc.GetRule(tenantID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alertSvc.UpdateRule(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rule": rule})
}

// DeleteAlertRule 删除预警规则
func (h *DatacenterHandler) DeleteAlertRule(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.alertSvc.DeleteRule(tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ToggleAlertRule 启用/停用预警规则
func (h *DatacenterHandler) ToggleAlertRule(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	status, _ := strconv.Atoi(c.PostForm("status"))

	if err := h.alertSvc.ToggleRule(tenantID, id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// GetAlertSummary 获取预警汇总
func (h *DatacenterHandler) GetAlertSummary(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	summary, err := h.alertSvc.GetAlertSummary(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// ListAlertRecords 列出预警记录
func (h *DatacenterHandler) ListAlertRecords(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := &models.AlertRecordFilter{
		Level:      c.Query("level"),
		SourceType: c.Query("source_type"),
	}

	if ruleIDStr := c.Query("rule_id"); ruleIDStr != "" {
		if ruleID, err := strconv.ParseInt(ruleIDStr, 10, 64); err == nil {
			filter.RuleID = &ruleID
		}
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &status
		}
	}

	records, total, err := h.alertSvc.ListAlertRecords(tenantID, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
	})
}

// HandleAlertRecord 处理预警记录
func (h *DatacenterHandler) HandleAlertRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alertSvc.HandleAlert(tenantID, id, userID, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处理成功"})
}

// IgnoreAlertRecord 忽略预警记录
func (h *DatacenterHandler) IgnoreAlertRecord(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.alertSvc.IgnoreAlert(tenantID, id, userID, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已忽略"})
}

// CheckAlerts 手动检查预警
func (h *DatacenterHandler) CheckAlerts(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	alerts, err := h.alertSvc.CheckAlerts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":     alerts,
		"new_count":  len(alerts),
	})
}

// GetAlertTypes 获取预警类型
func (h *DatacenterHandler) GetAlertTypes(c *gin.Context) {
	types := h.alertSvc.GetAlertTypes()
	c.JSON(http.StatusOK, gin.H{"types": types})
}

// GetNotifyLevels 获取预警级别
func (h *DatacenterHandler) GetNotifyLevels(c *gin.Context) {
	levels := h.alertSvc.GetNotifyLevels()
	c.JSON(http.StatusOK, gin.H{"levels": levels})
}

// =============== 辅助函数 ===============

func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	startStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	// 设置结束时间为当天的最后一刻
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	return start, end
}

func parseDateRangeWithPrefix(c *gin.Context, prefix string) (time.Time, time.Time) {
	startStr := c.Query(prefix + "start_date")
	endStr := c.Query(prefix + "end_date")

	if startStr == "" {
		startStr = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endStr == "" {
		endStr = time.Now().Format("2006-01-02")
	}

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	return start, end
}

func splitIDs(s string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
