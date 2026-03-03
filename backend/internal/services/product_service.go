package services

import (
	"context"

	"github.com/MorantHP/OURERP/backend/internal/cache"
	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/errors"
	"github.com/MorantHP/OURERP/backend/internal/repository"
)

// ProductService 商品服务
type ProductService struct {
	productRepo   *repository.ProductRepository
	inventoryRepo *repository.InventoryRepository
	cacheDecorator *CacheDecorator
}

// NewProductService 创建商品服务
func NewProductService(
	productRepo *repository.ProductRepository,
	inventoryRepo *repository.InventoryRepository,
	cacheService cache.CacheService,
) *ProductService {
	return &ProductService{
		productRepo:    productRepo,
		inventoryRepo:  inventoryRepo,
		cacheDecorator: NewCacheDecorator(cacheService, "product"),
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	// 检查SKU编码是否已存在
	if _, err := s.productRepo.FindBySkuCode(ctx, req.SkuCode); err == nil {
		return nil, errors.ErrProductDuplicate
	}

	product := &models.Product{
		SkuCode:   req.SkuCode,
		Name:      req.Name,
		Category:  req.Category,
		Brand:     req.Brand,
		Barcode:   req.Barcode,
		ImageURL:  req.ImageURL,
		Unit:      req.Unit,
		CostPrice: req.CostPrice,
		SalePrice: req.SalePrice,
		Specs:     req.Specs,
		Status:    models.ProductStatusOnline,
		Remark:    req.Remark,
	}

	if err := s.productRepo.CreateWithContext(ctx, product); err != nil {
		return nil, errors.WrapInternal(err, "创建商品失败")
	}

	// 使缓存失效
	tenantID := repository.GetTenantIDFromContext(ctx)
	_ = s.cacheDecorator.InvalidateProductCache(ctx, product.ID, tenantID)

	return product, nil
}

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(ctx context.Context, id int64) (*models.ProductWithInventory, error) {
	var result models.ProductWithInventory

	err := s.cacheDecorator.GetOrSet(ctx, cache.BuildKey(cache.CacheKeyProduct, id), &result, cache.TTLShort, func() (interface{}, error) {
		product, err := s.productRepo.FindByIDWithContext(ctx, id)
		if err != nil {
			return nil, errors.ErrProductNotFound
		}

		totalQty, _ := s.inventoryRepo.GetTotalQuantity(ctx, id)
		alertCount, _ := s.inventoryRepo.GetAlertWarehouseCount(ctx, id)

		return &models.ProductWithInventory{
			Product:      *product,
			TotalQuantity: totalQty,
			AlertCount:   alertCount,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListProducts 查询商品列表
func (s *ProductService) ListProducts(ctx context.Context, page, size int, category, brand, keyword string, status *int) ([]models.Product, int64, error) {
	return s.productRepo.ListWithContext(ctx, page, size, category, brand, keyword, status)
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, id int64, req *models.UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		return nil, errors.ErrProductNotFound
	}

	// 更新字段
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Brand != "" {
		product.Brand = req.Brand
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Unit != "" {
		product.Unit = req.Unit
	}
	if req.CostPrice > 0 {
		product.CostPrice = req.CostPrice
	}
	if req.SalePrice > 0 {
		product.SalePrice = req.SalePrice
	}
	if req.Specs != nil {
		product.Specs = req.Specs
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.Remark != "" {
		product.Remark = req.Remark
	}

	if err := s.productRepo.UpdateWithContext(ctx, product); err != nil {
		return nil, errors.WrapInternal(err, "更新商品失败")
	}

	// 使缓存失效
	tenantID := repository.GetTenantIDFromContext(ctx)
	_ = s.cacheDecorator.InvalidateProductCache(ctx, id, tenantID)

	return product, nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	if err := s.productRepo.DeleteWithContext(ctx, id); err != nil {
		return errors.WrapInternal(err, "删除商品失败")
	}

	// 使缓存失效
	tenantID := repository.GetTenantIDFromContext(ctx)
	_ = s.cacheDecorator.InvalidateProductCache(ctx, id, tenantID)

	return nil
}

// GetCategories 获取分类列表
func (s *ProductService) GetCategories(ctx context.Context) ([]string, error) {
	return s.productRepo.GetCategories(ctx)
}

// GetBrands 获取品牌列表
func (s *ProductService) GetBrands(ctx context.Context) ([]string, error) {
	return s.productRepo.GetBrands(ctx)
}
