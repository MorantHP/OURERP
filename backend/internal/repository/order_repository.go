package repository

import (
	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) FindByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").Where("order_no = ?", orderNo).First(&order).Error
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

func (r *OrderRepository) UpdateStatus(orderNo string, status int) error {
	return r.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Update("status", status).Error
}

func (r *OrderRepository) Ship(orderNo, company, no string) error {
	return r.db.Model(&models.Order{}).Where("order_no = ?", orderNo).Updates(map[string]interface{}{
		"status":            models.OrderStatusShipped,
		"logistics_company": company,
		"logistics_no":      no,
	}).Error
}

// Upsert 插入或更新订单
func (r *OrderRepository) Upsert(order *models.Order) error {
	// 检查订单是否已存在
	var existing Order
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