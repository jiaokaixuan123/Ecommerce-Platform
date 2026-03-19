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

// postJSON：发送 POST JSON 请求
func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// putJSON：发送 PUT JSON 请求
func putJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// getReq 发送GET请求
func getReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// deleteReq 发送DELETE请求
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

	w := postJSON(newRouter(svc, 1), "/api/v1/products", map[string]any{
		"name": "手机", "price": 900, "stock": 10, "category_id": 1,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestHandler_CreateProduct_InvalidParam 测试缺少必填字段时返回参数错误
func TestHandler_CreateProduct_InvalidParam(t *testing.T) {
	svc := new(mocks.MockProductService)

	w := postJSON(newRouter(svc, 1), "/api/v1/products", map[string]any{
		"name": "电脑", "stock": 10,
	})
	assert.Contains(t, w.Body.String(), `"code":10001`)
	svc.AssertNotCalled(t, "CreateProduct")
}

// TestHandler_CreateProduct_ServiceError 测试 service 返回错误时响应正确
func TestHandler_CreateProduct_ServiceError(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("CreateProduct", mock.Anything, mock.Anything).Return(errors.New("db error"))

	w := postJSON(newRouter(svc, 1), "/api/v1/products", map[string]any{
		"name": "手机", "price": 900, "stock": 10, "category_id": 1,
	})

	assert.Contains(t, w.Body.String(), `"code":1000`)
}

// ---- GetProduct ----

// TestHandler_GetProduct_Success 测试正常获取商品详情
func TestHandler_GetProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	product := &domain.Product{ID: 1, Name: "手机", Price: 299900}
	svc.On("GetProduct", mock.Anything, uint(1)).Return(product, nil)

	w := getReq(newRouter(svc, 0), "/api/v1/products/1")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "手机")
}

// TestHandler_GetProduct_NotFound 测试商品不存在时返回正确错误码
func TestHandler_GetProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("GetProduct", mock.Anything, uint(99)).Return(nil, errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	w := getReq(newRouter(svc, 0), "/api/v1/products/99")
	assert.Contains(t, w.Body.String(), `"code":12001`)
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

	w := getReq(newRouter(svc, 0), "/api/v1/products")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":2`)
}

// ---- UpdateProduct ----

// TestHandler_UpdateProduct_Success 测试正常更新商品
func TestHandler_UpdateProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("UpdateProduct", mock.Anything, uint(1), mock.Anything).Return(nil)

	w := putJSON(newRouter(svc, 1), "/api/v1/products/1", map[string]any{
		"name": "电视", "price": 2500,
	})
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestHandler_UpdateProduct_NotFound 测试更新不存在的商品
func TestHandler_UpdateProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("UpdateProduct", mock.Anything, uint(99), mock.Anything).
		Return(errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	w := putJSON(newRouter(svc, 1), "/api/v1/products/99", map[string]any{
		"name": "电视", "price": 2500,
	})
	assert.Contains(t, w.Body.String(), `"code":10000`)
}

// ---- DeleteProduct ----

// TestHandler_DeleteProduct_Success 测试正常删除商品
func TestHandler_DeleteProduct_Success(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("DeleteProduct", mock.Anything, uint(1)).Return(nil)

	w := deleteReq(newRouter(svc, 1), "/api/v1/products/1")
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestHandler_DeleteProduct_NotFound 测试删除不存在的商品
func TestHandler_DeleteProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockProductService)
	svc.On("DeleteProduct", mock.Anything, uint(99)).
		Return(errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound)))

	w := deleteReq(newRouter(svc, 1), "/api/v1/products/99")
	assert.Contains(t, w.Body.String(), `"code":10000`)
}
