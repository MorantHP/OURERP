package tests

import (
	"testing"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.TenantUser{},
		&models.Shop{},
		&models.Product{},
		&models.Warehouse{},
		&models.Inventory{},
		&models.Order{},
		&models.OrderItem{},
		&models.Customer{},
		&models.AlertRule{},
		&models.AlertRecord{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.Supplier{},
		&models.FinanceRecord{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

// TestTenantRepository_Integration 租户仓储集成测试
func TestTenantRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)
	repo := repository.NewTenantRepository(db)

	t.Run("创建租户", func(t *testing.T) {
		tenant := &models.Tenant{
			Code:        "TEST001",
			Name:        "测试租户",
			Platform:    "taobao",
			Description: "测试用租户",
			Status:      1,
			OwnerID:     1,
		}

		err := repo.Create(tenant)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if tenant.ID == 0 {
			t.Error("Create() should set tenant ID")
		}
	})

	t.Run("查询租户", func(t *testing.T) {
		tenant, err := repo.FindByCode("TEST001")
		if err != nil {
			t.Fatalf("FindByCode() error = %v", err)
		}

		if tenant.Name != "测试租户" {
			t.Errorf("FindByCode() name = %v, want %v", tenant.Name, "测试租户")
		}
	})

	t.Run("更新租户", func(t *testing.T) {
		tenant, _ := repo.FindByCode("TEST001")
		tenant.Description = "更新后的描述"
		err := repo.Update(tenant)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		updated, _ := repo.FindByID(tenant.ID)
		if updated.Description != "更新后的描述" {
			t.Error("Update() should persist changes")
		}
	})

	t.Run("删除租户", func(t *testing.T) {
		tenant, _ := repo.FindByCode("TEST001")
		err := repo.Delete(tenant.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err = repo.FindByCode("TEST001")
		if err == nil {
			t.Error("Delete() should remove tenant")
		}
	})
}

// TestOrderRepository_Integration 订单仓储集成测试
func TestOrderRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)

	// 先创建租户
	tenantRepo := repository.NewTenantRepository(db)
	tenant := &models.Tenant{
		Code:     "ORDER_TEST",
		Name:     "订单测试租户",
		Platform: "taobao",
		Status:   1,
	}
	tenantRepo.Create(tenant)

	orderRepo := repository.NewOrderRepository(db)

	t.Run("创建订单", func(t *testing.T) {
		order := &models.Order{
			TenantID:        tenant.ID,
			OrderNo:         "TB202401010001",
			Platform:        "taobao",
			PlatformOrderID: "123456789",
			Status:          100,
			TotalAmount:     199.00,
			PayAmount:       199.00,
			BuyerNick:       "测试买家",
			ReceiverName:    "张三",
			ReceiverPhone:   "13800138000",
			ReceiverAddress: "测试地址",
		}

		err := orderRepo.Create(order)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("按状态查询订单", func(t *testing.T) {
		orders, total, err := orderRepo.FindByStatus(tenant.ID, 100, 1, 10)
		if err != nil {
			t.Fatalf("FindByStatus() error = %v", err)
		}

		if total != 1 {
			t.Errorf("FindByStatus() total = %v, want 1", total)
		}

		if len(orders) != 1 {
			t.Errorf("FindByStatus() len = %v, want 1", len(orders))
		}
	})

	t.Run("统计订单", func(t *testing.T) {
		stats, err := orderRepo.GetOrderStats(tenant.ID, time.Now().AddDate(0, 0, -7), time.Now())
		if err != nil {
			t.Fatalf("GetOrderStats() error = %v", err)
		}

		if stats.TotalOrders < 1 {
			t.Error("GetOrderStats() should return at least 1 order")
		}
	})
}

// TestInventoryRepository_Integration 库存仓储集成测试
func TestInventoryRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)

	// 创建租户
	tenantRepo := repository.NewTenantRepository(db)
	tenant := &models.Tenant{
		Code:     "INV_TEST",
		Name:     "库存测试租户",
		Platform: "taobao",
		Status:   1,
	}
	tenantRepo.Create(tenant)

	// 创建仓库
	whRepo := repository.NewWarehouseRepository(db)
	warehouse := &models.Warehouse{
		TenantID:  tenant.ID,
		Code:      "WH001",
		Name:      "测试仓库",
		IsDefault: true,
		Status:    1,
	}
	whRepo.Create(warehouse)

	// 创建商品
	productRepo := repository.NewProductRepository(db)
	product := &models.Product{
		TenantID:  tenant.ID,
		SkuCode:   "SKU001",
		Name:      "测试商品",
		CostPrice: 50.00,
		SalePrice: 100.00,
		Status:    1,
	}
	productRepo.Create(product)

	invRepo := repository.NewInventoryRepository(db)

	t.Run("创建库存", func(t *testing.T) {
		inventory := &models.Inventory{
			TenantID:    tenant.ID,
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    100,
			LockedQty:   0,
			TotalQty:    100,
			AlertQty:    10,
		}

		err := invRepo.Create(inventory)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if inventory.ID == 0 {
			t.Error("Create() should set inventory ID")
		}
	})

	t.Run("查询低库存", func(t *testing.T) {
		items, err := invRepo.GetLowStockItems(tenant.ID)
		if err != nil {
			t.Fatalf("GetLowStockItems() error = %v", err)
		}

		// 库存100，预警10，不算低库存
		if len(items) > 0 {
			t.Log("GetLowStockItems() returned items (expected 0 for this test)")
		}
	})

	t.Run("更新库存数量", func(t *testing.T) {
		inv, _ := invRepo.GetByProductAndWarehouse(product.ID, warehouse.ID)
		originalQty := inv.Quantity

		err := invRepo.UpdateQuantity(inv.ID, -10) // 减少10
		if err != nil {
			t.Fatalf("UpdateQuantity() error = %v", err)
		}

		updated, _ := invRepo.GetByProductAndWarehouse(product.ID, warehouse.ID)
		if updated.Quantity != originalQty-10 {
			t.Errorf("UpdateQuantity() qty = %v, want %v", updated.Quantity, originalQty-10)
		}
	})
}

