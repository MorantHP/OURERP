package repository

import (
	"context"

	"github.com/MorantHP/OURERP/backend/internal/models"
	"gorm.io/gorm"
)

// AdjustQuantityWithLock 带锁的库存调整（防超卖）
func (r *InventoryRepository) AdjustQuantityWithLock(ctx context.Context, tenantID, productID, warehouseID int64, changeQty int, beforeUpdate func(*models.Inventory) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用 FOR UPDATE 悲观锁
		var inventory models.Inventory
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", productID, warehouseID, tenantID).
			First(&inventory).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			inventory = models.Inventory{
				TenantID:    tenantID,
				ProductID:   productID,
				WarehouseID: warehouseID,
				Quantity:    changeQty,
				LockedQty:   0,
				TotalQty:    changeQty,
			}
			if changeQty < 0 {
				// 不允许新记录为负数
				return gorm.ErrInvalidData
			}
			if err := tx.Create(&inventory).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// 检查库存是否充足
			newQty := inventory.Quantity + changeQty
			if newQty < 0 {
				return gorm.ErrInvalidData // 库存不足
			}

			// 更新库存
			inventory.Quantity = newQty
			inventory.TotalQty += changeQty
			if err := tx.Save(&inventory).Error; err != nil {
				return err
			}
		}

		// 调用前置函数
		if beforeUpdate != nil {
			return beforeUpdate(&inventory)
		}
		return nil
	})
}

// Transaction 事务包装器
func (r *InventoryRepository) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// FindByProductAndWarehouseWithTx 带事务的查询
func (r *InventoryRepository) FindByProductAndWarehouseWithTx(ctx context.Context, tx *gorm.DB, productID, warehouseID int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := tx.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&inventory).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// FindOrCreateByProductAndWarehouseWithTx 带事务的查询或创建
func (r *InventoryRepository) FindOrCreateByProductAndWarehouseWithTx(ctx context.Context, tx *gorm.DB, tenantID, productID, warehouseID int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := tx.WithContext(ctx).
		Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", productID, warehouseID, tenantID).
		First(&inventory).Error

	if err == gorm.ErrRecordNotFound {
		inventory = models.Inventory{
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    0,
			LockedQty:   0,
			TotalQty:    0,
		}
		if err := tx.Create(&inventory).Error; err != nil {
			return nil, err
		}
		return &inventory, nil
	}

	if err != nil {
		return nil, err
	}

	return &inventory, nil
}

// GetAlertWarehouseCount 获取预警仓库数量
func (r *InventoryRepository) GetAlertWarehouseCount(ctx context.Context, productID int64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Scopes(WithTenantFromContext(ctx)).
		Where("product_id = ? AND quantity <= alert_qty AND alert_qty > 0", productID).
		Count(&count).Error
	return int(count), err
}

// BatchUpdateQuantity 批量更新库存（带防超卖）
func (r *InventoryRepository) BatchUpdateQuantity(ctx context.Context, updates []InventoryUpdate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			// 使用 FOR UPDATE 锁定
			var inventory models.Inventory
			err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", 
					update.ProductID, update.WarehouseID, update.TenantID).
				First(&inventory).Error

			if err == gorm.ErrRecordNotFound {
				if update.ChangeQty < 0 {
					return gorm.ErrInvalidData // 库存不足
				}
				inventory = models.Inventory{
					TenantID:    update.TenantID,
					ProductID:   update.ProductID,
					WarehouseID: update.WarehouseID,
					Quantity:    update.ChangeQty,
					TotalQty:    update.ChangeQty,
				}
				if err := tx.Create(&inventory).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				newQty := inventory.Quantity + update.ChangeQty
				if newQty < 0 {
					return gorm.ErrInvalidData // 库存不足
				}
				inventory.Quantity = newQty
				inventory.TotalQty += update.ChangeQty
				if err := tx.Save(&inventory).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// InventoryUpdate 库存更新
type InventoryUpdate struct {
	TenantID    int64
	ProductID   int64
	WarehouseID int64
	ChangeQty   int
}
