package repository

import (
	"context"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// OrderRepository 订单仓库
// 注意：请始终使用带 Context 的方法（如 CreateWithContext, ListWithContext 等）
// 这些方法会自动注入租户隔离条件，确保数据安全
type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单（内部使用，建议使用 CreateWithContext）
// Deprecated: 请使用 CreateWithContext 以确保租户隔离
func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

// CreateWithContext 创建订单（带租户上下文）
func (r *OrderRepository) CreateWithContext(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Scopes(WithTenantFromContext(ctx)).Create(order).Error
}

func (r *OrderRepository) FindByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByOrderNoWithContext 根据订单号查询（带租户上下文）
func (r *OrderRepository) FindByOrderNoWithContext(ctx context.Context, orderNo string) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Preload("Items").
		Where("order_no = ?", orderNo).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByTenantID 根据租户ID查询订单
func (r *OrderRepository) FindByTenantID(tenantID, orderID int64) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").
		Where("id = ? AND tenant_id = ?", orderID, tenantID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) List(page, size int, status, platform, keyword string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Preload("Items")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR buyer_nick LIKE ? OR receiver_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&orders).Error

	return orders, total, err
}

// ListWithContext 分页查询订单（带租户上下文）
func (r *OrderRepository) ListWithContext(ctx context.Context, page, size int, status, platform, keyword string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Scopes(WithTenantFromContext(ctx)).
		Preload("Items")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR buyer_nick LIKE ? OR receiver_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&orders).Error

	return orders, total, err
}

// ListByTenantID 根据租户ID分页查询订单
func (r *OrderRepository) ListByTenantID(tenantID int64, page, size int, status, platform, keyword string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Preload("Items").Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR buyer_nick LIKE ? OR receiver_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) UpdateStatus(orderNo string, status int) error {
	return r.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Update("status", status).Error
}

// UpdateStatusWithContext 更新订单状态（带租户上下文）
func (r *OrderRepository) UpdateStatusWithContext(ctx context.Context, orderNo string, status int) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Order{}).
		Where("order_no = ?", orderNo).
		Update("status", status).Error
}

func (r *OrderRepository) Ship(orderNo, company, no string) error {
	return r.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Updates(map[string]interface{}{
		"status":            models.OrderStatusShipped,
		"logistics_company": company,
		"logistics_no":      no,
	}).Error
}

// ShipWithContext 发货（带租户上下文）
func (r *OrderRepository) ShipWithContext(ctx context.Context, orderNo, company, no string) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Order{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"status":            models.OrderStatusShipped,
			"logistics_company": company,
			"logistics_no":      no,
		}).Error
}

// Upsert 插入或更新订单
func (r *OrderRepository) Upsert(order *models.Order) error {
	// 检查订单是否已存在
	var existing models.Order
	err := r.db.Where("platform_order_id = ? AND platform = ?",
		order.PlatformOrderID, order.Platform).First(&existing).Error

	if err == nil {
		// 已存在，更新
		order.ID = existing.ID
		return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(order).Error
	}

	// 不存在，创建
	return r.db.Create(order).Error
}

// UpsertWithContext 插入或更新订单（带租户上下文）
func (r *OrderRepository) UpsertWithContext(ctx context.Context, order *models.Order) error {
	// 检查订单是否已存在
	var existing models.Order
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("platform_order_id = ? AND platform = ? AND tenant_id = ?",
			order.PlatformOrderID, order.Platform, order.TenantID).
		First(&existing).Error

	if err == nil {
		// 已存在，更新
		order.ID = existing.ID
		return r.db.WithContext(ctx).
			Scopes(WithTenantFromContext(ctx)).
			Session(&gorm.Session{FullSaveAssociations: true}).
			Updates(order).Error
	}

	// 不存在，创建
	return r.db.WithContext(ctx).Create(order).Error
}

// CountByTenantID 统计租户订单数量
func (r *OrderRepository) CountByTenantID(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Order{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
