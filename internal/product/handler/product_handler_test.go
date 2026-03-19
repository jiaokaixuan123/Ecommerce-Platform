package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecommerce-platform/internal/product/domain"
	"github.com/ecommerce-platform/internal/product/handler"
	"github.com/ecommerce-platform/internal/product/mocks"
	"github.com/ecommerce-platform/internal/product/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter 构建测试用路由，注入 mock service
// merchantID > 0 时模拟已登录的商家用户（写入 gin context）
func newRouter(svc *mocks.MockProductService, merchantID uint) *gin.Engine {
	r := gin.New()
	h := handler.NewProductHandler(svc)

	// 模拟 JWT 中间件：把 merchantID 写入 context
	authMiddleware := func(c *gin.Context) {
		if merchantID > 0 {
			c.Set("user_id", merchantID)
		}
		c.Next()
	}

	r.GET("/api/v1/products", h.ListProducts)
	r.GET("/api/v1/products/:id", h.GetProduct)
	r.POST("/api/v1/products", authMiddleware, h.CreateProduct)
	r.PUT("/api/v1/products/:id", authMiddleware, h.UpdateProduct)
	r.DELETE("/api/v1/products/:id", authMiddleware, h.DeleteProduct)
	return r
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func putJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func deleteReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- CreateProduct ----

// TestHandler_CreateProduct_Success 测试正常创建商品
func TestHandler_CreateProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("CreateProduct", mock.Anything, mock.Anything).Return(nil)

	// TODO:
	// 1. 调用 postJSON，传入合法的商品数据（name, price, stock, category_id）
	// 2. 断言 HTTP 状态码 200
	// 3. 断言响应体包含 `"code":0`

	_ = assert.New(t) // 占位，实现时删除
}

// TestHandler_CreateProduct_InvalidParam 测试缺少必填字段时返回参数错误
func TestHandler_CreateProduct_InvalidParam(t *testing.T) {
	svc := new(mocks.MockProductService)

	// TODO:
	// 1. 调用 postJSON，传入缺少 name 或 price 的请求体
	// 2. 断言响应体包含 `"code":10001`
	// 3. svc.AssertNotCalled(t, "CreateProduct")
}

// TestHandler_CreateProduct_ServiceError 测试 service 返回错误时响应正确
func TestHandler_CreateProduct_ServiceError(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("CreateProduct", mock.Anything, mock.Anything).Return(errors.New("db error"))

	// TODO:
	// 1. 调用 postJSON，传入合法请求体
	// 2. 断言响应体包含 `"code":10000`
}

// ---- GetProduct ----

// TestHandler_GetProduct_Success 测试正常获取商品详情
func TestHandler_GetProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	product := &domain.Product{ID: 1, Name: "手机", Price: 299900}
	svc.On("GetProduct", mock.Anything, uint(1)).Return(product, nil)

	// TODO:
	// 1. 调用 getReq(r, "/api/v1/products/1")
	// 2. 断言状态码 200
	// 3. 断言响应体包含 "手机"
}

// TestHandler_GetProduct_NotFound 测试商品不存在时返回正确错误码
func TestHandler_GetProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("GetProduct", mock.Anything, uint(99)).Return(nil, errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	// TODO:
	// 1. 调用 getReq(r, "/api/v1/products/99")
	// 2. 断言响应体包含 `"code":12001`
}

// ---- ListProducts ----

// TestHandler_ListProducts_Success 测试正常获取商品列表
func TestHandler_ListProducts_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	resp := &service.ListProductResp{
		Total:    2,
		Products: []*domain.Product{{ID: 1}, {ID: 2}},
	}
	svc.On("ListProducts", mock.Anything, mock.Anything).Return(resp, nil)

	// TODO:
	// 1. 调用 getReq(r, "/api/v1/products")
	// 2. 断言状态码 200
	// 3. 断言响应体包含 `"total":2`
}

// ---- UpdateProduct ----

// TestHandler_UpdateProduct_Success 测试正常更新商品
func TestHandler_UpdateProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("UpdateProduct", mock.Anything, uint(1), mock.Anything).Return(nil)

	// TODO:
	// 1. 调用 putJSON(r, "/api/v1/products/1", map 包含 name 或 price)
	// 2. 断言响应体包含 `"code":0`
}

// TestHandler_UpdateProduct_NotFound 测试更新不存在的商品
func TestHandler_UpdateProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("UpdateProduct", mock.Anything, uint(99), mock.Anything).
		Return(errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	// TODO:
	// 1. 调用 putJSON(r, "/api/v1/products/99", ...)
	// 2. 断言响应体包含 `"code":10000`
}

// ---- DeleteProduct ----

// TestHandler_DeleteProduct_Success 测试正常删除商品
func TestHandler_DeleteProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("DeleteProduct", mock.Anything, uint(1)).Return(nil)

	// TODO:
	// 1. 调用 deleteReq(r, "/api/v1/products/1")
	// 2. 断言响应体包含 `"code":0`
}

// TestHandler_DeleteProduct_NotFound 测试删除不存在的商品
func TestHandler_DeleteProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("DeleteProduct", mock.Anything, uint(99)).
		Return(errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	// TODO:
	// 1. 调用 deleteReq(r, "/api/v1/products/99")
	// 2. 断言响应体包含 `"code":10000`
}
