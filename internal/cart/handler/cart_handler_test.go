package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecommerce-platform/internal/cart/handler"
	"github.com/ecommerce-platform/internal/cart/mocks"
	"github.com/ecommerce-platform/internal/cart/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() { gin.SetMode(gin.TestMode) }

// newRouter 构建测试路由，userID>0 时模拟已登录
func newRouter(svc *mocks.MockCartService, userID uint) *gin.Engine {
	r := gin.New()
	h := handler.NewCartHandler(svc)

	auth := func(c *gin.Context) {
		if userID > 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	}

	r.GET("/api/v1/cart", auth, h.GetUserCart)
	r.POST("/api/v1/cart/items", auth, h.AddCartItem)
	r.PUT("/api/v1/cart/items/:id/quantity", auth, h.UpdateCartItemQuantity)
	r.PUT("/api/v1/cart/items/:id/selected", auth, h.UpdateCartItemSelected)
	r.DELETE("/api/v1/cart/items/:id", auth, h.DeleteCartItem)
	r.DELETE("/api/v1/cart", auth, h.ClearCart)
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

// ---- GetUserCart ----

func TestHandler_GetUserCart_Success(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("GetUserCart", mock.Anything, uint(1)).Return(&service.CartVO{
		UserID:         1,
		ItemCount:      2,
		TotalPrice:     10000,
		SelectedAmount: 8000,
		Items:          []*service.CartItemVO{{ID: 1}, {ID: 2}},
	}, nil)

	w := getReq(newRouter(svc, 1), "/api/v1/cart")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
	assert.Contains(t, w.Body.String(), `"total_price":10000`)
	svc.AssertExpectations(t)
}

func TestHandler_GetUserCart_ServiceError(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("GetUserCart", mock.Anything, uint(1)).Return(nil, errors.New("db error"))

	w := getReq(newRouter(svc, 1), "/api/v1/cart")

	assert.Contains(t, w.Body.String(), `"code":10004`) // ErrNotFound
}

// ---- AddCartItem ----

func TestHandler_AddCartItem_Success(t *testing.T) {
	svc := new(mocks.MockCartService)
	// UserID 由 middleware 注入，不由客户端传入
	svc.On("AddCartItem", mock.Anything, mock.MatchedBy(func(req *service.AddCartItemReq) bool {
		return req.UserID == 1 && req.ProductID == 101 && req.Quantity == 2
	})).Return(nil)

	w := postJSON(newRouter(svc, 1), "/api/v1/cart/items", map[string]any{
		"product_id":    101,
		"merchant_id":   10,
		"quantity":      2,
		"product_name":  "手机",
		"product_price": 5000,
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

func TestHandler_AddCartItem_InvalidParam(t *testing.T) {
	svc := new(mocks.MockCartService)

	// 发送非 JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc, 1).ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), `"code":10001`) // ErrParam
	svc.AssertNotCalled(t, "AddCartItem")
}

func TestHandler_AddCartItem_ServiceError(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("AddCartItem", mock.Anything, mock.Anything).Return(errors.New(pkgerrors.Msg(pkgerrors.ErrParam)))

	w := postJSON(newRouter(svc, 1), "/api/v1/cart/items", map[string]any{
		"product_id": 101, "quantity": 0,
	})

	assert.Contains(t, w.Body.String(), `"code":10000`) // ErrServer
}

// ---- UpdateCartItemQuantity ----

func TestHandler_UpdateCartItemQuantity_Success(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("UpdateCartItemQuantity", mock.Anything, mock.MatchedBy(func(req *service.UpdateCartItemQuantityReq) bool {
		return req.UserID == 1 && req.ItemID == 10 && req.Quantity == 5
	})).Return(nil)

	w := putJSON(newRouter(svc, 1), "/api/v1/cart/items/10/quantity", map[string]any{
		"quantity": 5,
	})

	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateCartItemQuantity_Forbidden(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("UpdateCartItemQuantity", mock.Anything, mock.Anything).
		Return(errors.New(pkgerrors.Msg(pkgerrors.ErrForbidden)))

	w := putJSON(newRouter(svc, 1), "/api/v1/cart/items/99/quantity", map[string]any{
		"quantity": 1,
	})

	assert.Contains(t, w.Body.String(), `"code":10000`)
}

// ---- UpdateCartItemSelected ----

func TestHandler_UpdateCartItemSelected_Success(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("UpdateCartItemSelected", mock.Anything, mock.MatchedBy(func(req *service.UpdateCartItemSelectedReq) bool {
		return req.UserID == 1 && req.ItemID == 10 && req.Selected == true
	})).Return(nil)

	w := putJSON(newRouter(svc, 1), "/api/v1/cart/items/10/selected", map[string]any{
		"selected": true,
	})

	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

// ---- DeleteCartItem ----

func TestHandler_DeleteCartItem_ByItemID(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("DeleteCartItem", mock.Anything, mock.MatchedBy(func(req *service.DeleteCartItemReq) bool {
		return req.UserID == 1 && req.ItemID == 10 && req.ProductID == 0
	})).Return(nil)

	w := deleteReq(newRouter(svc, 1), "/api/v1/cart/items/10")

	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

func TestHandler_DeleteCartItem_ByProductID(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("DeleteCartItem", mock.Anything, mock.MatchedBy(func(req *service.DeleteCartItemReq) bool {
		return req.UserID == 1 && req.ItemID == 0 && req.ProductID == 101
	})).Return(nil)

	// itemID=0（路径传0），productID 通过 query string 传入
	w := deleteReq(newRouter(svc, 1), "/api/v1/cart/items/0?product_id=101")

	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

// ---- ClearCart ----

func TestHandler_ClearCart_Success(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("ClearCart", mock.Anything, uint(1)).Return(nil)

	w := deleteReq(newRouter(svc, 1), "/api/v1/cart")

	assert.Contains(t, w.Body.String(), `"code":0`)
	svc.AssertExpectations(t)
}

func TestHandler_ClearCart_ServiceError(t *testing.T) {
	svc := new(mocks.MockCartService)
	svc.On("ClearCart", mock.Anything, uint(1)).Return(errors.New("db error"))

	w := deleteReq(newRouter(svc, 1), "/api/v1/cart")

	assert.Contains(t, w.Body.String(), `"code":10000`)
}
