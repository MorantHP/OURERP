package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSuite 测试套件
type TestSuite struct {
	DB            *gorm.DB
	UserRepo      *repository.UserRepository
	TenantRepo    *repository.TenantRepository
	ShopRepo      *repository.ShopRepository
	ProductRepo   *repository.ProductRepository
	WarehouseRepo *repository.WarehouseRepository
	InventoryRepo *repository.InventoryRepository
	OrderRepo     *repository.OrderRepository
	TenantID      int64
	TestUser      *models.User
}

// NewTestSuite 创建测试套件
func NewTestSuite(t *testing.T) *TestSuite {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.TenantUser{},
		&models.Shop{},
		&models.Product{},
		&models.Warehouse{},
		&models.Inventory{},
		&models.InventoryLog{},
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
		&models.ProductCost{},
	)

	suite := &TestSuite{
		DB:            db,
		UserRepo:      repository.NewUserRepository(db),
		TenantRepo:    repository.NewTenantRepository(db),
		ShopRepo:      repository.NewShopRepository(db),
		ProductRepo:   repository.NewProductRepository(db),
		WarehouseRepo: repository.NewWarehouseRepository(db),
		InventoryRepo: repository.NewInventoryRepository(db),
		OrderRepo:     repository.NewOrderRepository(db),
	}

	// 创建测试数据
	suite.SeedTestData(t)

	return suite
}

// SeedTestData 填充测试数据
func (s *TestSuite) SeedTestData(t *testing.T) {
	// 创建测试用户
	user := &models.User{
		Email:      "test@example.com",
		Name:       "测试用户",
		Phone:      "13800138000",
		IsApproved: true,
		Status:     1,
	}
	user.SetPassword("password123")
	if err := s.UserRepo.Create(user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	s.TestUser = user

	// 创建测试租户
	tenant := &models.Tenant{
		Code:        "TEST001",
		Name:        "测试租户",
		Platform:    "taobao",
		Description: "自动化测试租户",
		Status:      1,
		OwnerID:     user.ID,
	}
	if err := s.TenantRepo.Create(tenant); err != nil {
		t.Fatalf("failed to create test tenant: %v", err)
	}
	s.TenantID = tenant.ID

	// 创建租户用户关联
	tenantUser := &models.TenantUser{
		TenantID: tenant.ID,
		UserID:   user.ID,
		Role:     "owner",
	}
	s.DB.Create(tenantUser)
}

// CreateTestProduct 创建测试商品
func (s *TestSuite) CreateTestProduct(t *testing.T) *models.Product {
	product := &models.Product{
		TenantID:  s.TenantID,
		SkuCode:   fmt.Sprintf("SKU_%d", time.Now().UnixNano()),
		Name:      "测试商品",
		Category:  "测试类目",
		Brand:     "测试品牌",
		CostPrice: 50.00,
		SalePrice: 100.00,
		Status:    1,
	}
	if err := s.ProductRepo.Create(product); err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}
	return product
}

// CreateTestWarehouse 创建测试仓库
func (s *TestSuite) CreateTestWarehouse(t *testing.T) *models.Warehouse {
	warehouse := &models.Warehouse{
		TenantID:  s.TenantID,
		Code:      fmt.Sprintf("WH_%d", time.Now().UnixNano()),
		Name:      "测试仓库",
		Address:   "测试地址",
		IsDefault: true,
		Status:    1,
	}
	if err := s.WarehouseRepo.Create(warehouse); err != nil {
		t.Fatalf("failed to create test warehouse: %v", err)
	}
	return warehouse
}

// CreateTestOrder 创建测试订单
func (s *TestSuite) CreateTestOrder(t *testing.T) *models.Order {
	order := &models.Order{
		TenantID:        s.TenantID,
		OrderNo:         fmt.Sprintf("ORD_%d", time.Now().UnixNano()),
		Platform:        "taobao",
		PlatformOrderID: fmt.Sprintf("PO_%d", time.Now().UnixNano()),
		Status:          100,
		TotalAmount:     199.00,
		PayAmount:       199.00,
		BuyerNick:       "测试买家",
		ReceiverName:    "张三",
		ReceiverPhone:   "13800138000",
		ReceiverAddress: "测试地址",
	}
	if err := s.OrderRepo.Create(order); err != nil {
		t.Fatalf("failed to create test order: %v", err)
	}
	return order
}

// Cleanup 清理测试数据
func (s *TestSuite) Cleanup() {
	// SQLite 内存数据库会自动清理
}

// MockEnv 设置模拟环境变量
func MockEnv(key, value string) func() {
	original := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if original == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, original)
		}
	}
}

// AssertError asserts that an error occurred
func AssertError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", expectedMsg)
	}
}

// AssertNoError asserts that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertEqual asserts two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// AssertNotEmpty asserts a string is not empty
func AssertNotEmpty(t *testing.T, s string, field string) {
	t.Helper()
	if s == "" {
		t.Fatalf("%s should not be empty", field)
	}
}
