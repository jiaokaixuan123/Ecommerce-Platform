package handler

import (
	"net/http"
	"strconv"

	"github.com/ecommerce-platform/internal/product/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{productService: svc}
}

// TODO Step 1: 实现 CreateProduct handler
// POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// 提示：
	// 1. ShouldBindJSON 解析请求体到 service.CreateProductReq
	// 2. 调用 h.productService.CreateProduct
	// 3. 成功返回 response.Success，失败返回 response.Fail
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// TODO Step 2: 实现 GetProduct handler
// GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// 提示：
	// 1. 用 strconv.ParseUint(c.Param("id"), 10, 32) 解析路径参数
	// 2. 调用 h.productService.GetProduct
	// 3. 注意处理商品不存在的情况（返回 ErrProductNotFound）
	_ = strconv.ParseUint // 消除未使用警告，实现后删除这行
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// TODO Step 3: 实现 ListProducts handler
// GET /api/v1/products?page=1&page_size=20&category_id=1
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// 提示：
	// 1. 用 ShouldBindQuery 解析查询参数到 service.ListProductReq
	// 2. 调用 h.productService.ListProducts
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// TODO Step 4: 实现 UpdateProduct handler
// PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// 提示：
	// 1. 解析路径参数 id
	// 2. ShouldBindJSON 解析请求体到 service.UpdateProductReq
	// 3. 调用 h.productService.UpdateProduct
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// TODO Step 5: 实现 DeleteProduct handler
// DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// 提示：解析 id，调用 h.productService.DeleteProduct
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// 辅助函数：解析路径参数 id，失败时直接返回错误响应
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, "invalid id")
		return 0, false
	}
	return uint(id), true
}
