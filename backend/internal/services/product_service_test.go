package services

import (
	"context"
	"testing"

	"github.com/MorantHP/OURERP/backend/internal/cache"
	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/errors"
	"github.com/MorantHP/OURERP/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockProductRepository 商品仓库Mock
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) CreateWithContext(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) FindByIDWithContext(ctx context.Context, id int64) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) FindBySkuCode(ctx context.Context, skuCode string) (*models.Product, error) {
	args := m.Called(ctx, skuCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) ListWithContext(ctx context.Context, page, size int, category, brand, keyword string, status *int) ([]models.Product, int64, error) {
	args := m.Called(ctx, page, size, category, brand, keyword, status)
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) UpdateWithContext(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteWithContext(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) GetCategories(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProductRepository) GetBrands(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// TestProductService_CreateProduct 测试创建商品
func TestProductService_CreateProduct(t *testing.T) {
	// Setup
	mockRepo := new(MockProductRepository)
	mockInventoryRepo := &repository.InventoryRepository{}
	mockCache := cache.NewMemoryCache()
	
	service := NewProductService(mockRepo, mockInventoryRepo, mockCache)

	ctx := context.Background()
	req := &models.CreateProductRequest{
		SkuCode: "SKU001",
		Name:    "Test Product",
	}

	// Mock
	mockRepo.On("FindBySkuCode", ctx, "SKU001").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("CreateWithContext", ctx, mock.AnythingOfType("*models.Product")).Return(nil)

	// Test
	product, err := service.CreateProduct(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "SKU001", product.SkuCode)
	assert.Equal(t, "Test Product", product.Name)
	mockRepo.AssertExpectations(t)
}

// TestProductService_CreateProduct_Duplicate 测试创建重复商品
func TestProductService_CreateProduct_Duplicate(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockInventoryRepo := &repository.InventoryRepository{}
	mockCache := cache.NewMemoryCache()
	
	service := NewProductService(mockRepo, mockInventoryRepo, mockCache)

	ctx := context.Background()
	req := &models.CreateProductRequest{
		SkuCode: "SKU001",
		Name:    "Test Product",
	}

	// Mock - SKU已存在
	existingProduct := &models.Product{ID: 1, SkuCode: "SKU001"}
	mockRepo.On("FindBySkuCode", ctx, "SKU001").Return(existingProduct, nil)

	// Test
	product, err := service.CreateProduct(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, errors.ErrProductDuplicate, err)
	mockRepo.AssertExpectations(t)
}

// TestProductService_GetProduct 测试获取商品
func TestProductService_GetProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockInventoryRepo := &repository.InventoryRepository{}
	mockCache := cache.NewMemoryCache()
	
	service := NewProductService(mockRepo, mockInventoryRepo, mockCache)

	ctx := context.Background()
	expectedProduct := &models.Product{
		ID:      1,
		SkuCode: "SKU001",
		Name:    "Test Product",
	}

	// Mock
	mockRepo.On("FindByIDWithContext", ctx, int64(1)).Return(expectedProduct, nil)

	// Test
	result, err := service.GetProduct(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	mockRepo.AssertExpectations(t)
}

// TestProductService_DeleteProduct 测试删除商品
func TestProductService_DeleteProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockInventoryRepo := &repository.InventoryRepository{}
	mockCache := cache.NewMemoryCache()
	
	service := NewProductService(mockRepo, mockInventoryRepo, mockCache)

	ctx := context.Background()

	// Mock
	mockRepo.On("DeleteWithContext", ctx, int64(1)).Return(nil)

	// Test
	err := service.DeleteProduct(ctx, 1)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.Product{}, &models.Inventory{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
