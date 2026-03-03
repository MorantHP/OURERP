package handlers

import (
	"strconv"

	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/response"
	"github.com/MorantHP/OURERP/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ProductHandlerV2 商品处理器（重构版）
type ProductHandlerV2 struct {
	productService *services.ProductService
}

// NewProductHandlerV2 创建商品处理器
func NewProductHandlerV2(productService *services.ProductService) *ProductHandlerV2 {
	return &ProductHandlerV2{
		productService: productService,
	}
}

// List 商品列表
// GET /api/v1/products
func (h *ProductHandlerV2) List(c *gin.Context) {
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
	products, total, err := h.productService.ListProducts(ctx, page, size, category, brand, keyword, status)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPage(c, products, page, size, total)
}

// Get 商品详情
// GET /api/v1/products/:id
func (h *ProductHandlerV2) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	product, err := h.productService.GetProduct(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, product)
}

// Create 创建商品
// POST /api/v1/products
func (h *ProductHandlerV2) Create(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	product, err := h.productService.CreateProduct(ctx, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, product)
}

// Update 更新商品
// PUT /api/v1/products/:id
func (h *ProductHandlerV2) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	product, err := h.productService.UpdateProduct(ctx, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "更新成功", product)
}

// Delete 删除商品
// DELETE /api/v1/products/:id
func (h *ProductHandlerV2) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	ctx := c.Request.Context()
	if err := h.productService.DeleteProduct(ctx, id); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetCategories 获取分类列表
// GET /api/v1/products/categories
func (h *ProductHandlerV2) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.productService.GetCategories(ctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"categories": categories})
}

// GetBrands 获取品牌列表
// GET /api/v1/products/brands
func (h *ProductHandlerV2) GetBrands(c *gin.Context) {
	ctx := c.Request.Context()
	brands, err := h.productService.GetBrands(ctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"brands": brands})
}
