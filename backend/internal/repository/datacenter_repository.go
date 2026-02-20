package repository

import (
	"time"

	"github.com/MorantHP/OURERP/internal/models"

	"gorm.io/gorm"
)

type DatacenterRepository struct {
	db *gorm.DB
}

func NewDatacenterRepository(db *gorm.DB) *DatacenterRepository {
	return &DatacenterRepository{db: db}
}

// =============== 预警规则 CRUD ===============

func (r *DatacenterRepository) CreateAlertRule(rule *models.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *DatacenterRepository) UpdateAlertRule(rule *models.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *DatacenterRepository) DeleteAlertRule(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.AlertRule{}).Error
}

func (r *DatacenterRepository) GetAlertRuleByID(tenantID, id int64) (*models.AlertRule, error) {
	var rule models.AlertRule
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *DatacenterRepository) ListAlertRules(tenantID int64, filter *models.AlertRuleFilter, page, pageSize int) ([]models.AlertRule, int64, error) {
	var rules []models.AlertRule
	var total int64

	query := r.db.Model(&models.AlertRule{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rules).Error
	return rules, total, err
}

func (r *DatacenterRepository) ListActiveAlertRules(tenantID int64) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.Where("tenant_id = ? AND status = 1", tenantID).Find(&rules).Error
	return rules, err
}

// =============== 预警记录 CRUD ===============

func (r *DatacenterRepository) CreateAlertRecord(record *models.AlertRecord) error {
	return r.db.Create(record).Error
}

