package services

import (
	"context"
	"fmt"

	"github.com/MorantHP/OURERP/backend/internal/cache"
	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/errors"
	"github.com/MorantHP/OURERP/backend/internal/repository"
	"gorm.io/gorm"
)

// InventoryService 库存服务
type InventoryService struct {
	inventoryRepo  *repository.InventoryRepository
	productRepo    *repository.ProductRepository
	warehouseRepo  *repository.WarehouseRepository
	cacheDecorator *CacheDecorator
}

// NewInventoryService 创建库存服务
func NewInventoryService(
	inventoryRepo *repository.InventoryRepository,
	productRepo *repository.ProductRepository,
	warehouseRepo *repository.WarehouseRepository,
	cacheService cache.CacheService,
) *InventoryService {
	return &InventoryService{
		inventoryRepo:  inventoryRepo,
		productRepo:    productRepo,
		warehouseRepo:  warehouseRepo,
		cacheDecorator: NewCacheDecorator(cacheService, "inventory"),
	}
}

// AdjustInventory 调整库存（防超卖）
func (s *InventoryService) AdjustInventory(ctx context.Context, productID, warehouseID int64, changeQty int, refType, refID string, remark string) error {
	tenantID := repository.GetTenantIDFromContext(ctx)

	// 使用事务 + 悲观锁
	err := s.inventoryRepo.AdjustQuantityWithLock(ctx, tenantID, productID, warehouseID, changeQty, func(inventory *models.Inventory) error {
		// 检查库存是否充足（扣减时）
		if changeQty < 0 && inventory.Quantity < -changeQty {
			return errors.ErrInventoryNotEnough
		}

		// 创建库存流水
		log := &models.InventoryLog{
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: warehouseID,
			ChangeQty:   changeQty,
			BeforeQty:   inventory.Quantity,
			AfterQty:    inventory.Quantity + changeQty,
			RefType:     refType,
			RefID:       refID,
			Remark:      remark,
		}
		return s.inventoryRepo.CreateLog(log)
	})

	if err != nil {
		return err
	}

	// 使缓存失效
	_ = s.cacheDecorator.InvalidateInventoryCache(ctx, productID, warehouseID, tenantID)

	return nil
}

// DeductInventory 扣减库存（带预扣）
func (s *InventoryService) DeductInventory(ctx context.Context, productID, warehouseID int64, qty int, orderNo string) error {
	if qty <= 0 {
		return errors.NewAppError("INVALID_QUANTITY", "扣减数量必须大于0", 400, nil)
	}

	return s.AdjustInventory(ctx, productID, warehouseID, -qty, "order", orderNo, fmt.Sprintf("订单扣减: %s", orderNo))
}

// ReturnInventory 退还库存
func (s *InventoryService) ReturnInventory(ctx context.Context, productID, warehouseID int64, qty int, orderNo string) error {
	if qty <= 0 {
		return errors.NewAppError("INVALID_QUANTITY", "退还数量必须大于0", 400, nil)
	}

	return s.AdjustInventory(ctx, productID, warehouseID, qty, "return", orderNo, fmt.Sprintf("订单退还: %s", orderNo))
}

// GetInventory 获取库存详情
func (s *InventoryService) GetInventory(ctx context.Context, productID, warehouseID int64) (*models.Inventory, error) {
	return s.inventoryRepo.FindByProductAndWarehouse(ctx, productID, warehouseID)
}

// ListInventory 查询库存列表
func (s *InventoryService) ListInventory(ctx context.Context, page, size int, warehouseID, productID int64, lowStock bool, keyword string) ([]models.InventoryWithProduct, int64, error) {
	return s.inventoryRepo.ListWithDetails(ctx, page, size, warehouseID, productID, lowStock, keyword)
}

// GetLowStockAlert 获取库存预警列表
func (s *InventoryService) GetLowStockAlert(ctx context.Context) ([]models.Inventory, error) {
	return s.inventoryRepo.GetLowStockAlert(ctx)
}

// TransferStock 调拨库存
func (s *InventoryService) TransferStock(ctx context.Context, productID, fromWarehouseID, toWarehouseID int64, qty int, refNo string) error {
	if qty <= 0 {
		return errors.NewAppError("INVALID_QUANTITY", "调拨数量必须大于0", 400, nil)
	}

	tenantID := repository.GetTenantIDFromContext(ctx)

	// 使用事务
	return s.inventoryRepo.Transaction(ctx, func(tx *gorm.DB) error {
		// 从源仓库扣减
		fromInv, err := s.inventoryRepo.FindByProductAndWarehouseWithTx(ctx, tx, productID, fromWarehouseID)
		if err != nil {
			return errors.ErrNotFound
		}
		if fromInv.Quantity < qty {
			return errors.ErrInventoryNotEnough
		}
		fromInv.Quantity -= qty
		fromInv.TotalQty -= qty
		if err := tx.Save(fromInv).Error; err != nil {
			return err
		}

		// 创建出库流水
		outLog := &models.InventoryLog{
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: fromWarehouseID,
			ChangeQty:   -qty,
			BeforeQty:   fromInv.Quantity + qty,
			AfterQty:    fromInv.Quantity,
			RefType:     "transfer_out",
			RefID:       refNo,
			Remark:      fmt.Sprintf("调拨至仓库%d: %s", toWarehouseID, refNo),
		}
		if err := tx.Create(outLog).Error; err != nil {
			return err
		}

		// 到目标仓库增加
		toInv, err := s.inventoryRepo.FindOrCreateByProductAndWarehouseWithTx(ctx, tx, tenantID, productID, toWarehouseID)
		if err != nil {
			return err
		}
		toInv.Quantity += qty
		toInv.TotalQty += qty
		if err := tx.Save(toInv).Error; err != nil {
			return err
		}

		// 创建入库流水
		inLog := &models.InventoryLog{
			TenantID:    tenantID,
			ProductID:   productID,
			WarehouseID: toWarehouseID,
			ChangeQty:   qty,
			BeforeQty:   toInv.Quantity - qty,
			AfterQty:    toInv.Quantity,
			RefType:     "transfer_in",
			RefID:       refNo,
			Remark:      fmt.Sprintf("从仓库%d调拨: %s", fromWarehouseID, refNo),
		}
		return tx.Create(inLog).Error
	})
}
