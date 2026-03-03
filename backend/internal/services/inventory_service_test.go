package services

import (
	"context"
	"testing"

	"github.com/MorantHP/OURERP/backend/internal/cache"
	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/errors"
	"github.com/MorantHP/OURERP/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestInventoryService_DeductInventory 测试扣减库存
func TestInventoryService_DeductInventory(t *testing.T) {
	// Setup
	db := setupTestDBWithInventory(t)
	inventoryRepo := repository.NewInventoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	mockCache := cache.NewMemoryCache()

	service := NewInventoryService(inventoryRepo, productRepo, warehouseRepo, mockCache)

	// 创建测试数据
	product := &models.Product{SkuCode: "SKU001", Name: "Test", TenantID: 1}
	db.Create(product)

	warehouse := &models.Warehouse{Name: "WH001", TenantID: 1}
	db.Create(warehouse)

	inventory := &models.Inventory{
		TenantID:    1,
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    100,
		TotalQty:    100,
	}
	db.Create(inventory)

	// 设置租户上下文
	ctx := context.Background()
	ctx = repository.SetTenantIDToContext(ctx, 1)

	// Test - 扣减库存
	err := service.DeductInventory(ctx, product.ID, warehouse.ID, 10, "ORD001")

	// Assert
	assert.NoError(t, err)

	// 验证库存
	var updated models.Inventory
	db.First(&updated, inventory.ID)
	assert.Equal(t, 90, updated.Quantity)
}

// TestInventoryService_DeductInventory_NotEnough 测试库存不足
func TestInventoryService_DeductInventory_NotEnough(t *testing.T) {
	db := setupTestDBWithInventory(t)
	inventoryRepo := repository.NewInventoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	mockCache := cache.NewMemoryCache()

	service := NewInventoryService(inventoryRepo, productRepo, warehouseRepo, mockCache)

	// 创建测试数据
	product := &models.Product{SkuCode: "SKU001", Name: "Test", TenantID: 1}
	db.Create(product)

	warehouse := &models.Warehouse{Name: "WH001", TenantID: 1}
	db.Create(warehouse)

	inventory := &models.Inventory{
		TenantID:    1,
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    10,
		TotalQty:    10,
	}
	db.Create(inventory)

	ctx := context.Background()
	ctx = repository.SetTenantIDToContext(ctx, 1)

	// Test - 尝试扣减超过库存的数量
	err := service.DeductInventory(ctx, product.ID, warehouse.ID, 20, "ORD001")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInventoryNotEnough, err)

	// 验证库存未变
	var updated models.Inventory
	db.First(&updated, inventory.ID)
	assert.Equal(t, 10, updated.Quantity)
}

// TestInventoryService_TransferStock 测试调拨库存
func TestInventoryService_TransferStock(t *testing.T) {
	db := setupTestDBWithInventory(t)
	inventoryRepo := repository.NewInventoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	mockCache := cache.NewMemoryCache()

	service := NewInventoryService(inventoryRepo, productRepo, warehouseRepo, mockCache)

	// 创建测试数据
	product := &models.Product{SkuCode: "SKU001", Name: "Test", TenantID: 1}
	db.Create(product)

	warehouse1 := &models.Warehouse{Name: "WH001", TenantID: 1}
	warehouse2 := &models.Warehouse{Name: "WH002", TenantID: 1}
	db.Create(warehouse1)
	db.Create(warehouse2)

	inventory1 := &models.Inventory{
		TenantID:    1,
		ProductID:   product.ID,
		WarehouseID: warehouse1.ID,
		Quantity:    100,
		TotalQty:    100,
	}
	db.Create(inventory1)

	ctx := context.Background()
	ctx = repository.SetTenantIDToContext(ctx, 1)

	// Test - 调拨
	err := service.TransferStock(ctx, product.ID, warehouse1.ID, warehouse2.ID, 30, "TF001")

	// Assert
	assert.NoError(t, err)

	// 验证源仓库库存
	var updated1 models.Inventory
	db.First(&updated1, inventory1.ID)
	assert.Equal(t, 70, updated1.Quantity)

	// 验证目标仓库库存
	var updated2 models.Inventory
	db.Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse2.ID).First(&updated2)
	assert.Equal(t, 30, updated2.Quantity)
}

// TestInventoryService_ReturnInventory 测试退还库存
func TestInventoryService_ReturnInventory(t *testing.T) {
	db := setupTestDBWithInventory(t)
	inventoryRepo := repository.NewInventoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	mockCache := cache.NewMemoryCache()

	service := NewInventoryService(inventoryRepo, productRepo, warehouseRepo, mockCache)

	// 创建测试数据
	product := &models.Product{SkuCode: "SKU001", Name: "Test", TenantID: 1}
	db.Create(product)

	warehouse := &models.Warehouse{Name: "WH001", TenantID: 1}
	db.Create(warehouse)

	inventory := &models.Inventory{
		TenantID:    1,
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    50,
		TotalQty:    50,
	}
	db.Create(inventory)

	ctx := context.Background()
	ctx = repository.SetTenantIDToContext(ctx, 1)

	// Test - 退还库存
	err := service.ReturnInventory(ctx, product.ID, warehouse.ID, 20, "ORD001")

	// Assert
	assert.NoError(t, err)

	// 验证库存
	var updated models.Inventory
	db.First(&updated, inventory.ID)
	assert.Equal(t, 70, updated.Quantity)
}

// setupTestDBWithInventory 设置测试数据库（带库存表）
func setupTestDBWithInventory(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
		&models.Product{},
		&models.Warehouse{},
		&models.Inventory{},
		&models.InventoryLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