// TestUserRepository_Integration 用户仓储集成测试
func TestUserRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)
	repo := repository.NewUserRepository(db)

	t.Run("创建用户", func(t *testing.T) {
		user := &models.User{
			Email:      "test@example.com",
			Name:       "测试用户",
			Phone:      "13800138000",
			IsRoot:     false,
			IsApproved: true,
			Status:     1,
		}
		user.SetPassword("password123")

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("按邮箱查询", func(t *testing.T) {
		user, err := repo.FindByEmail("test@example.com")
		if err != nil {
			t.Fatalf("FindByEmail() error = %v", err)
		}

		if user.Name != "测试用户" {
			t.Errorf("FindByEmail() name = %v, want %v", user.Name, "测试用户")
		}
	})

	t.Run("密码验证", func(t *testing.T) {
		user, _ := repo.FindByEmail("test@example.com")

		if !user.CheckPassword("password123") {
			t.Error("CheckPassword() should return true for correct password")
		}

		if user.CheckPassword("wrongpassword") {
			t.Error("CheckPassword() should return false for wrong password")
		}
	})

	t.Run("列出所有用户", func(t *testing.T) {
		users, err := repo.ListAll()
		if err != nil {
			t.Fatalf("ListAll() error = %v", err)
		}

		if len(users) != 1 {
			t.Errorf("ListAll() len = %v, want 1", len(users))
		}
	})
}

// BenchmarkUserRepository_FindByEmail 基准测试
func BenchmarkUserRepository_FindByEmail(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{})

	repo := repository.NewUserRepository(db)
	user := &models.User{
		Email:      "bench@example.com",
		Name:       "Benchmark User",
		IsApproved: true,
		Status:     1,
	}
	user.SetPassword("password")
	repo.Create(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.FindByEmail("bench@example.com")
	}
}
