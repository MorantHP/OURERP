package repository

import (
	"context"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// InventoryRepository 库存仓库
type InventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository 创建库存仓库
func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Create 创建库存记录
func (r *InventoryRepository) Create(inventory *models.Inventory) error {
	return r.db.Create(inventory).Error
}

// FindByID 根据ID查询库存
func (r *InventoryRepository) FindByID(id int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := r.db.First(&inventory, id).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// FindByIDWithContext 根据ID查询库存（带租户上下文）
func (r *InventoryRepository) FindByIDWithContext(ctx context.Context, id int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		First(&inventory, id).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// FindByProductAndWarehouse 根据商品ID和仓库ID查询库存
func (r *InventoryRepository) FindByProductAndWarehouse(ctx context.Context, productID, warehouseID int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&inventory).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

// FindOrCreateByProductAndWarehouse 查询或创建库存记录
func (r *InventoryRepository) FindOrCreateByProductAndWarehouse(ctx context.Context, tenantID, productID, warehouseID int64) (*models.Inventory, error) {
	var inventory models.Inventory
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		First(&inventory).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新的库存记录
		inventory = models.Inventory{
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    0,
			LockedQty:   0,
			TotalQty:    0,
		}
		if err := r.db.Create(&inventory).Error; err != nil {
			return nil, err
		}
		return &inventory, nil
	}

	if err != nil {
		return nil, err
	}

	return &inventory, nil
}

// List 分页查询库存列表
func (r *InventoryRepository) List(page, size int, warehouseID, productID int64, lowStock bool) ([]models.Inventory, int64, error) {
	var inventories []models.Inventory
	var total int64

	query := r.db.Model(&models.Inventory{})

	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if lowStock {
		query = query.Where("quantity <= alert_qty AND alert_qty > 0")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&inventories).Error
	return inventories, total, err
}

// ListWithContext 分页查询库存列表（带租户上下文）
func (r *InventoryRepository) ListWithContext(ctx context.Context, page, size int, warehouseID, productID int64, lowStock bool) ([]models.Inventory, int64, error) {
	var inventories []models.Inventory
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Scopes(WithTenantFromContext(ctx))

	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if lowStock {
		query = query.Where("quantity <= alert_qty AND alert_qty > 0")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&inventories).Error
	return inventories, total, err
}

// ListWithDetails 分页查询库存列表（带商品和仓库信息）
func (r *InventoryRepository) ListWithDetails(ctx context.Context, page, size int, warehouseID, productID int64, lowStock bool, keyword string) ([]models.InventoryWithProduct, int64, error) {
	var inventories []models.InventoryWithProduct
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Scopes(WithTenantFromContext(ctx)).
		Preload("Product").
		Preload("Warehouse")

	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if lowStock {
		query = query.Where("quantity <= alert_qty AND alert_qty > 0")
	}
	if keyword != "" {
		query = query.Joins("JOIN products ON products.id = inventories.product_id").
			Where("products.sku_code LIKE ? OR products.name LIKE ? OR products.barcode LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("inventories.id DESC").Offset(offset).Limit(size).Find(&inventories).Error
	return inventories, total, err
}

// Update 更新库存
func (r *InventoryRepository) Update(inventory *models.Inventory) error {
	return r.db.Save(inventory).Error
}

// UpdateWithContext 更新库存（带租户上下文）
func (r *InventoryRepository) UpdateWithContext(ctx context.Context, inventory *models.Inventory) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Save(inventory).Error
}

// UpdateQuantity 更新库存数量
func (r *InventoryRepository) UpdateQuantity(ctx context.Context, productID, warehouseID int64, changeQty int) error {
	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Updates(map[string]interface{}{
			"quantity":  gorm.Expr("quantity + ?", changeQty),
			"total_qty": gorm.Expr("total_qty + ?", changeQty),
		}).Error
}

// AdjustQuantity 调整库存数量（带事务）
func (r *InventoryRepository) AdjustQuantity(ctx context.Context, tenantID, productID, warehouseID int64, changeQty int, beforeUpdate func(*models.Inventory) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 获取或创建库存记录
		var inventory models.Inventory
		err := tx.Where("product_id = ? AND warehouse_id = ? AND tenant_id = ?", productID, warehouseID, tenantID).
			First(&inventory).Error

		if err == gorm.ErrRecordNotFound {
			inventory = models.Inventory{
				TenantID:    tenantID,
				ProductID:   productID,
				WarehouseID: warehouseID,
				Quantity:    changeQty,
				LockedQty:   0,
				TotalQty:    changeQty,
			}
			if err := tx.Create(&inventory).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// 更新库存
			inventory.Quantity += changeQty
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

// CreateLog 创建库存流水记录
func (r *InventoryRepository) CreateLog(log *models.InventoryLog) error {
	return r.db.Create(log).Error
}

// ListLogs 查询库存流水
func (r *InventoryRepository) ListLogs(ctx context.Context, page, size int, productID, warehouseID int64, refType string) ([]models.InventoryLog, int64, error) {
	var logs []models.InventoryLog
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.InventoryLog{}).
		Scopes(WithTenantFromContext(ctx))

	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if warehouseID > 0 {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if refType != "" {
		query = query.Where("ref_type = ?", refType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&logs).Error
	return logs, total, err
}

// GetLowStockAlert 获取库存预警列表
func (r *InventoryRepository) GetLowStockAlert(ctx context.Context) ([]models.Inventory, error) {
	var inventories []models.Inventory
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("quantity <= alert_qty AND alert_qty > 0").
		Order("quantity ASC").
		Find(&inventories).Error
	return inventories, err
}

// GetTotalQuantity 获取商品总库存
func (r *InventoryRepository) GetTotalQuantity(ctx context.Context, productID int64) (int, error) {
	var total int
	err := r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Scopes(WithTenantFromContext(ctx)).
		Where("product_id = ?", productID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	return total, err
}
