package services

import (
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

type AlertService struct {
	repo         *repository.DatacenterRepository
	productRepo  *repository.ProductRepository
	orderRepo    *repository.OrderRepository
}

func NewAlertService(repo *repository.DatacenterRepository, productRepo *repository.ProductRepository, orderRepo *repository.OrderRepository) *AlertService {
	return &AlertService{
		repo:        repo,
		productRepo: productRepo,
		orderRepo:   orderRepo,
	}
}

// AlertSummary 预警汇总
type AlertSummary struct {
	TotalAlerts     int `json:"total_alerts"`
	UnhandledAlerts int `json:"unhandled_alerts"`
	CriticalCount   int `json:"critical_count"`
	WarningCount    int `json:"warning_count"`
	InfoCount       int `json:"info_count"`
	TodayAlerts     int `json:"today_alerts"`
}

// CreateRule 创建预警规则
func (s *AlertService) CreateRule(rule *models.AlertRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	return s.repo.CreateAlertRule(rule)
}

// UpdateRule 更新预警规则
func (s *AlertService) UpdateRule(rule *models.AlertRule) error {
	rule.UpdatedAt = time.Now()
	return s.repo.UpdateAlertRule(rule)
}

// DeleteRule 删除预警规则
func (s *AlertService) DeleteRule(tenantID, ruleID int64) error {
	return s.repo.DeleteAlertRule(tenantID, ruleID)
}

// GetRule 获取预警规则
func (s *AlertService) GetRule(tenantID, ruleID int64) (*models.AlertRule, error) {
	return s.repo.GetAlertRuleByID(tenantID, ruleID)
}

// ListRules 列出预警规则
func (s *AlertService) ListRules(tenantID int64, filter *models.AlertRuleFilter, page, pageSize int) ([]models.AlertRule, int64, error) {
	return s.repo.ListAlertRules(tenantID, filter, page, pageSize)
}

// ToggleRule 启用/停用预警规则
func (s *AlertService) ToggleRule(tenantID, ruleID int64, status int) error {
	rule, err := s.repo.GetAlertRuleByID(tenantID, ruleID)
	if err != nil {
		return err
	}
	rule.Status = status
	rule.UpdatedAt = time.Now()
	return s.repo.UpdateAlertRule(rule)
}

// GetAlertSummary 获取预警汇总
func (s *AlertService) GetAlertSummary(tenantID int64) (*AlertSummary, error) {
	summary := &AlertSummary{}

	// 获取未处理预警数
	unhandled, err := s.repo.CountUnhandledAlerts(tenantID)
	if err != nil {
		return nil, err
	}
	summary.UnhandledAlerts = int(unhandled)

	// 获取今日预警
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	filter := &models.AlertRecordFilter{
		StartDate: &todayStart,
	}
	records, total, err := s.repo.ListAlertRecords(tenantID, filter, 1, 1000)
	if err != nil {
		return nil, err
	}

	summary.TotalAlerts = int(total)
	summary.TodayAlerts = int(total)

	for _, record := range records {
		switch record.Level {
		case "critical":
			summary.CriticalCount++
		case "warning":
			summary.WarningCount++
		case "info":
			summary.InfoCount++
		}
	}

	return summary, nil
}

// ListAlertRecords 列出预警记录
func (s *AlertService) ListAlertRecords(tenantID int64, filter *models.AlertRecordFilter, page, pageSize int) ([]models.AlertRecord, int64, error) {
	return s.repo.ListAlertRecords(tenantID, filter, page, pageSize)
}

// HandleAlert 处理预警
func (s *AlertService) HandleAlert(tenantID, recordID, handlerID int64, note string) error {
	record, err := s.repo.GetAlertRecordByID(tenantID, recordID)
	if err != nil {
		return err
	}

	now := time.Now()
	record.Status = 1
	record.HandledBy = &handlerID
	record.HandledAt = &now
	record.HandleNote = note

	return s.repo.UpdateAlertRecord(record)
}

// IgnoreAlert 忽略预警
func (s *AlertService) IgnoreAlert(tenantID, recordID, handlerID int64, note string) error {
	record, err := s.repo.GetAlertRecordByID(tenantID, recordID)
	if err != nil {
		return err
	}

	now := time.Now()
	record.Status = 2
	record.HandledBy = &handlerID
	record.HandledAt = &now
	record.HandleNote = note

	return s.repo.UpdateAlertRecord(record)
}

// CheckAlerts 检查预警
func (s *AlertService) CheckAlerts(tenantID int64) ([]models.AlertRecord, error) {
	// 获取所有启用的规则
	rules, err := s.repo.ListActiveAlertRules(tenantID)
	if err != nil {
		return nil, err
	}

	var newAlerts []models.AlertRecord

	for _, rule := range rules {
		switch rule.Type {
		case "inventory":
			alerts, err := s.checkInventoryAlert(tenantID, &rule)
			if err == nil {
				newAlerts = append(newAlerts, alerts...)
			}
		case "sales":
			alerts, err := s.checkSalesAlert(tenantID, &rule)
			if err == nil {
				newAlerts = append(newAlerts, alerts...)
			}
		case "order":
			alerts, err := s.checkOrderAlert(tenantID, &rule)
			if err == nil {
				newAlerts = append(newAlerts, alerts...)
			}
		}
	}

	return newAlerts, nil
}

// checkInventoryAlert 检查库存预警
func (s *AlertService) checkInventoryAlert(tenantID int64, rule *models.AlertRule) ([]models.AlertRecord, error) {
	var alerts []models.AlertRecord

	// 获取库存水位
	levels, err := s.repo.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	for _, item := range levels {
		var productID int64
		var productName string
		var quantity int
		var stockLevel string

		if v, ok := item["product_id"].(int64); ok {
			productID = v
		}
		if v, ok := item["product_name"].(string); ok {
			productName = v
		}
		if v, ok := item["quantity"].(int64); ok {
			quantity = int(v)
		}
		if v, ok := item["stock_level"].(string); ok {
			stockLevel = v
		}

		// 检查是否触发规则
		shouldAlert := false
		if stockLevel == "out_of_stock" {
			shouldAlert = true
		} else if stockLevel == "low" && rule.Threshold > 0 && float64(quantity) <= rule.Threshold {
			shouldAlert = true
		}

		if shouldAlert {
			alert := models.AlertRecord{
				TenantID:   tenantID,
				RuleID:     rule.ID,
				Title:      "库存预警: " + productName,
				Content:    "商品【" + productName + "】库存不足，当前库存: " + string(rune(quantity)),
				Level:      rule.Level,
				SourceType: "product",
				SourceID:   productID,
				Status:     0,
				CreatedAt:  time.Now(),
			}
			alerts = append(alerts, alert)

			// 保存到数据库
			s.repo.CreateAlertRecord(&alert)
		}
	}

	return alerts, nil
}

// checkSalesAlert 检查销售预警
func (s *AlertService) checkSalesAlert(tenantID int64, rule *models.AlertRule) ([]models.AlertRecord, error) {
	var alerts []models.AlertRecord

	// 获取今日销售数据
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats, err := s.repo.GetRealtimeStats(tenantID, todayStart)
	if err != nil {
		return nil, err
	}

	// 检查销售额是否低于阈值
	if rule.Threshold > 0 {
		if orderAmount, ok := stats["order_amount"].(float64); ok {
			if orderAmount < rule.Threshold {
				alert := models.AlertRecord{
					TenantID:   tenantID,
					RuleID:     rule.ID,
					Title:      "销售预警: 日销售额低于预期",
					Content:    "今日销售额为 ¥" + string(rune(int64(orderAmount))) + "，低于设定阈值 ¥" + string(rune(int64(rule.Threshold))),
					Level:      rule.Level,
					SourceType: "sales",
					SourceID:   0,
					Status:     0,
					CreatedAt:  time.Now(),
				}
				alerts = append(alerts, alert)
				s.repo.CreateAlertRecord(&alert)
			}
		}
	}

	return alerts, nil
}

// checkOrderAlert 检查订单预警
func (s *AlertService) checkOrderAlert(tenantID int64, rule *models.AlertRule) ([]models.AlertRecord, error) {
	var alerts []models.AlertRecord

	// 检查待处理订单超时等情况
	// 这里简化处理，实际应该查询超时订单

	return alerts, nil
}

// NotifyAlert 发送预警通知
func (s *AlertService) NotifyAlert(alert *models.AlertRecord, rule *models.AlertRule) error {
	switch rule.NotifyType {
	case "email":
		// 发送邮件通知
		return s.sendEmailNotification(alert, rule)
	case "sms":
		// 发送短信通知
		return s.sendSMSNotification(alert, rule)
	case "webhook":
		// 发送Webhook通知
		return s.sendWebhookNotification(alert, rule)
	case "system":
		// 系统内通知，已经在数据库中记录
		return nil
	}
	return nil
}

// sendEmailNotification 发送邮件通知
func (s *AlertService) sendEmailNotification(alert *models.AlertRecord, rule *models.AlertRule) error {
	// 这里应该调用邮件服务发送邮件
	// 简化处理，暂不实现
	return nil
}

// sendSMSNotification 发送短信通知
func (s *AlertService) sendSMSNotification(alert *models.AlertRecord, rule *models.AlertRule) error {
	// 这里应该调用短信服务发送短信
	// 简化处理，暂不实现
	return nil
}

// sendWebhookNotification 发送Webhook通知
func (s *AlertService) sendWebhookNotification(alert *models.AlertRecord, rule *models.AlertRule) error {
	// 这里应该调用Webhook URL发送通知
	// 简化处理，暂不实现
	return nil
}

// AlertType 预警类型
type AlertType struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetAlertTypes 获取预警类型列表
func (s *AlertService) GetAlertTypes() []AlertType {
	return []AlertType{
		{Type: "inventory", Name: "库存预警", Description: "当库存低于阈值时触发"},
		{Type: "sales", Name: "销售预警", Description: "当销售额低于阈值时触发"},
		{Type: "order", Name: "订单预警", Description: "当订单状态异常时触发"},
		{Type: "customer", Name: "客户预警", Description: "当客户行为异常时触发"},
		{Type: "finance", Name: "财务预警", Description: "当财务指标异常时触发"},
	}
}

// NotifyLevel 预警级别
type NotifyLevel struct {
	Level       string `json:"level"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// GetNotifyLevels 获取预警级别列表
func (s *AlertService) GetNotifyLevels() []NotifyLevel {
	return []NotifyLevel{
		{Level: "info", Name: "信息", Color: "#909399", Description: "一般性通知"},
		{Level: "warning", Name: "警告", Color: "#E6A23C", Description: "需要关注的问题"},
		{Level: "critical", Name: "严重", Color: "#F56C6C", Description: "需要立即处理的问题"},
	}
}
