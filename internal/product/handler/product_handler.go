package handler

import (
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

// 实现 CreateProduct handler
// POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// 解析注册请求参数
	var req service.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	if err := h.productService.CreateProduct(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

// 实现 GetProduct handler
// GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	product, err := h.productService.GetProduct(c.Request.Context(), uint(productID))
	if err != nil {
		response.Fail(c, pkgerrors.ErrProductNotFound, err.Error())
		return
	}
	
	response.Success(c, product)
}

// 实现 ListProducts handler
// GET /api/v1/products?page=1&page_size=20&category_id=1
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req service.ListProductReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	resp, err := h.productService.ListProducts(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}
	
	response.Success(c, resp)
}

// 实现 UpdateProduct handler
// PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID, ok := parseID(c)
	if !ok {
		return
	}

	var req service.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	if err := h.productService.UpdateProduct(c.Request.Context(), productID, &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
    	return
	}
	response.Success(c, nil)
	
}

// 实现 DeleteProduct handler
// DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// 提示：解析 id，调用 h.productService.DeleteProduct
	productID, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.productService.DeleteProduct(c.Request.Context(), productID); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
    	return
	}

	response.Success(c, nil)
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
