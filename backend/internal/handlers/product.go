package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// ProductHandler 商品处理器
type ProductHandler struct {
	productRepo   *repository.ProductRepository
	inventoryRepo *repository.InventoryRepository
}

// NewProductHandler 创建商品处理器
func NewProductHandler(productRepo *repository.ProductRepository, inventoryRepo *repository.InventoryRepository) *ProductHandler {
	return &ProductHandler{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

// List 商品列表
// GET /api/v1/products
func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")
	brand := c.Query("brand")
	keyword := c.Query("keyword")

	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	ctx := c.Request.Context()
	products, total, err := h.productRepo.ListWithContext(ctx, page, size, category, brand, keyword, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": products,
		"pagination": gin.H{
			"page":  page,
			"size": size,
			"total": total,
		},
	})
}

// Get 商品详情
// GET /api/v1/products/:id
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ctx := c.Request.Context()
	product, err := h.productRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// 获取总库存
	totalQty, _ := h.inventoryRepo.GetTotalQuantity(ctx, id)

	c.JSON(http.StatusOK, gin.H{
		"product":       product,
		"total_quantity": totalQty,
	})
}

// Create 创建商品
// POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取租户ID
	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择账套"})
		return
	}

	// 检查SKU编码是否已存在
	ctx := c.Request.Context()
	if _, err := h.productRepo.FindBySkuCode(ctx, req.SkuCode); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品编码已存在"})
		return
	}

	product := &models.Product{
		TenantID:  tenantID,
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

	if err := h.productRepo.Create(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"product": product,
	})
}

// Update 更新商品
// PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	product, err := h.productRepo.FindByIDWithContext(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
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

	if err := h.productRepo.Update(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"product": product,
	})
}

// Delete 删除商品
// DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.productRepo.DeleteWithContext(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetCategories 获取分类列表
// GET /api/v1/products/categories
func (h *ProductHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.productRepo.GetCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetBrands 获取品牌列表
// GET /api/v1/products/brands
func (h *ProductHandler) GetBrands(c *gin.Context) {
	ctx := c.Request.Context()
	brands, err := h.productRepo.GetBrands(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"brands": brands})
}