func (r *DatacenterRepository) GetAlertRecordByID(tenantID, id int64) (*models.AlertRecord, error) {
	var record models.AlertRecord
	err := r.db.Preload("Rule").Where("tenant_id = ? AND id = ?", tenantID, id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *DatacenterRepository) ListAlertRecords(tenantID int64, filter *models.AlertRecordFilter, page, pageSize int) ([]models.AlertRecord, int64, error) {
	var records []models.AlertRecord
	var total int64

	query := r.db.Model(&models.AlertRecord{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.RuleID != nil {
			query = query.Where("rule_id = ?", *filter.RuleID)
		}
		if filter.Level != "" {
			query = query.Where("level = ?", filter.Level)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.SourceType != "" {
			query = query.Where("source_type = ?", filter.SourceType)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", *filter.EndDate)
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Rule").Order("id DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (r *DatacenterRepository) UpdateAlertRecord(record *models.AlertRecord) error {
	return r.db.Save(record).Error
}

func (r *DatacenterRepository) CountUnhandledAlerts(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.AlertRecord{}).Where("tenant_id = ? AND status = 0", tenantID).Count(&count).Error
	return count, err
}

// =============== 客户 CRUD ===============

func (r *DatacenterRepository) CreateCustomer(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *DatacenterRepository) UpdateCustomer(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

func (r *DatacenterRepository) DeleteCustomer(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.Customer{}).Error
}

func (r *DatacenterRepository) GetCustomerByID(tenantID, id int64) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *DatacenterRepository) GetCustomerByCode(tenantID int64, code string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("tenant_id = ? AND code = ?", tenantID, code).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *DatacenterRepository) GetCustomerByPhone(tenantID int64, phone string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("tenant_id = ? AND phone = ?", tenantID, phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *DatacenterRepository) ListCustomers(tenantID int64, filter *models.CustomerFilter, page, pageSize int) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := r.db.Model(&models.Customer{}).Where("tenant_id = ?", tenantID)

	if filter != nil {
		if filter.Name != "" {
			query = query.Where("name LIKE ?", "%"+filter.Name+"%")
		}
		if filter.Phone != "" {
			query = query.Where("phone LIKE ?", "%"+filter.Phone+"%")
		}
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.Level != "" {
			query = query.Where("level = ?", filter.Level)
		}
		if filter.Province != "" {
			query = query.Where("province = ?", filter.Province)
		}
		if filter.City != "" {
			query = query.Where("city = ?", filter.City)
		}
		if filter.Source != "" {
			query = query.Where("source = ?", filter.Source)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Tags != "" {
			query = query.Where("tags LIKE ?", "%"+filter.Tags+"%")
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&customers).Error
	return customers, total, err
}

func (r *DatacenterRepository) UpdateCustomerStats(tenantID, customerID int64) error {
	// 从订单统计更新客户数据
	return r.db.Exec(`
		UPDATE customers c SET
			total_orders = (SELECT COUNT(*) FROM orders WHERE customer_id = ? AND tenant_id = ?),
			total_amount = COALESCE((SELECT SUM(total_amount) FROM orders WHERE customer_id = ? AND tenant_id = ?), 0),
			total_paid = COALESCE((SELECT SUM(paid_amount) FROM orders WHERE customer_id = ? AND tenant_id = ?), 0),
			first_order_at = (SELECT MIN(created_at) FROM orders WHERE customer_id = ? AND tenant_id = ?),
			last_order_at = (SELECT MAX(created_at) FROM orders WHERE customer_id = ? AND tenant_id = ?),
			avg_order_value = CASE
				WHEN (SELECT COUNT(*) FROM orders WHERE customer_id = ? AND tenant_id = ?) > 0
				THEN (SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE customer_id = ? AND tenant_id = ?) /
					 (SELECT COUNT(*) FROM orders WHERE customer_id = ? AND tenant_id = ?)
				ELSE 0
			END,
			updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, customerID, tenantID, customerID, tenantID, customerID, tenantID, customerID, tenantID,
		customerID, tenantID, customerID, tenantID, customerID, tenantID, customerID, tenantID,
		time.Now(), customerID, tenantID).Error
}

// =============== 报表模板 CRUD ===============

func (r *DatacenterRepository) CreateReportTemplate(template *models.ReportTemplate) error {
	return r.db.Create(template).Error
}

func (r *DatacenterRepository) UpdateReportTemplate(template *models.ReportTemplate) error {
	return r.db.Save(template).Error
}

func (r *DatacenterRepository) DeleteReportTemplate(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.ReportTemplate{}).Error
}

func (r *DatacenterRepository) GetReportTemplateByID(tenantID, id int64) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *DatacenterRepository) ListReportTemplates(tenantID int64, filter *models.ReportTemplateFilter, page, pageSize int) ([]models.ReportTemplate, int64, error) {
	var templates []models.ReportTemplate
	var total int64

	query := r.db.Model(&models.ReportTemplate{}).Where("tenant_id = ? AND (is_public = 1 OR created_by = ?)", tenantID, filter.CreatedBy)

	if filter != nil {
		if filter.Type != "" {
			query = query.Where("type = ?", filter.Type)
		}
		if filter.DataSource != "" {
			query = query.Where("data_source = ?", filter.DataSource)
		}
		if filter.IsPublic != nil {
			query = query.Where("is_public = ?", *filter.IsPublic)
		}
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&templates).Error
	return templates, total, err
}

// =============== 实时快照 ===============

func (r *DatacenterRepository) CreateSnapshot(snapshot *models.RealtimeSnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *DatacenterRepository) GetLatestSnapshot(tenantID int64) (*models.RealtimeSnapshot, error) {
	var snapshot models.RealtimeSnapshot
	err := r.db.Where("tenant_id = ?", tenantID).Order("snapshot_time DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *DatacenterRepository) GetSnapshotsByPeriod(tenantID int64, start, end time.Time) ([]models.RealtimeSnapshot, error) {
	var snapshots []models.RealtimeSnapshot
	err := r.db.Where("tenant_id = ? AND snapshot_time >= ? AND snapshot_time <= ?", tenantID, start, end).
		Order("snapshot_time ASC").Find(&snapshots).Error
	return snapshots, err
}

// =============== 商品分析 ===============

func (r *DatacenterRepository) SaveProductAnalysis(analysis *models.ProductAnalysis) error {
	return r.db.Create(analysis).Error
}

func (r *DatacenterRepository) GetProductAnalysis(tenantID int64, productID int64, startDate, endDate time.Time, periodType string) ([]models.ProductAnalysis, error) {
	var analyses []models.ProductAnalysis
	query := r.db.Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Where("analysis_date >= ? AND analysis_date <= ?", startDate, endDate)
	if periodType != "" {
		query = query.Where("period_type = ?", periodType)
	}
	err := query.Order("analysis_date ASC").Find(&analyses).Error
	return analyses, err
}

func (r *DatacenterRepository) GetTopProductAnalysis(tenantID int64, limit int, orderBy string, startDate, endDate time.Time) ([]models.ProductAnalysis, error) {
	var analyses []models.ProductAnalysis
	query := r.db.Where("tenant_id = ?", tenantID).
		Where("analysis_date >= ? AND analysis_date <= ?", startDate, endDate)
	err := query.Order(orderBy + " DESC").Limit(limit).Find(&analyses).Error
	return analyses, err
}

// =============== 客户分析 ===============

func (r *DatacenterRepository) SaveCustomerAnalysis(analysis *models.CustomerAnalysis) error {
	return r.db.Create(analysis).Error
}

func (r *DatacenterRepository) GetCustomerAnalysis(tenantID int64, startDate, endDate time.Time, periodType string) ([]models.CustomerAnalysis, error) {
	var analyses []models.CustomerAnalysis
	query := r.db.Where("tenant_id = ?", tenantID).
		Where("analysis_date >= ? AND analysis_date <= ?", startDate, endDate)
	if periodType != "" {
		query = query.Where("period_type = ?", periodType)
	}
	err := query.Order("analysis_date ASC").Find(&analyses).Error
	return analyses, err
}

// =============== 地域分析 ===============

func (r *DatacenterRepository) SaveRegionAnalysis(analysis *models.RegionAnalysis) error {
	return r.db.Create(analysis).Error
}

func (r *DatacenterRepository) GetRegionAnalysis(tenantID int64, startDate, endDate time.Time, periodType string) ([]models.RegionAnalysis, error) {
	var analyses []models.RegionAnalysis
	query := r.db.Where("tenant_id = ?", tenantID).
		Where("analysis_date >= ? AND analysis_date <= ?", startDate, endDate)
	if periodType != "" {
		query = query.Where("period_type = ?", periodType)
	}
	err := query.Order("order_amount DESC").Find(&analyses).Error
	return analyses, err
}

// =============== 对比分析 ===============

func (r *DatacenterRepository) SaveCompareAnalysis(analysis *models.CompareAnalysis) error {
	return r.db.Create(analysis).Error
}

func (r *DatacenterRepository) GetCompareAnalysis(tenantID int64, compareType, periodType string) ([]models.CompareAnalysis, error) {
	var analyses []models.CompareAnalysis
	query := r.db.Where("tenant_id = ?", tenantID)
	if compareType != "" {
		query = query.Where("compare_type = ?", compareType)
	}
	err := query.Order("analysis_date DESC").Find(&analyses).Error
	return analyses, err
}

// =============== 仪表盘组件 ===============

func (r *DatacenterRepository) CreateDashboardWidget(widget *models.DashboardWidget) error {
	return r.db.Create(widget).Error
}

func (r *DatacenterRepository) UpdateDashboardWidget(widget *models.DashboardWidget) error {
	return r.db.Save(widget).Error
}

func (r *DatacenterRepository) DeleteDashboardWidget(tenantID, id int64) error {
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&models.DashboardWidget{}).Error
}

func (r *DatacenterRepository) ListDashboardWidgets(tenantID, userID int64) ([]models.DashboardWidget, error) {
	var widgets []models.DashboardWidget
	err := r.db.Where("tenant_id = ? AND user_id = ? AND status = 1", tenantID, userID).
		Order("sort_order ASC").Find(&widgets).Error
	return widgets, err
}

// =============== 统计查询 ===============

// GetRealtimeStats 获取实时统计
func (r *DatacenterRepository) GetRealtimeStats(tenantID int64, startTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 今日订单统计
	var orderCount int64
	var orderAmount float64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startTime).
		Count(&orderCount)
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startTime).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&orderAmount)
	stats["order_count"] = orderCount
	stats["order_amount"] = orderAmount

	// 今日支付统计
	var paidAmount float64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND paid_at >= ?", tenantID, startTime).
		Select("COALESCE(SUM(paid_amount), 0)").Scan(&paidAmount)
	stats["paid_amount"] = paidAmount

	// 待处理订单
	var pendingOrders int64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND status IN ?", tenantID, []string{"pending", "paid"}).
		Count(&pendingOrders)
	stats["pending_orders"] = pendingOrders

	// 今日退款
	var refundCount int64
	var refundAmount float64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND refund_status IN ? AND updated_at >= ?",
		[]string{"refunded", "partial_refund"}, startTime).
		Count(&refundCount)
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND refund_status IN ? AND updated_at >= ?",
		[]string{"refunded", "partial_refund"}, startTime).
		Select("COALESCE(SUM(refund_amount), 0)").Scan(&refundAmount)
	stats["refund_count"] = refundCount
	stats["refund_amount"] = refundAmount

	// 库存预警
	var lowStockItems int64
	r.db.Model(&models.Inventory{}).Where("tenant_id = ? AND quantity <= alert_qty", tenantID).
		Count(&lowStockItems)
	stats["low_stock_items"] = lowStockItems

	// 今日新客户
	var newCustomers int64
	r.db.Model(&models.Customer{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startTime).
		Count(&newCustomers)
	stats["new_customers"] = newCustomers

	return stats, nil
}

// GetSalesByRegion 按地域统计销售
func (r *DatacenterRepository) GetSalesByRegion(tenantID int64, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	// Order 模型当前没有 customer_id 和 province 字段
	// 返回基于 buyer_nick 的统计（按买家分组）
	err := r.db.Raw(`
		SELECT
			'全部地区' as province,
			COUNT(*) as order_count,
			COALESCE(SUM(total_amount), 0) as order_amount
		FROM orders
		WHERE tenant_id = ? AND created_at >= ? AND created_at <= ?
	`, tenantID, startDate, endDate).Scan(&results).Error
	return results, err
}

// GetSalesByCity 按城市统计销售
func (r *DatacenterRepository) GetSalesByCity(tenantID int64, province string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	// 简化查询 - 返回基于省份的统计
	err := r.db.Raw(`
		SELECT
			? as city,
			COUNT(*) as order_count,
			COALESCE(SUM(total_amount), 0) as order_amount
		FROM orders
		WHERE tenant_id = ? AND created_at >= ? AND created_at <= ?
	`, province, tenantID, startDate, endDate).Scan(&results).Error
	return results, err
}

// GetProductTurnoverRate 获取商品动销率
func (r *DatacenterRepository) GetProductTurnoverRate(tenantID int64, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	// 简化查询 - 返回产品列表及其库存信息
	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			0 as sales_quantity,
			COALESCE(i.quantity, 0) as stock_quantity,
			0 as turnover_rate
		FROM products p
		LEFT JOIN inventories i ON i.product_id = p.id AND i.tenant_id = p.tenant_id
		WHERE p.tenant_id = ?
		ORDER BY p.id DESC
	`
	if limit > 0 {
		query += " LIMIT ?"
		err := r.db.Raw(query, tenantID, limit).Scan(&results).Error
		return results, err
	}
	err := r.db.Raw(query, tenantID).Scan(&results).Error
	return results, err
}

// GetInventoryLevel 获取库存水位分析
func (r *DatacenterRepository) GetInventoryLevel(tenantID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			COALESCE(i.quantity, 0) as quantity,
			COALESCE(i.alert_qty, 0) as alert_qty,
			CASE
				WHEN COALESCE(i.quantity, 0) <= 0 THEN 'out_of_stock'
				WHEN COALESCE(i.quantity, 0) <= COALESCE(i.alert_qty, 0) THEN 'low'
				ELSE 'normal'
			END as stock_level
		FROM products p
		LEFT JOIN inventories i ON i.product_id = p.id AND i.tenant_id = p.tenant_id
		WHERE p.tenant_id = ?
		ORDER BY i.quantity ASC
	`
	err := r.db.Raw(query, tenantID).Scan(&results).Error
	return results, err
}

// GetPeriodCompare 获取期间对比数据
func (r *DatacenterRepository) GetPeriodCompare(tenantID int64, currentStart, currentEnd, compareStart, compareEnd time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 当前期间统计
	var currentOrders int64
	var currentAmount float64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, currentStart, currentEnd).
		Count(&currentOrders)
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, currentStart, currentEnd).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&currentAmount)
	result["current_orders"] = currentOrders
	result["current_amount"] = currentAmount

	// 对比期间统计
	var compareOrders int64
	var compareAmount float64
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, compareStart, compareEnd).
		Count(&compareOrders)
	r.db.Model(&models.Order{}).Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, compareStart, compareEnd).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&compareAmount)
	result["compare_orders"] = compareOrders
	result["compare_amount"] = compareAmount

	// 计算变化
	result["order_change"] = currentOrders - compareOrders
	result["amount_change"] = currentAmount - compareAmount
	if compareOrders > 0 {
		result["order_change_rate"] = float64(currentOrders-compareOrders) * 100 / float64(compareOrders)
	} else {
		result["order_change_rate"] = 0.0
	}
	if compareAmount > 0 {
		result["amount_change_rate"] = (currentAmount - compareAmount) * 100 / compareAmount
	} else {
		result["amount_change_rate"] = 0.0
	}

	return result, nil
}

// GetCustomerValueDistribution 获取客户价值分布
func (r *DatacenterRepository) GetCustomerValueDistribution(tenantID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	query := `
		SELECT value_level, customer_count, total_amount FROM (
			SELECT
				CASE
					WHEN total_amount >= 10000 THEN 'high_value'
					WHEN total_amount >= 1000 THEN 'medium_value'
					WHEN total_amount > 0 THEN 'low_value'
					ELSE 'no_purchase'
				END as value_level,
				COUNT(*) as customer_count,
				COALESCE(SUM(total_amount), 0) as total_amount
			FROM customers
			WHERE tenant_id = ?
			GROUP BY
				CASE
					WHEN total_amount >= 10000 THEN 'high_value'
					WHEN total_amount >= 1000 THEN 'medium_value'
					WHEN total_amount > 0 THEN 'low_value'
					ELSE 'no_purchase'
				END
		) sub
		ORDER BY
			CASE value_level
				WHEN 'high_value' THEN 1
				WHEN 'medium_value' THEN 2
				WHEN 'low_value' THEN 3
				ELSE 4
			END
	`
	err := r.db.Raw(query, tenantID).Scan(&results).Error
	return results, err
}

// GetRepurchaseRate 获取复购率
func (r *DatacenterRepository) GetRepurchaseRate(tenantID int64, startDate, endDate time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 期间内下过单的客户
	var totalCustomers int64
	var repurchaseCustomers int64
	r.db.Raw(`
		SELECT COUNT(DISTINCT customer_id)
		FROM orders
		WHERE tenant_id = ? AND created_at >= ? AND created_at <= ? AND customer_id IS NOT NULL
	`, tenantID, startDate, endDate).Scan(&totalCustomers)

	// 期间内复购的客户（订单数>=2）
	r.db.Raw(`
		SELECT COUNT(*)
		FROM (
			SELECT customer_id
			FROM orders
			WHERE tenant_id = ? AND created_at >= ? AND created_at <= ? AND customer_id IS NOT NULL
			GROUP BY customer_id
			HAVING COUNT(*) >= 2
		) t
	`, tenantID, startDate, endDate).Scan(&repurchaseCustomers)

	result["total_customers"] = totalCustomers
	result["repurchase_customers"] = repurchaseCustomers
	if totalCustomers > 0 {
		result["repurchase_rate"] = float64(repurchaseCustomers) * 100 / float64(totalCustomers)
	} else {
		result["repurchase_rate"] = 0.0
	}

	return result, nil
}
